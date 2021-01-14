package app

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type set struct {
	ID          int       `json:"id"`
	UserID      string    `json:"userId"`
	Weight      float64   `json:"weight" validate:"required"`
	Exercise    string    `json:"exercise" validate:"required"`
	Repetitions int       `json:"repetitions" validate:"required"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
}

type sets struct {
	Results int   `json:"results"`
	Skip    int   `json:"skip"`
	Limit   int   `json:"limit"`
	Sets    []set `json:"sets"`
}

func (s *Server) handleGetSet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user information
		claims, err := parseTokenCookie(w, r)
		if err != nil {
			log.Println(err.Error())
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Logic
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			log.Println(err.Error())
			respondWithError(w, http.StatusBadRequest, "Invalid set ID")
			return
		}

		set := set{ID: id}
		if err := set.getSet(s.DB, claims.UserID); err != nil {
			switch err {
			case sql.ErrNoRows:
				log.Println(err.Error())
				respondWithError(w, http.StatusNotFound, "Set not found")
			default:
				log.Println(err.Error())
				respondWithError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		respondWithJSON(w, http.StatusOK, set)
	}
}

func (s *Server) handleGetSets() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user information
		claims, err := parseTokenCookie(w, r)
		if err != nil {
			log.Println(err.Error())
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Logic
		var set set
		skip, _ := strconv.Atoi(r.FormValue("skip"))
		limit, _ := strconv.Atoi(r.FormValue("limit"))

		if limit > 10 || limit < 1 {
			limit = 10
		}
		if skip < 0 {
			skip = 0
		}

		sets := sets{Skip: skip, Limit: limit}
		result, err := set.getSets(s.DB, skip, limit, claims.UserID)
		if err != nil {
			log.Println(err.Error())
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		sets.Sets = result
		sets.Results = len(result)
		respondWithJSON(w, http.StatusOK, sets)
	}
}

func (s *Server) handleCreateSet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user information
		claims, err := parseTokenCookie(w, r)
		if err != nil {
			log.Println(err.Error())
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Logic
		var set set
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&set); err != nil {
			log.Println(err.Error())
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
		defer r.Body.Close()

		// Validate set
		err = s.Validator.Struct(set)
		if err != nil {
			log.Printf(err.Error())
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := set.createSet(s.DB, claims.UserID); err != nil {
			log.Println(err.Error())
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		respondWithJSON(w, http.StatusCreated, set)
	}
}

func (s *Server) handleUpdateSet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user information
		claims, err := parseTokenCookie(w, r)
		if err != nil {
			log.Println(err.Error())
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Logic
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			log.Println(err.Error())
			respondWithError(w, http.StatusBadRequest, "Invalid set ID")
			return
		}

		var set set
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&set); err != nil {
			log.Println(err.Error())
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
		defer r.Body.Close()

		// Validate set
		err = s.Validator.Struct(set)
		if err != nil {
			log.Printf(err.Error())
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		set.ID = id
		affectedRows, err := set.updateSet(s.DB, claims.UserID)
		if err != nil {
			log.Println(err.Error())
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		if affectedRows == 0 {
			respondWithError(w, http.StatusNotFound, "Not found")
			return
		}

		respondWithJSON(w, http.StatusOK, set)
	}

}

func (s *Server) handleDeleteSet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user information
		claims, err := parseTokenCookie(w, r)
		if err != nil {
			log.Println(err.Error())
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Logic
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			log.Println(err.Error())
			respondWithError(w, http.StatusBadRequest, "Invalid Set ID")
			return
		}

		set := set{ID: id}
		affectedRows, err := set.deleteSet(s.DB, claims.UserID)
		if err != nil {
			log.Println(err.Error())
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		if affectedRows == 0 {
			respondWithError(w, http.StatusNotFound, "Not found")
			return
		}
		respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
	}
}

func (s *set) getSet(db *sql.DB, userID string) error {
	return db.QueryRow("SELECT user_id, weight, exercise, repetitions, created, modified FROM sets WHERE id=$1 AND user_id=$2",
		s.ID, userID).Scan(&s.UserID, &s.Weight, &s.Exercise, &s.Repetitions, &s.Created, &s.Modified)
}

func (s *set) getSets(db *sql.DB, start, count int, userID string) ([]set, error) {
	rows, err := db.Query(
		"SELECT id, user_id, weight, exercise, repetitions, created, modified FROM sets WHERE user_id=$1 ORDER BY created DESC LIMIT $2 OFFSET $3",
		userID, count, start)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sets := []set{}
	for rows.Next() {
		var s set
		if err := rows.Scan(&s.ID, &s.UserID, &s.Weight, &s.Exercise, &s.Repetitions, &s.Created, &s.Modified); err != nil {
			return nil, err
		}
		sets = append(sets, s)
	}

	return sets, nil
}

func (s *set) updateSet(db *sql.DB, userID string) (int64, error) {
	result, err :=
		db.Exec("UPDATE sets SET weight=$3, exercise=$4, repetitions=$5, modified=$6 WHERE id=$1 AND user_id=$2",
			s.ID, userID, s.Weight, s.Exercise, s.Repetitions, time.Now())

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, err
}

func (s *set) deleteSet(db *sql.DB, userID string) (int64, error) {
	result, err := db.Exec("DELETE FROM sets WHERE id=$1 and user_id=$2", s.ID, userID)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, err
}

func (s *set) createSet(db *sql.DB, userID string) error {
	current := time.Now()
	err := db.QueryRow(
		"INSERT INTO sets(user_id, weight, exercise, repetitions, created, modified) VALUES($1, $2, $3, $4, $5, $6) RETURNING id, created, modified",
		userID, s.Weight, s.Exercise, s.Repetitions, current, current).Scan(&s.ID, &s.Created, &s.Modified)

	if err != nil {
		return err
	}

	return nil
}
