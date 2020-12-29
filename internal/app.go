package internal

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// Application is an instance of an application with router and db-connection
type Application struct {
	Router *mux.Router
	DB     *sql.DB
}

// Initialize initializes the app
func (a *Application) Initialize(user, password, dbname string) {
	connectionString :=
		fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbname)

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	a.Router = mux.NewRouter()

	a.initializeRoutes()
}

// Run starts the HTTP-server
func (a *Application) Run(addr string) {
	log.Println("Starting an HTTP-server!")
	log.Fatal(http.ListenAndServe(":8010", a.Router))
}

func (a *Application) getSet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusBadRequest, "Invalid set ID")
		return
	}

	s := Set{ID: id}
	if err := s.GetSet(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			log.Println(err.Error())
			respondWithError(w, http.StatusNotFound, "Set not found")
		default:
			log.Println(err.Error())
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, s)
}

func (a *Application) getSets(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	products, err := GetSets(a.DB, start, count)
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, products)
}

func (a *Application) createSet(w http.ResponseWriter, r *http.Request) {
	var s Set
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&s); err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := s.CreateSet(a.DB); err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, s)
}

func (a *Application) updateSet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusBadRequest, "Invalid set ID")
		return
	}

	var s Set
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&s); err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	s.ID = id

	if err := s.UpdateSet(a.DB); err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, s)
}

func (a *Application) deleteSet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusBadRequest, "Invalid Set ID")
		return
	}

	s := Set{ID: id}
	if err := s.DeleteSet(a.DB); err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

// utils

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (a *Application) initializeRoutes() {
	a.Router.HandleFunc("/api/v1/sets", a.getSets).Methods("GET")
	a.Router.HandleFunc("/api/v1/sets", a.createSet).Methods("POST")
	a.Router.HandleFunc("/api/v1/sets/{id:[0-9]+}", a.getSet).Methods("GET")
	a.Router.HandleFunc("/api/v1/sets/{id:[0-9]+}", a.updateSet).Methods("PUT")
	a.Router.HandleFunc("/api/v1/sets/{id:[0-9]+}", a.deleteSet).Methods("DELETE")
	a.Router.HandleFunc("/api/login", Login).Methods("POST")
	a.Router.HandleFunc("/api/welcome", Welcome).Methods("GET")
	a.Router.HandleFunc("/api/refresh", Refresh).Methods("POST")
}

// jwt

// switch to env variable
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
	expirationTime := time.Now().Add(5 * time.Minute)
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
	// We can obtain the session token from auth requests cookies
	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If cookie is not set, set unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// For any other type of request return bad request
		w.WriteHeader(http.StatusBadRequest)
	}

	// Get JWT string from token
	tknStr := c.Value

	// Initialize a new instance of claims
	claims := &Claims{}

	// Parse the JWT string and store the result in `claims`.
	// Note that we are passing the key in this method as well. This method will return an error
	// if the token is invalid (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Finally return welcome to the user along with the username
	w.Write([]byte(fmt.Sprintf("Welcome %s!", claims.Username)))
}

// Refresh provides a way to refresh a JWT
func Refresh(w http.ResponseWriter, r *http.Request) {
	// (BEGIN) The code uptil this point is the same as the first part of the `Welcome` route
	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			fmt.Println("No cookie present")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tknStr := c.Value
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			fmt.Println("Err sign invalid")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !tkn.Valid {
		fmt.Println("token invalid")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// (END) The code up-till this point is the same as the first part of the `Welcome` route

	// We ensure that a new token is not issued until enough time has elapsed
	// In this case, a new token will only be issued if the old token is within
	// 30 seconds of expiry. Otherwise, return a bad request status
	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 30*time.Second {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Now, create a new token for the current use, with a renewed expiration time
	expirationTime := time.Now().Add(5 * time.Minute)
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
