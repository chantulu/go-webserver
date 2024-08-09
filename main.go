package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/", fs)
	server := http.Server{Handler: mux, Addr: "localhost:8080"}
	log.Fatal(server.ListenAndServe())
}
