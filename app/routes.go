package app

import (
	"log"
	"net/http"
	"strconv"
	"time"
)

func (s *Server) routes() {
	// authentication
	s.Router.HandleFunc("/api/login", s.logHTTP(s.handleLogin())).Methods("POST")
	s.Router.HandleFunc("/api/refresh", s.logHTTP(s.handleRefresh())).Methods("POST")
	s.Router.HandleFunc("/api/register", s.logHTTP(s.handleRegister())).Methods("POST")

	// heartbeat
	s.Router.HandleFunc("/api/heartbeat", s.logHTTP(s.handleHeartbeat())).Methods("GET")

	// manage sets
	s.Router.HandleFunc("/api/v1/sets", s.logHTTP(s.handleGetSets())).Methods("GET")
	s.Router.HandleFunc("/api/v1/sets", s.logHTTP(s.handleCreateSet())).Methods("POST")
	s.Router.HandleFunc("/api/v1/sets/{id:[0-9]+}", s.logHTTP(s.handleGetSet())).Methods("GET")
	s.Router.HandleFunc("/api/v1/sets/{id:[0-9]+}", s.logHTTP(s.handleUpdateSet())).Methods("PUT")
	s.Router.HandleFunc("/api/v1/sets/{id:[0-9]+}", s.logHTTP(s.handleDeleteSet())).Methods("DELETE")
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
