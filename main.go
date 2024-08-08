package main

import (
	"net/http"
)



func main() {
	mux := http.NewServeMux()
	server := http.Server{Handler: mux, Addr: "localhost:8080"}
	server.ListenAndServe()
}