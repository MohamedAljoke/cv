// Command serve previews the generated site locally: go run ./cmd/serve
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "listen address")
	root := flag.String("root", ".", "directory to serve")
	flag.Parse()
	fmt.Printf("serving %s on http://%s/\n", *root, *addr)
	log.Fatal(http.ListenAndServe(*addr, http.FileServer(http.Dir(*root))))
}
