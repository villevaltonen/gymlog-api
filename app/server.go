package app

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"

	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"

	// blank to load the driver
	_ "github.com/lib/pq"
)

// Server is an instance of an application with router and db-connection
type Server struct {
	Router    *mux.Router
	DB        *sql.DB
	Validator *validator.Validate
}

// Initialize initializes the app
func (s *Server) Initialize(user, password, dbname, dbhost string) {
	// db connection
	connectionString :=
		fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", dbhost, user, password, dbname)
	var err error
	s.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	// validator
	s.Validator = validator.New()
	// register function to get tag name from json tags.
	s.Validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// router
	s.Router = mux.NewRouter()
	s.routes()
}

// Run starts the HTTP-server
func (s *Server) Run(addr string) {
	log.Println("Starting an HTTP-server!")
	log.Fatal(http.ListenAndServe(":8010", s.Router))
}
