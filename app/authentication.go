package app

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte(os.Getenv("JWT_KEY"))

type user struct {
	Password string `json:"password" validate:"required"`
	Username string `json:"username" validate:"required,email"`
	UserID   string `json:"userId"`
}

// Claims is a struct for JWT cookie
// Includes embedded type jwt.StandardClaims to provide additional fields like expiry time
type Claims struct {
	Username string `json:"username" validate:"required"`
	UserID   string `json:"userId" validate:"required"`
	jwt.StandardClaims
}

func (s *Server) handleLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds user

		// Decode JSON body to get credentials
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			// invalid structure results to HTTP error
			log.Println(err.Error())
			respondWithError(w, http.StatusBadRequest, "Can't decode credentials, check the structure")
			return
		}

		// Validate user input
		err = s.Validator.Struct(creds)
		if err != nil {
			log.Printf(err.Error())
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Authenticate user
		user := creds
		if err := user.getUserByUsername(s.DB); err != nil {
			switch err {
			case sql.ErrNoRows:
				log.Println(err.Error())
				respondWithError(w, http.StatusNotFound, "User not found")
				return
			default:
				log.Println(err.Error())
				respondWithError(w, http.StatusInternalServerError, "Internal server error")
				return
			}
		}

		// Check password: match => continue, not match => unauthorized
		if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Create JWT claims, which include username and expiration time
		expirationTime := time.Now().Add(1 * time.Minute)
		claims := &Claims{
			Username: user.Username,
			UserID:   user.UserID,
			StandardClaims: jwt.StandardClaims{
				// In JWT, expiration time is given as unix milliseconds
				ExpiresAt: expirationTime.Unix(),
			},
		}

		// Declare the token with algorithm used for signing and the claims
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		// Create JWT string
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			// In case of error, return internal server error
			log.Println(err.Error())
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Finally, we set the client cookie for token and use the same expiration time
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    tokenString,
			Expires:  expirationTime,
			HttpOnly: true,
		})

		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) handleRefresh() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validate token
		claims, err := validateToken(w, r)
		if err != nil {
			return
		}

		// Validate claims
		err = s.Validator.Struct(claims)
		if err != nil {
			log.Printf(err.Error())
			respondWithError(w, http.StatusBadRequest, err.Error())
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
			log.Println(err.Error())
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Set the new token as the users `token` cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    tokenString,
			Expires:  expirationTime,
			HttpOnly: true,
		})
	}
}

func (s *Server) handleRegister() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds user

		// Decode JSON body to get credentials
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			// invalid structure results to HTTP error
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Validate user input
		err = s.Validator.Struct(creds)
		if err != nil {
			log.Printf(err.Error())
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Check if username is already taken
		exists, err := creds.checkIfUserExists(s.DB)
		if err != nil {
			log.Println(err.Error())
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		if exists {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Hash with bcrypt
		// The second argument is the cost of hashing, which we arbitrarily set as 8 (this value can be more or less, depending on the computing power you wish to utilize)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)
		if err != nil {
			log.Println(err.Error())
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Generate userID as UUID v4
		userID, err := uuid.NewRandom()
		if err != nil {
			log.Println(err.Error())
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Insert credentials into database
		err = creds.createUser(s.DB, userID.String(), string(hashedPassword))
		if err != nil {
			log.Println(err.Error())
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
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

func (c *user) createUser(db *sql.DB, userID, hashedPassword string) error {
	_, err := db.Exec(
		"INSERT INTO users(user_id, username, password) VALUES($1, $2, $3)",
		userID, c.Username, hashedPassword)

	if err != nil {
		return err
	}

	return nil
}

func (c *user) checkIfUserExists(db *sql.DB) (bool, error) {
	rows, err := db.Query(
		"SELECT COUNT(username) FROM users WHERE username=$1",
		c.Username)
	if err != nil {
		return true, err
	}
	defer rows.Close()

	count := 0

	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			log.Println(err.Error())
			return true, err
		}
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}

func (c *user) getUserByUsername(db *sql.DB) error {
	return db.QueryRow(
		"SELECT user_id, username, password FROM users WHERE username=$1",
		c.Username).Scan(&c.UserID, &c.Username, &c.Password)
}
