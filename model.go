package main

import (
	"database/sql"
)

type set struct {
	ID          int     `json:"id"`
	UserID      string  `json:"userId"`
	Weight      float64 `json:"weight"`
	Exercise    string  `json:"exercise"`
	Repetitions int     `json:"repetitions"`
}

func (s *set) getSet(db *sql.DB) error {
	return db.QueryRow("SELECT user_id, weight, exercise, repetitions FROM sets WHERE id=$1",
		s.ID).Scan(&s.UserID, &s.Weight, &s.Exercise, &s.Repetitions)
}

func (s *set) updateSet(db *sql.DB) error {
	_, err :=
		db.Exec("UPDATE sets SET user_id=$2, weight=$3, exercise=$4, repetitions=$5 WHERE id=$1",
			s.ID, s.UserID, s.Weight, s.Exercise, s.Repetitions)

	return err
}

func (s *set) deleteSet(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM sets WHERE id=$1", s.ID)

	return err
}

func (s *set) createSet(db *sql.DB) error {
	err := db.QueryRow(
		"INSERT INTO sets(user_id, weight, exercise, repetitions) VALUES($1, $2, $3, $4) RETURNING id",
		s.UserID, s.Weight, s.Exercise, s.Repetitions).Scan(&s.ID)

	if err != nil {
		return err
	}

	return nil
}

func getSets(db *sql.DB, start, count int) ([]set, error) {
	rows, err := db.Query(
		"SELECT id, user_id, weight, exercise, repetitions FROM sets LIMIT $1 OFFSET $2",
		count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sets := []set{}

	for rows.Next() {
		var s set
		if err := rows.Scan(&s.ID, &s.UserID, &s.Weight, s.Exercise, s.Repetitions); err != nil {
			return nil, err
		}
		sets = append(sets, s)
	}

	return sets, nil
}
