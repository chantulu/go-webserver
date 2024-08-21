package internal

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sort"
	"sync"
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
}


type User struct {
	Email string `json:"email"`
	ID int `json:"id"`
	Password string `json:"password"`
}

type UserExternal struct {
	Email string `json:"email"`
	ID int `json:"id"`
}


func DbUsertoUserX(dbUser User) UserExternal {
    return UserExternal{
        ID:       dbUser.ID,
        Email:    dbUser.Email,
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
func (db *DB) CreateChirp(body string) (Chirp, error) {
	chirps, err := db.GetChirps()
	if err != nil {
		return Chirp{}, errors.New("an error occurred getting chirps")
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
	}
	chirps = append(chirps, newChirp)
	dbContent := DBStructure{}
	dbContent.Chirps = make(map[int]Chirp)
	for _,chirp := range chirps{
		dbContent.Chirps[chirp.ID] = chirp
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
	for _,user := range users{
		dbContent.Users[user.ID] = user
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