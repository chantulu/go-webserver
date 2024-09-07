package internal

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}
func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

type MyCustomClaims struct {
    jwt.RegisteredClaims
    Subject string `json:"subject"` // Add this field if needed in your claims
}

func CreateJWT(secretKey string, params map[string]interface{}) (string, error) {
    // Create claims with standard claims and optional custom claims
    claims := MyCustomClaims{
        RegisteredClaims: jwt.RegisteredClaims{
            Issuer:   "chirpy",
            IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
            ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(params["Expires"].(int)) * time.Second)),
        },
        Subject: params["Subject"].(string), // Access subject from params if present
    }

    // Create a new token object with HS256 signing method
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    // Sign the token using the secret key
    tokenString, err := token.SignedString([]byte(secretKey))
    if err != nil {
        return "", fmt.Errorf("error signing token: %w", err)
    }

    return tokenString, nil
}

func ParseJWT(tokenString string, secretKey string) (map[string]interface{}, error) {
    // Parse the token with HS256 signing method
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return []byte(secretKey), nil
    })
    if err != nil {
        if err == jwt.ErrSignatureInvalid {
            return nil, fmt.Errorf("invalid token signature")
        }
        return nil, fmt.Errorf("error parsing token: %w", err)
    }

    if !token.Valid {
        return nil, fmt.Errorf("invalid token")
    }

    // Extract claims as a map[string]interface{}
    claims := token.Claims.(jwt.MapClaims)

    // Return the claims as a map
    return claims, nil
}

func IsAuthenticated(tokenString, secret string) (int,bool){
    if tokenString == "" || secret == ""{
        return 0,false
    }
    token, err := ParseJWT(tokenString, secret)
    if err != nil{
		log.Printf("Error decoding parameters: %s", err)
        return 0,false
	}
    if subject, ok := token["subject"].(string); ok {
        userID, err := strconv.Atoi(subject)
        if err != nil{
            log.Printf("Error parsing parameters: %s", err)
            return 0,false
        }
        return userID, true
    }
    return 0,false
}
