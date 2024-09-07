package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users map[int]User `json:"users"`
}

type Chirp struct {
	Body string `json:"body"`
	ID int `json:"id"`
	AuthorID int `json:"author_id"`
}


type User struct {
	Email string `json:"email"`
	ID int `json:"id"`
	Password string `json:"password"`
	RefreshToken string `json:"refresh_token"`
	RefreshExpiry time.Time `json:"refresh_expiry"`
	IsChirpyRed bool `json:"is_chirpy_red"`
}

type UpdateUserParams struct {
    Email        string
    Password     string
    RefreshToken string
	RefreshExpiry time.Time
} 

type UserExternal struct {
	Email string `json:"email"`
	ID int `json:"id"`
	IsChirpyRed bool `json:"is_chirpy_red"`
}


func DbUsertoUserX(dbUser User) UserExternal {
    return UserExternal{
        ID:       dbUser.ID,
        Email:    dbUser.Email,
		IsChirpyRed: dbUser.IsChirpyRed,
    }
}

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error){
	db := &DB{
        path: path,
        mux:  &sync.RWMutex{},
    }
	db.ensureDB()
	return db, nil
}	

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string, userId int) (Chirp, error) {
	chirps, err := db.GetChirps()
	if err != nil {
		return Chirp{}, errors.New("an error occurred getting chirps")
	}
	users, err := db.GetUsers()
	if err != nil {
		return Chirp{}, errors.New("an error occurred getting users")
	}
	sort.Slice(chirps, func(i, j int) bool{
		return chirps[i].ID < chirps[j].ID
	})
	lastIndex := 0
	if len(chirps) > 0 {
		lastIndex = chirps[len(chirps)-1].ID
	}
	newChirp := Chirp{
		ID: lastIndex+1,
		Body: body,
		AuthorID: userId,
	}
	chirps = append(chirps, newChirp)
	dbContent := DBStructure{}
	dbContent.Chirps = make(map[int]Chirp)
	dbContent.Users = make(map[int]User)
	for _,chirp := range chirps{
		dbContent.Chirps[chirp.ID] = chirp
	}
	for _,user := range users{
		dbContent.Users[user.ID] = user
	}
	db.writeDB(dbContent)
	return newChirp, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	dbContent, err := db.loadDB()
	if err != nil {
		return []Chirp{},err
	}
	chirpSlice := []Chirp{}
	for _, chirp := range dbContent.Chirps{
		chirpSlice = append(chirpSlice, chirp)
	}
	return chirpSlice, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetUsers() ([]User, error) {
	dbContent, err := db.loadDB()
	if err != nil {
		return []User{},err
	}
	chirpSlice := []User{}
	for _, user := range dbContent.Users{
		chirpSlice = append(chirpSlice, user)
	}
	return chirpSlice, nil
}


// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateUser(email, password string) (UserExternal, error) {
	users, err := db.GetUsers()
	if err != nil {
		return UserExternal{}, errors.New("an error occurred getting users")
	}
	chirps, err := db.GetChirps()
	if err != nil {
		return UserExternal{}, errors.New("an error occurred getting chirps")
	}
	sort.Slice(users, func(i, j int) bool{
		return users[i].ID < users[j].ID
	})
	lastIndex := 0
	if len(users) > 0 {
		lastIndex = users[len(users)-1].ID
		if _,ok := db.GetSingleUserByEmail(email); ok{
			return UserExternal{}, errors.New("User already exists")
		}
	}
	pass, _ := HashPassword(password)
	newUser := User{
		ID: lastIndex+1,
		Email: email,
		Password: pass,
	}
	users = append(users, newUser)
	dbContent := DBStructure{}
	dbContent.Users = make(map[int]User)
	dbContent.Chirps = make(map[int]Chirp)
	for _,user := range users{
		dbContent.Users[user.ID] = user
	}
	for _,chirp := range chirps{
		dbContent.Chirps[chirp.ID] = chirp
	}
	db.writeDB(dbContent)
	return DbUsertoUserX(newUser), nil
}

func (db *DB) GetSingleUserByEmail(email string) (User, bool) {
	dbstructure,err := db.loadDB()
	if err != nil{
		return  User{},false
	}
	
	for _,user := range dbstructure.Users{
		if user.Email == email{
			return user, true
		}
	}
	return User{},false
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	if _, err := os.Stat(db.path); errors.Is(err, os.ErrNotExist) {
		initialContent := []byte("{\"chirps\":{}, \"users\":{}}")
		err := os.WriteFile(db.path, initialContent, 0666)
		if err != nil {
			log.Fatalf("%v", err)
		}
	}
	return nil
}

