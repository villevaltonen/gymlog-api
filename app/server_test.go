package app

import (
	"log"
	"os"
	"testing"
)

var testServer Server

func TestMain(m *testing.M) {
	CheckEnvVariableExists("JWT_KEY")
	testServer.Initialize(
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_HOST"))

	ensureTablesExist()
	code := m.Run()
	clearTables()
	os.Exit(code)
}

func ensureTablesExist() {
	var tables []string
	tables = append(tables, setsTableCreationQuery)
	tables = append(tables, usersTableCreationQuery)

	for _, table := range tables {
		if _, err := testServer.DB.Exec(table); err != nil {
			log.Fatal(err)
		}
	}
}

func clearTables() {
	testServer.DB.Exec("DELETE FROM sets")
	testServer.DB.Exec("DELETE FROM users")
	testServer.DB.Exec("ALTER SEQUENCE sets_id_seq RESTART WITH 1")
}

const setsTableCreationQuery = `CREATE TABLE IF NOT EXISTS sets
(
    id SERIAL,
    user_id TEXT NOT NULL,
	weight NUMERIC(10,2) NOT NULL DEFAULT 0.00,
	exercise TEXT NOT NULL,
	repetitions INTEGER,
	CONSTRAINT sets_pkey PRIMARY KEY (id)
)`

const usersTableCreationQuery = `CREATE TABLE IF NOT EXISTS users
(
    user_id TEXT NOT NULL,
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    enabled INTEGER NOT NULL DEFAULT 1,
    CONSTRAINT users_pkey PRIMARY KEY (user_id)
)`
