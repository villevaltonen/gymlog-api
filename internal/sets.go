package internal

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

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

// Set is a basic entity for holding sets
type Set struct {
	ID          int     `json:"id"`
	UserID      string  `json:"userId"`
	Weight      float64 `json:"weight"`
	Exercise    string  `json:"exercise"`
	Repetitions int     `json:"repetitions"`
}

// GetSet fetches a set from database with id
func (s *Set) GetSet(db *sql.DB) error {
	return db.QueryRow("SELECT user_id, weight, exercise, repetitions FROM sets WHERE id=$1",
		s.ID).Scan(&s.UserID, &s.Weight, &s.Exercise, &s.Repetitions)
}

// UpdateSet executes update query to database
func (s *Set) UpdateSet(db *sql.DB) error {
	_, err :=
		db.Exec("UPDATE sets SET user_id=$2, weight=$3, exercise=$4, repetitions=$5 WHERE id=$1",
			s.ID, s.UserID, s.Weight, s.Exercise, s.Repetitions)

	return err
}

// DeleteSet deletes a set from database with
func (s *Set) DeleteSet(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM sets WHERE id=$1", s.ID)

	return err
}

// CreateSet creates a set into database with given JSON
func (s *Set) CreateSet(db *sql.DB) error {
	err := db.QueryRow(
		"INSERT INTO sets(user_id, weight, exercise, repetitions) VALUES($1, $2, $3, $4) RETURNING id",
		s.UserID, s.Weight, s.Exercise, s.Repetitions).Scan(&s.ID)

	if err != nil {
		return err
	}

	return nil
}

// TODO: user id param
// GetSets fetches multiple sets from database with user id
func GetSets(db *sql.DB, start, count int) ([]Set, error) {
	rows, err := db.Query(
		"SELECT id, user_id, weight, exercise, repetitions FROM sets LIMIT $1 OFFSET $2",
		count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sets := []Set{}

	for rows.Next() {
		var s Set
		if err := rows.Scan(&s.ID, &s.UserID, &s.Weight, s.Exercise, s.Repetitions); err != nil {
			return nil, err
		}
		sets = append(sets, s)
	}

	return sets, nil
}