func (db *DB) deleteDB() error {
	if _, err := os.Stat(db.path); err == nil {
		err := os.Remove(db.path)
		if err != nil {
			log.Fatalf("%v", err)
		}
	}
	return nil
}

func (db *DB) ResetDB() error {
	err := db.deleteDB()
	if err != nil {
		log.Fatalf("%v", err)
	}
	err = db.ensureDB()
	if err != nil {
		log.Fatalf("%v", err)
	}
	return nil
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error){
	db.mux.Lock()
	defer db.mux.Unlock()
	dbContent := DBStructure{}
	content, err := os.ReadFile(db.path)
	if err != nil {
		return dbContent, errors.New("could not read db file")
	}
	jerr := json.Unmarshal(content, &dbContent)
	if jerr != nil {
		return dbContent, jerr
	}
	return dbContent, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	data, err := json.Marshal(dbStructure)
	if err != nil {
		return errors.New("cannot Marshal file")
	}
	err = os.WriteFile(db.path, data, 0666)
	if err != nil {
		return errors.New("cannot save new chirp to file")
	}
	return nil
}

func (db *DB) GetSingleChirp(id int) (Chirp, bool) {
	dbstructure,err := db.loadDB()
	if err != nil{
		return  Chirp{},false
	}
	chirp, ok := dbstructure.Chirps[id]
	if !ok {
		return Chirp{},false
	}
	return chirp, true
}
func (db *DB) GetSingleUser(id int) (User, bool) {
	dbstructure,err := db.loadDB()
	if err != nil{
		return  User{},false
	}
	user, ok := dbstructure.Users[id]
	if !ok {
		return User{},false
	}
	return user, true
}
func (db *DB) UpdateSingleUser(id int, params UpdateUserParams, hash bool) (UserExternal,bool) {
	dbstructure,err := db.loadDB()
	if err != nil{
		return  UserExternal{},false
	}
	usr, ok := dbstructure.Users[id]
	if !ok {
		return UserExternal{},false
	}

	pass := params.Password

	if hash{
		pass, _ = HashPassword(params.Password)
	}

	refreshToken := params.RefreshToken
	if refreshToken == ""{
		refreshToken = usr.RefreshToken
	}
	refreshExpiry := params.RefreshExpiry
	if refreshExpiry.IsZero() {
		refreshExpiry = usr.RefreshExpiry
	}

	dbstructure.Users[id] = User{
		Email: params.Email,
		Password: pass,
		ID: id,
		RefreshToken: refreshToken,
		RefreshExpiry: refreshExpiry,
	}
	db.writeDB(dbstructure)
	return UserExternal{
		Email: params.Email,
		ID: id,
	},true
}

func (db *DB) RefreshToken(token string, secret string) (string, error){
	dbstructure,err := db.loadDB()
	if err != nil{
		return  "",errors.New("failed to read db")
	}
	
	for _,user := range dbstructure.Users{
		if user.RefreshToken == token && !user.RefreshExpiry.Before(time.Now())  {
			tokenString, err := CreateJWT(secret,map[string]interface{}{
				"Expires": 5000, "Subject": strconv.Itoa(user.ID),
			})
			if err != nil{
				return "",errors.New("failed to generate token")
			}	
			return tokenString, nil
		}
	}
	return "",errors.New("unexpected error")
}
func (db *DB) RevokeToken(token string) (error){
	dbstructure,err := db.loadDB()
	if err != nil{
		return  errors.New("failed to read db")
	}
	
	for _,user := range dbstructure.Users{
		if user.RefreshToken == token{
			_, ok := db.UpdateSingleUser(user.ID, UpdateUserParams{
				Email: user.Email,
				Password: user.Password,
				RefreshToken: "0",
				RefreshExpiry: time.Unix(1,1),
			}, false)
			if ok {
				return nil
			}
		}
	}
	return errors.New("unexpected error")
}

func (db *DB) DeleteChirp(id, userid int) error {
	dbstructure,err := db.loadDB()
	if err != nil{
		return  errors.New("cannot load db");
	}
	chirp, ok := dbstructure.Chirps[id]
	fmt.Println(chirp, userid)
	if ok && chirp.AuthorID == userid {
		delete(dbstructure.Chirps, id)
		db.writeDB(dbstructure)
		return nil
	}
	return errors.New("cannot find matching user")
}

func(db *DB) UpgradeUser(userid int) error {
	dbstructure,err := db.loadDB()
	if err != nil{
		return  errors.New("cannot load db");
	}
	user, ok := dbstructure.Users[userid]
	if ok {
		user.IsChirpyRed = true
		dbstructure.Users[userid] = user // Re-assign the modified user back to the map
	}
	return db.writeDB(dbstructure)
}