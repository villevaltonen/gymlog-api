package app

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func (s *Server) routes() {
	// Authentication
	s.Router.HandleFunc("/api/login", s.corsHandler(s.logHTTP(s.handleLogin()))).Methods(http.MethodPost, http.MethodOptions)
	s.Router.HandleFunc("/api/refresh", s.corsHandler(s.authenticate(s.logHTTP(s.handleRefresh())))).Methods(http.MethodPost, http.MethodOptions)
	s.Router.HandleFunc("/api/register", s.corsHandler(s.logHTTP(s.handleRegister()))).Methods(http.MethodPost, http.MethodOptions)

	// Heartbeat
	s.Router.HandleFunc("/api/heartbeat", s.corsHandler(s.authenticate(s.logHTTP(s.handleHeartbeat())))).Methods(http.MethodGet, http.MethodOptions)

	// Manage sets
	s.Router.HandleFunc("/api/v1/sets", s.corsHandler(s.authenticate(s.logHTTP(s.handleGetSets())))).Methods(http.MethodGet, http.MethodOptions)
	s.Router.HandleFunc("/api/v1/sets", s.corsHandler(s.authenticate(s.logHTTP(s.handleCreateSet())))).Methods(http.MethodPost, http.MethodOptions)
	s.Router.HandleFunc("/api/v1/sets/{id:[0-9]+}", s.corsHandler(s.authenticate(s.logHTTP(s.handleGetSet())))).Methods(http.MethodGet, http.MethodOptions)
	s.Router.HandleFunc("/api/v1/sets/{id:[0-9]+}", s.corsHandler(s.authenticate(s.logHTTP(s.handleUpdateSet())))).Methods(http.MethodPut, http.MethodOptions)
	s.Router.HandleFunc("/api/v1/sets/{id:[0-9]+}", s.corsHandler(s.authenticate(s.logHTTP(s.handleDeleteSet())))).Methods(http.MethodDelete, http.MethodOptions)
}

func (s *Server) corsHandler(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			return
		}
		h.ServeHTTP(w, r)
	}
}

func (s *Server) authenticate(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := &Claims{}

		// Check that cookie is present
		c, err := r.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				log.Println("No token cookie present")
				respondWithError(w, http.StatusUnauthorized, "No token cookie present")
				return
			}
			respondWithError(w, http.StatusBadRequest, "Invalid cookie")
			return
		}

		// Validate token
		tkn, err := jwt.ParseWithClaims(c.Value, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				log.Println("Err sign invalid")
				respondWithError(w, http.StatusUnauthorized, "Err sign invalid")
				return
			}
			respondWithError(w, http.StatusBadRequest, "Invalid cookie")
			return
		}
		if !tkn.Valid {
			log.Println("Invalid token")
			respondWithError(w, http.StatusBadRequest, "Invalid token")
			return
		}

		// Validate claims
		err = s.Validator.Struct(claims)
		if err != nil {
			log.Printf(err.Error())
			respondWithError(w, http.StatusBadRequest, "Invalid token")
			return
		}
		h.ServeHTTP(w, r)
	}
}

func (s *Server) logHTTP(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := statusWriter{ResponseWriter: w}
		h.ServeHTTP(&wrapped, r)
		log.Printf("Request: %s %s %s %s", r.Method, r.URL.EscapedPath(), strconv.Itoa(wrapped.status), time.Since(start).String())
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}
