// SPDX-License-Identifier: MIT

// httpdump listens for incoming http traffic, prints it on stdout, and responds "Ok".
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
)

var (
	path = flag.String("P", "/", "Root path to handle")
	port = flag.Uint("p", 8090, "Port to listen on")
	all  = flag.Bool("a", false, "Print all (of the content)")
)

func main() {
	flag.Parse()
	http.HandleFunc(*path, func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("URL:    ", r.URL.String())
		fmt.Println("Method: ", r.Method)
		fmt.Println("Type:   ", r.Header["Content-Type"])
		if len(r.Header["Authorization"]) > 0 {
			fmt.Println("Auth:   ", r.Header["Authorization"])
		}
		fmt.Println("Length: ", r.ContentLength)
		payload := make([]byte, r.ContentLength)
		haveRead := 0
		for haveRead < int(r.ContentLength) {
			n, err := r.Body.Read(payload[haveRead:])
			haveRead += n
			if err != nil {
				if err == io.EOF && haveRead == int(r.ContentLength) {
					break
				}
				w.WriteHeader(400)
				return
			}
		}
		if !*all {
			payload = payload[0:min(len(payload), 100)]
		}
		fmt.Println(string(payload))
		w.WriteHeader(200)
	})
	http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
}
