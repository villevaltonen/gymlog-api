// main_test.go

package main

import (
	"log"
	"os"
	"testing"

	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
)

var a Application

func TestMain(m *testing.M) {
	a.Initialize(
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"))

	ensureTableExists()
	code := m.Run()
	clearTable()
	os.Exit(code)
}

func TestEmptyTable(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/api/v1/sets", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetNonExistentSet(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/api/v1/sets/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Set not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Set not found'. Got '%s'", m["error"])
	}
}

func TestCreateSet(t *testing.T) {

	clearTable()

	var jsonStr = []byte(`{"userId":"test user id", "weight": 11.22, "exercise":"squat", "repetitions":10}`)
	req, _ := http.NewRequest("POST", "/api/v1/sets", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["userId"] != "test user id" {
		t.Errorf("Expected user id to be 'test user id'. Got '%v'", m["userId"])
	}

	if m["weight"] != 11.22 {
		t.Errorf("Expected weight to be '11.22'. Got '%v'", m["weight"])
	}

	if m["exercise"] != "squat" {
		t.Errorf("Expected exercise to be 'squat'. Got '%v'", m["exercise"])
	}

	// repetitions is compared to 10.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["repetitions"] != 10.0 {
		t.Errorf("Expected repetitions to be '10'. Got '%v'", m["repetitions"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("Expected set ID to be '1'. Got '%v'", m["id"])
	}
}

func TestGetSet(t *testing.T) {
	clearTable()
	addSets(1)

	req, _ := http.NewRequest("GET", "/api/v1/sets/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestUpdateSet(t *testing.T) {

	clearTable()
	addSets(1)

	req, _ := http.NewRequest("GET", "/api/v1/sets/1", nil)
	response := executeRequest(req)
	var originalSet map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalSet)

	var jsonStr = []byte(`{"user_id":"test user updated id", "weight": 22.22, "exercise":"bench", "repetitions":15}`)
	req, _ = http.NewRequest("PUT", "/api/v1/sets/1", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != originalSet["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalSet["id"], m["id"])
	}

	if m["userId"] == originalSet["userId"] {
		t.Errorf("Expected the userId to change from '%v' to '%v'. Got '%v'", originalSet["userId"], m["userId"], m["userId"])
	}

	if m["weight"] == originalSet["weight"] {
		t.Errorf("Expected the weight to change from '%v' to '%v'. Got '%v'", originalSet["weight"], m["weight"], m["weight"])
	}

	if m["exercise"] == originalSet["exercise"] {
		t.Errorf("Expected the exercise to change from '%v' to '%v'. Got '%v'", originalSet["exercise"], m["exercise"], m["exercise"])
	}

	if m["repetitions"] == originalSet["repetitions"] {
		t.Errorf("Expected the repetitions to change from '%v' to '%v'. Got '%v'", originalSet["repetitions"], m["repetitions"], m["repetitions"])
	}
}

func TestDeleteSet(t *testing.T) {
	clearTable()
	addSets(1)

	req, _ := http.NewRequest("GET", "/api/v1/sets/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/api/v1/sets/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/api/v1/sets/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

// utils

func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM sets")
	a.DB.Exec("ALTER SEQUENCE sets_id_seq RESTART WITH 1")
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS sets
(
    id SERIAL,
    user_id TEXT NOT NULL,
	weight NUMERIC(3,1) NOT NULL DEFAULT 0.00,
	exercise TEXT NOT NULL,
	repetitions INTEGER,
	CONSTRAINT sets_pkey PRIMARY KEY (id)
)`

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func addSets(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		a.DB.Exec("INSERT INTO sets(user_id, weight, exercise, repetitions) VALUES($1, $2, $3, $4)", "User ID "+strconv.Itoa(i), (i+1.0)*10, "squat", count*2)
	}
}
