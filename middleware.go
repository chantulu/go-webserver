package main

import (
	"fmt"
	"net/http"
)

type apiConfig struct {
	fileserverHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r) // Call the next handler
	})
}

func MetricsHandler(w http.ResponseWriter, r *http.Request, cfg *apiConfig) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(fmt.Sprintf("Hits: %v", cfg.fileserverHits)))
}

func ResetHandler(w http.ResponseWriter, r *http.Request, cfg *apiConfig) {
	cfg.fileserverHits = 0
}
