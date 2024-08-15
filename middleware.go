package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
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
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(fmt.Sprintf(`<html>
<body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
</body>
</html>`, cfg.fileserverHits)))
}

func ResetHandler(w http.ResponseWriter, r *http.Request, cfg *apiConfig) {
	cfg.fileserverHits = 0
}

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type retError struct {
		Error string `json:"error"`
	}
	type returnVal struct {
		Valid bool `json:"valid"`
		CleanedBody string `json:"cleaned_body"`
	}
	
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		errMsg := retError{Error: err.Error()}
		dat, _ := json.Marshal(errMsg)
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		w.Write(dat)
		return
    }

	if len(params.Body) > 140 {
		errMsg := retError{Error: "Chirp is too long"}
		dat, _ := json.Marshal(errMsg)
		w.WriteHeader(400)
		w.Write(dat)
		return
	}

	respBody := returnVal {
		Valid: true,
		CleanedBody: replaceProfanity(params.Body),
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		errMsg := retError{Error: err.Error()}
		log.Printf("Error decoding parameters: %s", err)
		dat, _ := json.Marshal(errMsg)
		w.WriteHeader(500)
		w.Write(dat)
		return
    }
	w.WriteHeader(200)
	w.Write(dat)
}

func replaceProfanity(s string) string{
	textSplit := strings.Split(s, " ")
	profanityWords := []string{"kerfuffle", "sharbert", "fornax"} 
	for index,word := range textSplit{
		if slices.Contains(profanityWords, strings.ToLower(word)) {
			textSplit[index] = "****"
		}
	}
	return strings.Join(textSplit, " ")
}