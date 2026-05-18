// HttpRepo manages a directory tree remotely over HTTP / HTTPS.
//
// Usage:
//   httprepo [options] directory
//
// Options:
//   -cert  Server cert file, if using (requires key)
//   -key   Server key file, if using (requires cert)
//   -p     Port, default 8080
//   -v     Verbose
//
// GET /name will serve the file of that name or 404 if not present.
//
// HEAD /name will serve the metadata, ditto
//
// PUT /name will replace the file or create a new one with the input given, and may create new
// subdirectories.
//
// POST /name will append to the file, creating it if necessary, and may create new subdirectories.
//
// DELETE /name will delete the file.
package main

// TODO: for properly configured hosts there's already an official cert and we could be doing
//       https without the cert/key, and there should be an option for that?
// TODO: better metadata for GET/HEAD, notably mime type, mod date, and size
// TODO: Could implement GET on .../ as a command to list the contents of that directory

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

var (
	cert    = flag.String("cert", "", "Name of server cert `filename`, if https, requires key")
	key     = flag.String("key", "", "Name of private key `filename`, if https, requires cert")
	port    = flag.Uint("p", 8080, "Server `port`")
	verbose = flag.Bool("v", false, "Verbose logging")
	dirName string
)

func main() {
	o := flag.CommandLine.Output()
	flag.Usage = func() {
		cmd := os.Args[0]
		fmt.Fprintf(o, "Serve files in a directory in response to GET and replace them in response to PUT.\n\n")
		fmt.Fprintf(o, "Usage of %s:\n", cmd)
		fmt.Fprintf(o, "  %s [options] directory\n\n", cmd)
		fmt.Fprintf(o, "Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if len(flag.Args()) != 1 {
		fmt.Fprintf(flag.CommandLine.Output(), "Error: Exactly one directory name required\n\n")
		flag.Usage()
		os.Exit(2)
	}
	if (*cert == "") != (*key == "") {
		fmt.Fprintf(o, "Error: Both cert and key required if either present\n\n")
		flag.Usage()
		os.Exit(2)
	}
	dirName = path.Clean(flag.Args()[0])

	if *verbose {
		log.Printf("Listening on port %d for directory %s", *port, dirName)
	}
	dir := os.DirFS(dirName)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Some headers are sent automatically: Date, Content-Type, Transfer-Encoding, unless set
		// to nil in the header map.
		w.Header()["Server"] = []string{"Comanche"}
		filename := path.Clean(r.URL.Path)[1:]
		if strings.HasPrefix(filename, "/") || strings.HasPrefix(filename, "..") {
			if *verbose {
				log.Printf("Bad file %s", filename)
			}
			w.WriteHeader(422)
			return
		}
		if *verbose {
			log.Println("HEADERS")
			for k, v := range r.Header {
				log.Printf(" %s", k)
				for _, val := range v {
					log.Printf("  %s", val)
				}
			}
		}
		// At this point, path.Join(dirName, filename) should give us a name below the dir always,
		// necessary for operations not available through the `dir` object.
		switch r.Method {
		case "HEAD":
			if *verbose {
				log.Printf("HEAD %s", filename)
			}
			ifo, err := dir.(fs.StatFS).Stat(filename)
			if err != nil {
				if *verbose {
					log.Printf("File not found: %s", filename)
				}
				w.WriteHeader(404)
				return
			}
			w.Header()["Last-Modified"] = []string{ifo.ModTime().Format("Mon, 02 Jan 2006 15:04:05 GMT")}
			w.WriteHeader(200)

		case "GET":
			if *verbose {
				log.Printf("GET %s", filename)
			}
			ifo, err := dir.(fs.StatFS).Stat(filename)
			if err != nil {
				if *verbose {
					log.Printf("File not found: %s", filename)
				}
				w.WriteHeader(404)
				return
			}
			// Reading everything before writing it is OK for all but the largest files.
			contents, err := dir.(fs.ReadFileFS).ReadFile(filename)
			if err != nil {
				if *verbose {
					log.Printf("File not found: %s", filename)
				}
				w.WriteHeader(404)
				return
			}
			w.Header()["Last-Modified"] = []string{ifo.ModTime().Format("Mon, 02 Jan 2006 15:04:05 GMT")}
			if rng := r.Header["Range"]; rng != nil {
				if len(rng) == 1 {
					var from, to int
					n, err := fmt.Sscanf(rng[0], "bytes=%d-%d", &from, &to)
					if err == nil && n == 2 && from <= to && to < len(contents) {
						w.WriteHeader(206)
						w.Write(contents[from : to+1])
						return
					}
				}
			}

			w.WriteHeader(200)
			// Ignore errors
			w.Write(contents)

		case "DELETE":
			if *verbose {
				log.Printf("DELETE %s", filename)
			}
			fullname := path.Join(dirName, filename)
			if err := os.Remove(fullname); err != nil {
				// Must return a success code for a file that was not there.  This is not the ideal
				// way to do it, but Go's error reporting here is weak.
				if _, err := dir.(fs.StatFS).Stat(filename); err != nil {
					w.WriteHeader(204)
				} else {
					w.WriteHeader(404)
				}
			} else {
				w.WriteHeader(204)
			}

		case "PUT", "POST":
			isPost := r.Method == "POST"
			fullname := path.Join(dirName, filename)
			subdirname := path.Dir(fullname)
			err := os.MkdirAll(subdirname, 0o777)
			if err != nil {
				if *verbose {
					log.Printf("Could not mkdir")
				}
				w.WriteHeader(422)
				return
			}
			if *verbose {
				log.Printf("%s %s", r.Method, fullname)
			}
			bytes, err := io.ReadAll(r.Body)
			if err != nil {
				if *verbose {
					log.Printf("Failed to read input")
				}
				w.WriteHeader(422)
				return
			}
			if isPost {
				err = AppendFile(fullname, bytes, 0o664)
			} else {
				err = os.WriteFile(fullname, bytes, 0o664)
			}
			if err != nil {
				if *verbose {
					log.Printf("Failed to write output")
				}
				w.WriteHeader(422)
				return
			}
			w.WriteHeader(204)

		default:
			if *verbose {
				log.Printf("Bad method %s", r.Method)
			}
			w.WriteHeader(405)
		}
	})

	var e any
	if *cert != "" {
		e = http.ListenAndServeTLS(fmt.Sprintf(":%d", *port), *cert, *key, nil)
	} else {
		e = http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	}
	log.Fatal(e)
}

func AppendFile(filename string, bs []byte, mode os.FileMode) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(bs)
	return err
}
