// Manage a directory tree remotely over HTTP.
//
// GET /name will serve the file of that name or 404 if not present.
// HEAD /name will serve the metadata
//
// PUT /name will replace the file or create a new one with the input given, and may create new
// subdirectories.
//
// TODO: better metadata for GET/HEAD, notably mime type, mod date, and size
// TODO: Could implement GET on .../ as a command to list the contents of that directory
// TODO: Could implement DELETE

package main

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
	port = flag.Uint("p", 8080, "Server `port`")
	verbose = flag.Bool("v", false, "Verbose logging")
	dirName string
)

func main() {
	flag.Usage = func () {
		o := flag.CommandLine.Output()
		cmd := os.Args[0]
		fmt.Fprintf(o, "Serve files in a directory in response to GET and replace them in response to PUT.\n\n")
		fmt.Fprintf(o, "Usage of %s:\n", cmd)
		fmt.Fprintf(o, "  %s [options] directory\n\n", cmd)
		fmt.Fprintf(o, "Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(2)
	}
	dirName = path.Clean(flag.Args()[0])

	if *verbose {
		log.Printf("Listening on port %d for directory %s", *port, dirName)
	}
	dir := os.DirFS(dirName)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		filename := path.Clean(r.URL.Path)[1:]
		switch r.Method {
		case "HEAD":
			if *verbose {
				log.Printf("HEAD %s", filename)
			}
			if _, err := dir.(fs.StatFS).Stat(filename); err != nil {
				w.WriteHeader(404)
			} else {
				w.WriteHeader(200)
			}
		case "GET":
			if *verbose {
				log.Printf("GET %s", filename)
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
			w.WriteHeader(200)
			// TODO: Maybe want to write some headers?  Defaults are probably OK for now.
			// Ignore errors
			w.Write(contents)

		case "PUT":
			fullname := path.Join(dirName, filename)
			if strings.HasPrefix(fullname, "/") || strings.HasPrefix(fullname, "..") {
				if *verbose {
					log.Printf("Bad file %s %s", fullname, filename)
				}
				w.WriteHeader(422)
				return
			}
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
				log.Printf("PUT %s", fullname)
			}
			bytes, err := io.ReadAll(r.Body)
			if err != nil {
				if *verbose {
					log.Printf("Failed to read input")
				}
				w.WriteHeader(422)
				return
			}
			err = os.WriteFile(fullname, bytes, 0o664)
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

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
