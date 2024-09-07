package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"server/internal"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

type apiConfig struct {
	fileserverHits int
	jwtSecret string
	polkaKey string
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

func CreateChirpHandler(w http.ResponseWriter, r *http.Request, db *internal.DB, cfg *apiConfig) {
	type parameters struct {
		Body string `json:"body"`
	}
	type retError struct {
		Error string `json:"error"`
	}
	
	tokenString := r.Header.Get("Authorization")
	tokenString = strings.Replace(tokenString,"Bearer ","",1)
	userID, ok := internal.IsAuthenticated(tokenString, cfg.jwtSecret)
	if !ok {
		errMsg := retError{Error: "Log in again"}
		dat, _ := json.Marshal(errMsg)
		log.Printf("Error authenticating")
		w.WriteHeader(401)
		w.Write(dat)
		return
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

	newChirp,err := db.CreateChirp(replaceProfanity(params.Body), userID)
	if err != nil {
		errMsg := retError{Error: err.Error()}
		log.Printf("Error decoding parameters: %s", err)
		dat, _ := json.Marshal(errMsg)
		w.WriteHeader(500)
		w.Write(dat)
		return
	}
	dat, err := json.Marshal(newChirp)
	if err != nil {
		errMsg := retError{Error: err.Error()}
		log.Printf("Error decoding parameters: %s", err)
		dat, _ := json.Marshal(errMsg)
		w.WriteHeader(500)
		w.Write(dat)
		return
    }
	w.WriteHeader(201)
	w.Write(dat)
}


func GetChirpsHandler(w http.ResponseWriter, r *http.Request, db *internal.DB) {
	type retError struct {
		Error string `json:"error"`
	}
	s := r.URL.Query().Get("author_id")
	chirps, err := db.GetChirps()
	so := r.URL.Query().Get("sort")
	if err != nil {
		errMsg := retError{Error: err.Error()}
		log.Printf("Error loading chirps: %v", err)
		dat, _ := json.Marshal(errMsg)
		w.WriteHeader(500)
		w.Write(dat)
		return
	}
	if sint, err := strconv.Atoi(s); err == nil && s != "" {
		var tempchirp []internal.Chirp;
		tempchirp = make([]internal.Chirp, 0)
		for _, c := range chirps {
			if c.AuthorID == sint {
				tempchirp = append(tempchirp, c)
			}
		}
		chirps = tempchirp
	}
	if so != "" && so == "desc"{
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].ID > chirps[j].ID
		})
	} else {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].ID < chirps[j].ID
		})
	}
	dat, _ := json.Marshal(chirps)
	w.WriteHeader(200)
	w.Write(dat)
}

func GetChirpHandler(w http.ResponseWriter, r *http.Request, db *internal.DB, chirpID int) {
	type retError struct {
		Error string `json:"error"`
	}
  chirp, ok := db.GetSingleChirp(chirpID)
  if !ok {
	errMsg := retError{Error: "Chirp not found"}
	dat, _ := json.Marshal(errMsg)
	w.WriteHeader(404)
	w.Write(dat)
	return
  }
  dat, err := json.Marshal(chirp)
  if err != nil {
	  w.WriteHeader(500)
	  return
	}
  w.WriteHeader(200)
  w.Write(dat)
}



func GetUsersHandler(w http.ResponseWriter, r *http.Request, db *internal.DB) {
	type retError struct {
		Error string `json:"error"`
	}
	users, err := db.GetUsers()
	if err != nil {
		errMsg := retError{Error: err.Error()}
		log.Printf("Error loading users: %v", err)
		dat, _ := json.Marshal(errMsg)
		w.WriteHeader(500)
		w.Write(dat)
		return
	}
	dat, _ := json.Marshal(users)
	w.WriteHeader(200)
	w.Write(dat)
}


