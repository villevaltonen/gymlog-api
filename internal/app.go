package internal

import (
	"database/sql"
	"fmt"
	"log"

	"net/http"

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
