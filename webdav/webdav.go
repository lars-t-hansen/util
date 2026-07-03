// Webdav serves a directory over http using the WebDAV protocol.  Run with -h for options.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/webdav"
)

var (
	directory = flag.String("d", ".", "directory to share")
	iface     = flag.String("i", "localhost:8080", "interface to serve at")
	verbose   = flag.Bool("v", false, "verbose logging")
)

func main() {
	flag.Usage = func() {
		o := flag.CommandLine.Output()
		fmt.Fprintf(o, "Serve a directory via the WebDAV protocol.\n\n")
		fmt.Fprintf(o, "Usage of webdav:\n")
		fmt.Fprintf(o, "  %s [options] directory\n\n", os.Args[0])
		fmt.Fprintf(o, "Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	handler := webdav.Handler{
		FileSystem: webdav.Dir(*directory),
		LockSystem: webdav.NewMemLS(),
		Logger:     mylog,
	}
	http.ListenAndServe(*iface, &handler)
}

func mylog(r *http.Request, err error) {
	if *verbose {
		log.Printf("Request: %s %s  err: %v", r.Method, r.URL.Path, err)
	}
}
