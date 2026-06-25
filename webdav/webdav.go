package main

import (
	"flag"
	"log"
	"net/http"

	"golang.org/x/net/webdav"
)

var (
	directory = flag.String("d", ".", "directory to share")
	iface     = flag.String("i", "localhost:8080", "interface to serve at")
	verbose   = flag.Bool("v", false, "verbose logging")
)

func main() {
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
