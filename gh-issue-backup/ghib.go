// Ghib performs periodic backups of github issues for a repo.
//
// Usage:
//
//	ghib [options] github-repo
//
// where github-repo is the name of the repo as it appears in github URLs, "repo-name" or
// "org-name/repo-name".
//
// Options:
//
//	-full
//	  The default is to attempt to get updates for open issues only, and to scan for new issues.
//	  With -full, additionally check closed issues for a changed time stamp and refresh if
//	  necessary.
//
//	-v
//	  Verbose logging
//
//	-dump
//	  Dump the database and exit, the repo-name must be present but is ignored
//
// ghib is not necessarily concurrency-safe.
//
// Storage:
//
// There is an sqlite database in the working directory called ghib.db that contains the state.  In
// addition there is a subdirectory called issues that has individual subdirectories for each
// numbered issue or pull request (eg issues/312 has data for issue 312).  Files in these
// directories are the JSON blobs received from Github (issues/312/issue, issues/312/comments, ...).
// When an issue is updated the existing file is overwritten with new data.
package main

// This appears to work OK but we get cut off because of the rate limiting.
//
// TODO:
// - probably some more verbose logging?
// - rate limit is 60 requests per hour for unauthenticated users
// - rate limit is 5000/hour for authenticated users (personal access token)
// - so we need an option to supply a token, and then we still need to
//   pay attention to the limit, somehow

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	fullFlag = flag.Bool("full", false, "Check closed issues for updates")
	verbose  = flag.Bool("v", false, "Verbose logging")
	dump     = flag.Bool("dump", false, "Dump the database and exit")
	repo     string
)

var db *sql.DB

