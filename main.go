package main

import (
	"log"
	"net/http"
)

func HealzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("OK"))
}

func main() {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/app/*", http.StripPrefix("/app", fileServer))
	mux.HandleFunc("/healthz", HealzHandler)
	server := http.Server{Handler: mux, Addr: "localhost:8080"}
	log.Fatal(server.ListenAndServe())
}
