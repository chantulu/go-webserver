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
	cfg := apiConfig{fileserverHits: 0}
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/app/*", cfg.middlewareMetricsInc(http.StripPrefix("/app", fileServer)))
	mux.HandleFunc("/healthz", HealzHandler)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		MetricsHandler(w, r, &cfg)
	})
	mux.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		ResetHandler(w, r, &cfg)
	})
	server := http.Server{Handler: mux, Addr: "localhost:8080"}
	log.Fatal(server.ListenAndServe())
}