func main() {
	var err error

	rest := FlagParse("ghib", []string{"repo"})
	repo = rest[0]

	db, err = sql.Open("sqlite3", "ghib.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	// We want every field that we may need for our own logic in the database so that we never have
	// to go and read and parse a file from the file store.  (Clearly we have to parse some of the
	// fetched data.)
	createDb := `
CREATE TABLE IF NOT EXISTS issues (
  id       INTEGER PRIMARY KEY,
  open     BOOLEAN,
  modified DATETIME
);
`
	if _, err := db.Exec(createDb); err != nil {
		log.Fatal(err)
	}

	if *dump {
		rows, err := db.Query("SELECT * FROM issues;")
		if err != nil {
			log.Fatal(err)
		}
		for rows.Next() {
			var id int
			var open bool
			var modified time.Time
			if err := rows.Scan(&id, &open, &modified); err != nil {
				log.Fatal(err)
			}
			fmt.Fprintf(os.Stderr, "%d %s %v\n", id, sel(open, "open", "closed"), modified)
		}
		return
	}

	err = os.Mkdir("issues", 0755)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	// First, check all open issues.  Note an issue may have been deleted, but in this case, just
	// mark it as closed, do not delete data.  For existing issues, update metadata.  If modtime
	// has changed, we need to update the issue: grab comments in addition, and save both files
	// in the issue dir.

	rows, err := db.Query("SELECT * FROM issues WHERE open = TRUE;")
	if err != nil {
		log.Fatal(err)
	}
	poll(rows)
	rows.Close()

	// Second, if -full, also check known closed issues in the same manner.

	if *fullFlag {
		rows, err := db.Query("SELECT * FROM issues WHERE open = FALSE;")
		if err != nil {
			log.Fatal(err)
		}
		poll(rows)
		rows.Close()
	}

	// Third, find the highest index in the DB and then start scanning for issues above that number,
	// whether open or closed (there could be none).  Record anything that's found with current
	// metadata.
	//
	// TODO: There's an interesting wrinkle, which is that if we know about 1..N, say, and N+1 and
	// N+2 are created, and then N+1 is deleted, how will we know about the existence of N+2?
	next := 1
	result, err := db.Query("SELECT MAX(id) FROM issues;")
	if rows.Next() {
		var id int
		if err := result.Scan(&id); err != nil {
			log.Fatal(err)
		}
		next = id + 1
	}
	if *verbose {
		fmt.Fprintf(os.Stderr, "next=%d\n", next)
	}
	for {
		iOpen, iModified, iText, err := getIssue(next)
		if err == notFound {
			// FIXME: Mark it as closed
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		err = downloadAndStore(next, iText)
		if err != nil {
			log.Fatal(err)
		}
		_, err = db.Exec("INSERT INTO issues VALUES($1, $2, $3);", next, iOpen, iModified)
		if err != nil {
			log.Fatal(err)
		}
		next++
	}
}

func poll(rows *sql.Rows) {
	for rows.Next() {
		var id int
		var open bool
		var modified time.Time
		if err := rows.Scan(&id, &open, &modified); err != nil {
			log.Fatal(err)
		}
		if *verbose {
			fmt.Fprintf(os.Stderr, "%d %s %v\n", id, sel(open, "open", "closed"), modified)
		}
		iOpen, iModified, iText, err := getIssue(id)
		if err == notFound {
			// FIXME: Mark it as closed
			continue
		}
		if err != nil {
			log.Fatal(err)
		}
		if iModified.After(modified) {
			err := downloadAndStore(id, iText)
			if err != nil {
				log.Fatal(err)
			}
			_, err = db.Exec("UPDATE issues SET open = $1, modified = $2 WHERE id = $3;", iOpen, iModified, id)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

var notFound = errors.New("Not found")

func getIssue(id int) (open bool, mtime time.Time, issueText string, err error) {
	if *verbose {
		fmt.Fprintf(os.Stderr, "Reading issue %d\n", id)
	}
	issueText, err = getIt("issues/" + fmt.Sprint(id))
	if err != nil {
		return
	}
	type Issue struct {
		State   string `json:"state"`
		Updated string `json:"updated_at"`
	}
	dec := json.NewDecoder(strings.NewReader(issueText))
	var m Issue
	err = dec.Decode(&m)
	if err != nil {
		return
	}
	switch m.State {
	case "open":
		open = true
	case "closed":
		open = false
	default:
		log.Printf("Unknown issue state %s", m.State)
	}
	mtime, err = time.Parse(time.RFC3339, m.Updated)
	if err != nil {
		log.Printf("Unknown time format %s", m.Updated)
	}
	return
}

func downloadAndStore(id int, issueText string) error {
	dirname := "issues/" + fmt.Sprint(id)
	err := os.Mkdir(dirname, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}
	ifile, err := os.Create(dirname + "/issue")
	if err != nil {
		return err
	}
	defer ifile.Close()
	_, err = io.WriteString(ifile, issueText)
	if err != nil {
		return err
	}
	// Really the comments URL comes from the issue itself
	commentText, err := getIt("issues/" + fmt.Sprint(id) + "/comments")
	if err != nil {
		return err
	}
	cfile, err := os.Create(dirname + "/comments")
	if err != nil {
		return err
	}
	defer cfile.Close()
	_, err = io.WriteString(cfile, commentText)
	if err != nil {
		return err
	}
	return nil
}

// FIXME: Implement notFound!
func getIt(path string) (bodyText string, err error) {
	// curl -H "Accept: application/vnd.github.v3+json" https://api.github.com/repos/NordicHPC/sonar/...
	url := "https://api.github.com/repos/" + repo + "/" + path
	if *verbose {
		fmt.Fprintf(os.Stderr, "  URL: %s\n", url)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header = map[string][]string{
		"Accept": {"application/vnd.github.v3+json"},
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode > 299 {
		err = fmt.Errorf("Response failed: %d", resp.StatusCode)
		return
	}
	if err != nil {
		return
	}
	bodyText = string(body)
	return
}

func sel[T any](flag bool, x, y T) T {
	if flag {
		return x
	}
	return y
}
