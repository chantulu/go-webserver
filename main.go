package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"server/internal"
	"strconv"
	"strings"

	"github.com/lpernett/godotenv"
)



func HealzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("OK"))
}

func main() {
	mux := http.NewServeMux()
	cfg := apiConfig{fileserverHits: 0}
	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	fileServer := http.FileServer(http.Dir("./static"))
	dbPath := "./db.json"
	db, err := internal.NewDB(dbPath)
	errEnv := godotenv.Load()
	if errEnv != nil {
		log.Fatal("ERROR: Cannot initialize env")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	pokaKey := os.Getenv("POLKA_KEY")
	cfg.jwtSecret = jwtSecret
	cfg.polkaKey = pokaKey
	if *dbg{
		db.ResetDB()	
	}
	if err != nil {
		log.Fatal("ERROR: cannot initialize database at "+ dbPath)
	}
	chirps, err := db.GetChirps()
	if err != nil {
		log.Fatalf("ERROR: cannot retrieve chirps: %v", err)
	}
	fmt.Println(chirps)
	mux.Handle("/app/*", cfg.middlewareMetricsInc(http.StripPrefix("/app", fileServer)))
	mux.HandleFunc("GET /api/healthz", HealzHandler)
	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		MetricsHandler(w, r, &cfg)
	})
	mux.HandleFunc("/api/reset", func(w http.ResponseWriter, r *http.Request) {
		ResetHandler(w, r, &cfg)
	})
	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)
	mux.HandleFunc("POST /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		CreateChirpHandler(w, r, db, &cfg)
	})
	mux.HandleFunc("GET /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		GetChirpsHandler(w, r, db)
	})
	mux.HandleFunc("GET /api/chirps/{id}", func(w http.ResponseWriter, r *http.Request) {
		type retError struct {
			Error string `json:"error"`
		}
		chirpID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil{
			errMsg := retError{Error: err.Error()}
			dat, _ := json.Marshal(errMsg)
			w.WriteHeader(400)
			w.Write(dat)
			return
		}
		GetChirpHandler(w, r, db, chirpID)
	})
	mux.HandleFunc("GET /api/users", func(w http.ResponseWriter, r *http.Request) {
		GetUsersHandler(w, r, db)
	})
	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		CreateUsersHandler(w, r, db)
	})
	mux.HandleFunc("PUT /api/users", func(w http.ResponseWriter, r *http.Request) {
		UpdateUserHandler(w, r, db, &cfg)
	})
	mux.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		ValidateUserHandler(w, r, db, &cfg)
	})
	mux.HandleFunc("POST /api/refresh", func(w http.ResponseWriter, r *http.Request) {
		RefreshTokenHandler(w, r, db, &cfg)
	})
	mux.HandleFunc("POST /api/revoke", func(w http.ResponseWriter, r *http.Request) {
		RevokeTokenHandler(w, r, db, &cfg)
	})
	mux.HandleFunc("DELETE /api/chirps/{id}", func(w http.ResponseWriter, r *http.Request) {
		type retError struct {
			Error string `json:"error"`
		}
		chirpID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil{
			errMsg := retError{Error: err.Error()}
			dat, _ := json.Marshal(errMsg)
			w.WriteHeader(400)
			w.Write(dat)
			return
		}
		DeleteChirpHandler(w, r, db, &cfg, chirpID)
	})
	mux.HandleFunc("POST /api/polka/webhooks", func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("Authorization")
		apiKey = strings.Replace(apiKey,"ApiKey ","",1)
		if apiKey != cfg.polkaKey {
			w.WriteHeader(401)
			return
		}
		HandlePolkaWebhook(w, r, db, &cfg)
	})

	server := http.Server{Handler: mux, Addr: "localhost:8080"}
	log.Fatal(server.ListenAndServe())
}