func CreateUsersHandler(w http.ResponseWriter, r *http.Request, db *internal.DB) {
	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	type retError struct {
		Error string `json:"error"`
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

	if len(params.Email) > 140 {
		errMsg := retError{Error: "Email is too long"}
		dat, _ := json.Marshal(errMsg)
		w.WriteHeader(400)
		w.Write(dat)
		return
	}

	newUser,err := db.CreateUser(params.Email, params.Password)
	if err != nil {
		errMsg := retError{Error: err.Error()}
		log.Printf("Error decoding parameters: %s", err)
		dat, _ := json.Marshal(errMsg)
		w.WriteHeader(500)
		w.Write(dat)
		return
	}
	dat, err := json.Marshal(newUser)
	if err != nil {
		errMsg := retError{Error: err.Error()}
		log.Printf("Error decoding parameters: %s", err)
		dat, _ := json.Marshal(errMsg)
		w.WriteHeader(500)
		w.Write(dat)
		return
    }
	w.WriteHeader(201)
	w.Write(dat)
}

func ValidateUserHandler(w http.ResponseWriter, r *http.Request, db *internal.DB, cfg *apiConfig) {
	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
		Expires int `json:"expires_in_seconds"`
	}
	type authRes struct{
		Email string `json:"email"`
		ID int `json:"id"`
		Token string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		IsChirpyRed bool `json:"is_chirpy_red"`
	}
	type retError struct {
		Error string `json:"error"`
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

	user, ok := db.GetSingleUserByEmail(params.Email)

	if !ok || !(internal.CheckPasswordHash(params.Password, user.Password)) {
		errMsg := retError{Error: "User not found"}
		dat, _ := json.Marshal(errMsg)
		w.WriteHeader(401)
		w.Write(dat)
		return
	}

	if params.Expires == 0 {
		params.Expires = 5000
	}


	tokenString, err := internal.CreateJWT(cfg.jwtSecret,map[string]interface{}{
		"Expires": params.Expires, "Subject": strconv.Itoa(user.ID),
	})

	if err != nil {
		errMsg := retError{Error: err.Error()}
		dat, _ := json.Marshal(errMsg)
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		w.Write(dat)
		return
	}

	srcToken := make([]byte, 15)
	_, err = rand.Read(srcToken)

	refreshToken := hex.EncodeToString(srcToken)
	refreshExpiry := time.Now().UTC().Add(time.Duration(1440) * time.Hour)

	if err != nil {
		errMsg := retError{Error: err.Error()}
		dat, _ := json.Marshal(errMsg)
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		w.Write(dat)
		return
	}

	db.UpdateSingleUser(user.ID, internal.UpdateUserParams{
		Email: user.Email, Password: user.Password, RefreshToken:  refreshToken, RefreshExpiry: refreshExpiry,
		}, false)


	res := authRes{
		Email: user.Email,
		ID: user.ID,
		Token: tokenString,
		RefreshToken: refreshToken,
		IsChirpyRed: user.IsChirpyRed,
	}

	dat, _ := json.Marshal(res)
	w.WriteHeader(200)
	w.Write(dat)

}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request, db *internal.DB, cfg *apiConfig) {
	type retError struct {
		Error string `json:"error"`
	}
	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	tokenString := r.Header.Get("Authorization")
	tokenString = strings.Replace(tokenString,"Bearer ","",1)
	token, err := internal.ParseJWT(tokenString, cfg.jwtSecret)
	if err != nil{
		errMsg := retError{Error: "Log in again"}
		dat, _ := json.Marshal(errMsg)
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(401)
		w.Write(dat)
		return
	}
	if subject, ok := token["subject"].(string); ok {
		userID, err := strconv.Atoi(subject)
		if err != nil{
			errMsg := retError{Error: err.Error()}
			dat, _ := json.Marshal(errMsg)
			log.Printf("Error decoding parameters: %s", err)
			w.WriteHeader(401)
			w.Write(dat)
			return
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err = decoder.Decode(&params)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			errMsg := retError{Error: err.Error()}
			dat, _ := json.Marshal(errMsg)
			log.Printf("Error decoding parameters: %s", err)
			w.WriteHeader(401)
			w.Write(dat)
			return
		}

		user, ok := db.UpdateSingleUser(userID, internal.UpdateUserParams{
			Email: params.Email, Password: params.Password,
		}, true)

		if ok{
			dat, _ := json.Marshal(user)
			w.WriteHeader(200)
			w.Write(dat)
			return
		}
	} 
	errMsg := retError{Error: "Cannot find user"}
	dat, _ := json.Marshal(errMsg)
	log.Printf("Error decoding parameters: %s", err)
	w.WriteHeader(404)
	w.Write(dat)
}

