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
	mux.HandleFunc("GET /api/healthz", HealzHandler)
	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		MetricsHandler(w, r, &cfg)
	})
	mux.HandleFunc("/api/reset", func(w http.ResponseWriter, r *http.Request) {
		ResetHandler(w, r, &cfg)
	})
	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)
	server := http.Server{Handler: mux, Addr: "localhost:8080"}
	log.Fatal(server.ListenAndServe())
}
