package app

import (
	"database/sql"
	"fmt"
	"log"

	"net/http"

	"github.com/gorilla/mux"
	// blank to load the driver
	_ "github.com/lib/pq"
)

// Server is an instance of an application with router and db-connection
type Server struct {
	Router *mux.Router
	DB     *sql.DB
}

// Initialize initializes the app
func (s *Server) Initialize(user, password, dbname string) {
	connectionString :=
		fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbname)

	var err error
	s.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	s.Router = mux.NewRouter()
	s.routes()
}

// Run starts the HTTP-server
func (s *Server) Run(addr string) {
	log.Println("Starting an HTTP-server!")
	log.Fatal(http.ListenAndServe(":8010", s.Router))
}
