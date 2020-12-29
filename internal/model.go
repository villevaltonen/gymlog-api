package internal

import (
	"database/sql"
)

// Set is a basic entity for holding sets
type Set struct {
	ID          int     `json:"id"`
	UserID      string  `json:"userId"`
	Weight      float64 `json:"weight"`
	Exercise    string  `json:"exercise"`
	Repetitions int     `json:"repetitions"`
}

func (s *Set) GetSet(db *sql.DB) error {
	return db.QueryRow("SELECT user_id, weight, exercise, repetitions FROM sets WHERE id=$1",
		s.ID).Scan(&s.UserID, &s.Weight, &s.Exercise, &s.Repetitions)
}

func (s *Set) UpdateSet(db *sql.DB) error {
	_, err :=
		db.Exec("UPDATE sets SET user_id=$2, weight=$3, exercise=$4, repetitions=$5 WHERE id=$1",
			s.ID, s.UserID, s.Weight, s.Exercise, s.Repetitions)

	return err
}

func (s *Set) DeleteSet(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM sets WHERE id=$1", s.ID)

	return err
}

func (s *Set) CreateSet(db *sql.DB) error {
	err := db.QueryRow(
		"INSERT INTO sets(user_id, weight, exercise, repetitions) VALUES($1, $2, $3, $4) RETURNING id",
		s.UserID, s.Weight, s.Exercise, s.Repetitions).Scan(&s.ID)

	if err != nil {
		return err
	}

	return nil
}

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
