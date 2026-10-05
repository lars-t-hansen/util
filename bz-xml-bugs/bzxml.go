// Bzxml reads and parses XML-exported bugzilla data, then "does something" with it, see options.
// In the absence of options, prints every bug ID and title.
//
// Usage:
//
//	bzxml [options] filename.xml
//
// Options:
//
//	-sel n
//	  Print bug n
//
// Bzxml handles only the Bugzilla XML structure as created by our quite old bugzilla install:
//
//	<bugzilla>
//	  <bug>...</bug>
//	  ...
//	</bugzilla>
//
// See code for the exact structure supported.
package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type Timestamp time.Time

func (t *Timestamp) UnmarshalText(text []byte) error {
	x, err := time.Parse("2006-01-02 15:04:05 -0700", string(text))
	if err != nil {
		return err
	}
	*t = Timestamp(x)
	return nil
}

func (t *Timestamp) Format(fmt string) string {
	return time.Time(*t).Format(fmt)
}

type Email struct {
	Name string `xml:"name,attr"`
	Addr string `xml:",chardata"`
}

type Bugzilla struct {
	Bugs []BzBug `xml:"bug"`
}

type BzBug struct {
	BugId       uint      `xml:"bug_id"`
	Creation    Timestamp `xml:"creation_ts"`
	ShortDesc   string    `xml:"short_desc"`
	Product     string    `xml:"product"`
	Component   string    `xml:"component"`
	RepPlatform string    `xml:"rep_platform"`
	BugStatus   string    `xml:"bug_status"`
	Resolution  string    `xml:"resolution"`
	Priority    string    `xml:"priority"`
	// there are many more fields
	Comments []BzLongDesc `xml:"long_desc"`
}

type BzLongDesc struct {
	CommentId    uint      `xml:"commentid"`
	CommentCount uint      `xml:"comment_count"`
	Who          Email     `xml:"who"`
	When         Timestamp `xml:"bug_when"`
	TheText      string    `xml:"thetext"`
	// there are more fields
}

var (
	selFlag = flag.Uint("sel", 0, "Select a specific bug")
)

func main() {
	rest := FlagParse("bzxml", "filename")
	infilename := rest[0]
	f, err := os.Open(infilename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	blob, err := io.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	var bz Bugzilla
	if err := xml.Unmarshal(blob, &bz); err != nil {
		log.Fatal(err)
	}

	if *selFlag > 0 {
		for _, b := range bz.Bugs {
			if b.BugId == *selFlag {
				fmt.Printf("%d  %s\n", b.BugId, b.ShortDesc)
				fmt.Printf("  %s / %s @ %s\n", b.Product, b.Component, b.Creation.Format(time.RFC3339))
				for _, c := range b.Comments {
					fmt.Printf("-------\n%s <%s> @ %s\n%s\n", c.Who.Name, c.Who.Addr, c.When.Format(time.RFC3339), c.TheText)
				}
				fmt.Printf("-------\n")
				break
			}
		}
	} else {
		// The default is to just print the ID and title
		for _, b := range bz.Bugs {
			fmt.Printf("%d  %s\n", b.BugId, b.ShortDesc)
		}
	}
}
