package app

import (
	"log"
	"os"
	"testing"
)

var testServer Server

func TestMain(m *testing.M) {
	testServer.Initialize(
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"))

	ensureTablesExist()
	code := m.Run()
	clearTables()
	os.Exit(code)
}

func ensureTablesExist() {
	if _, err := testServer.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTables() {
	testServer.DB.Exec("DELETE FROM sets")
	testServer.DB.Exec("ALTER SEQUENCE sets_id_seq RESTART WITH 1")
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS sets
(
    id SERIAL,
    user_id TEXT NOT NULL,
	weight NUMERIC(10,2) NOT NULL DEFAULT 0.00,
	exercise TEXT NOT NULL,
	repetitions INTEGER,
	CONSTRAINT sets_pkey PRIMARY KEY (id)
)`
