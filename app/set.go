package app

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type set struct {
	ID          int     `json:"id"`
	UserID      string  `json:"userId"`
	Weight      float64 `json:"weight" validate:"required"`
	Exercise    string  `json:"exercise" validate:"required"`
	Repetitions int     `json:"repetitions" validate:"required"`
}

func (s *Server) handleGetSet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Auth
		claims, err := validateToken(w, r)
		if err != nil {
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
		// Auth
		claims, err := validateToken(w, r)
		if err != nil {
			return
		}

		// Logic
		var set set
		count, _ := strconv.Atoi(r.FormValue("count"))
		start, _ := strconv.Atoi(r.FormValue("start"))

		if count > 10 || count < 1 {
			count = 10
		}
		if start < 0 {
			start = 0
		}

		sets, err := set.getSets(s.DB, start, count, claims.UserID)
		if err != nil {
			log.Println(err.Error())
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		respondWithJSON(w, http.StatusOK, sets)
	}
}

func (s *Server) handleCreateSet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Auth
		claims, err := validateToken(w, r)
		if err != nil {
			return
		}

		// Validate claims
		err = s.Validator.Struct(claims)
		if err != nil {
			log.Printf(err.Error())
			respondWithError(w, http.StatusBadRequest, err.Error())
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
		// Auth
		claims, err := validateToken(w, r)
		if err != nil {
			return
		}
		// Validate claims
		err = s.Validator.Struct(claims)
		if err != nil {
			log.Printf(err.Error())
			respondWithError(w, http.StatusBadRequest, err.Error())
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
		}

		set.ID = id

		if err := set.updateSet(s.DB, claims.UserID); err != nil {
			log.Println(err.Error())
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		respondWithJSON(w, http.StatusOK, set)
	}

}

func (s *Server) handleDeleteSet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Auth
		claims, err := validateToken(w, r)
		if err != nil {
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
		if err := set.deleteSet(s.DB, claims.UserID); err != nil {
			log.Println(err.Error())
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
	}
}

func (s *set) getSet(db *sql.DB, userID string) error {
	return db.QueryRow("SELECT user_id, weight, exercise, repetitions FROM sets WHERE id=$1 AND user_id=$2",
		s.ID, userID).Scan(&s.UserID, &s.Weight, &s.Exercise, &s.Repetitions)
}

func (s *set) getSets(db *sql.DB, start, count int, userID string) ([]set, error) {
	rows, err := db.Query(
		"SELECT id, user_id, weight, exercise, repetitions FROM sets WHERE user_id=$1 LIMIT $2 OFFSET $3",
		userID, count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sets := []set{}

	for rows.Next() {
		var s set
		if err := rows.Scan(&s.ID, &s.UserID, &s.Weight, &s.Exercise, &s.Repetitions); err != nil {
			return nil, err
		}
		sets = append(sets, s)
	}

	return sets, nil
}

func (s *set) updateSet(db *sql.DB, userID string) error {
	_, err :=
		db.Exec("UPDATE sets SET user_id=$2, weight=$3, exercise=$4, repetitions=$5 WHERE id=$1 AND user_id=$2",
			s.ID, userID, s.Weight, s.Exercise, s.Repetitions)

	return err
}

func (s *set) deleteSet(db *sql.DB, userID string) error {
	_, err := db.Exec("DELETE FROM sets WHERE id=$1 and user_id=$2", s.ID, userID)

	return err
}

func (s *set) createSet(db *sql.DB, userID string) error {
	err := db.QueryRow(
		"INSERT INTO sets(user_id, weight, exercise, repetitions) VALUES($1, $2, $3, $4) RETURNING id",
		userID, s.Weight, s.Exercise, s.Repetitions).Scan(&s.ID)

	if err != nil {
		return err
	}

	return nil
}
