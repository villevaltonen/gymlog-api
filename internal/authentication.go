package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// TODO: switch to env variable
var jwtKey = []byte("my_secret_key")

var users = map[string]string{
	"user1": "password1",
	"user2": "password2",
}

// Credentials contain username and password
type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

// Claims is a struct, which is used in JWT cookie
// Includes embedded type jwt.StandardClaims to provide additional fields like expiry time
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// Login handles login requests
func Login(w http.ResponseWriter, r *http.Request) {
	var creds Credentials

	// Decode JSON body to get credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		// invalid structure results to HTTP error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get expected password from our in memory map
	// TODO: Use database
	expectedPassword, ok := users[creds.Username]

	// If password exists for the given user
	// AND, if it is the same as in request body, we can move ahead
	// if NOT, return unauthorized status
	if !ok || expectedPassword != creds.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Declare expiration time for token
	expirationTime := time.Now().Add(1 * time.Minute)
	// Create JWT claims, which include username and expiration time
	claims := &Claims{
		Username: creds.Username,
		StandardClaims: jwt.StandardClaims{
			// in JWT, expiration time is given as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Declare the token with algorithm used for signing and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create JWT string
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		// In case of error, return internal server error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Finally, we set the client cookie for token and use the same expiration time
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
}

// Welcome is a handler for a test API for authentication
func Welcome(w http.ResponseWriter, r *http.Request) {
	claims, err := validateToken(w, r)
	if err != nil {
		return
	}

	// Finally return welcome to the user along with the username
	w.Write([]byte(fmt.Sprintf("Welcome %s!", claims.Username)))
}

// Refresh provides a way to refresh a JWT
func Refresh(w http.ResponseWriter, r *http.Request) {
	// Validate token
	claims, err := validateToken(w, r)
	if err != nil {
		return
	}

	// We ensure that a new token is not issued until enough time has elapsed
	// In this case, a new token will only be issued if the old token is within
	// 30 seconds of expiry. Otherwise, return a bad request status
	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 30*time.Second {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Now, create a new token for the current use, with a renewed expiration time
	expirationTime := time.Now().Add(1 * time.Minute)
	claims.ExpiresAt = expirationTime.Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the new token as the users `token` cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
}

func validateToken(w http.ResponseWriter, r *http.Request) (*Claims, error) {
	claims := &Claims{}

	// Check that cookie is present
	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			log.Println("No cookie present")
			w.WriteHeader(http.StatusUnauthorized)
			return claims, err
		}
		w.WriteHeader(http.StatusBadRequest)
		return claims, err
	}

	// Validate token
	tkn, err := jwt.ParseWithClaims(c.Value, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			log.Println("Err sign invalid")
			w.WriteHeader(http.StatusUnauthorized)
			return claims, err
		}
		w.WriteHeader(http.StatusBadRequest)
		return claims, err
	}
	if !tkn.Valid {
		log.Println("token invalid")
		w.WriteHeader(http.StatusUnauthorized)
		return claims, err
	}
	return claims, nil
}