func RefreshTokenHandler(w http.ResponseWriter, r *http.Request, db *internal.DB, cfg *apiConfig) {
	type retError struct {
		Error string `json:"error"`
	}
	type TokenRes struct {
		Token string `json:"token"`
	}
	tokenString := r.Header.Get("Authorization")
	tokenString = strings.Replace(tokenString,"Bearer ","",1)
	newToken, err := db.RefreshToken(tokenString, cfg.jwtSecret)
	if err != nil{
		errMsg := retError{Error: err.Error()}
		dat, _ := json.Marshal(errMsg)
		log.Printf("error refreshing token: %s", err)
		w.WriteHeader(401)
		w.Write(dat)
		return
	}
	t := TokenRes{Token: newToken}
	dat, _ := json.Marshal(t)
	w.WriteHeader(200)
	w.Write(dat)
}

func RevokeTokenHandler(w http.ResponseWriter, r *http.Request, db *internal.DB, cfg *apiConfig) {
	type retError struct {
		Error string `json:"error"`
	}
	tokenString := r.Header.Get("Authorization")
	tokenString = strings.Replace(tokenString,"Bearer ","",1)
	err := db.RevokeToken(tokenString)
	if err != nil {
		errMsg := retError{Error: err.Error()}
		dat, _ := json.Marshal(errMsg)
		log.Printf("error revoking token")
		w.WriteHeader(500)
		w.Write(dat)
		return
	}
	w.WriteHeader(204)
}

func DeleteChirpHandler(w http.ResponseWriter, r *http.Request, db *internal.DB, cfg *apiConfig, chirpID int) {
	type retError struct {
		Error string `json:"error"`
	}
	
	tokenString := r.Header.Get("Authorization")
	tokenString = strings.Replace(tokenString,"Bearer ","",1)
	userID, ok := internal.IsAuthenticated(tokenString, cfg.jwtSecret)
	if !ok {
		errMsg := retError{Error: "Log in again"}
		dat, _ := json.Marshal(errMsg)
		log.Printf("Error authenticating")
		w.WriteHeader(401)
		w.Write(dat)
		return
	}
	err := db.DeleteChirp(chirpID, userID)
	if err != nil {
		errMsg := retError{Error: err.Error()}
		log.Printf("Error decoding parameters: %s", err)
		dat, _ := json.Marshal(errMsg)
		w.WriteHeader(403)
		w.Write(dat)
		return
	}
	w.WriteHeader(204)
}

func HandlePolkaWebhook(w http.ResponseWriter, r *http.Request, db *internal.DB, cfg *apiConfig) {
	type WebhookReq struct {
		Event string `json:"event"`
		Data  struct {
			UserID int `json:"user_id"`
		} `json:"data"`
	}

	defer r.Body.Close() // Ensure the request body is closed to free resources
	decoder := json.NewDecoder(r.Body)

	var params WebhookReq
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding parameters: %s", err)
		http.Error(w, "Error decoding parameters", http.StatusInternalServerError)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Upgrade the user here
	if err := db.UpgradeUser(params.Data.UserID); err != nil {
		log.Printf("Error upgrading user %d: %s", params.Data.UserID, err)
		http.Error(w, "User upgrade failed", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}