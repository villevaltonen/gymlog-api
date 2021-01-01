package app

import (
	"log"
	"math/rand"
	"testing"

	"bytes"
	"encoding/json"
	"net/http"
)

func TestEmptyTable(t *testing.T) {
	clearTables()
	createTestUsers()

	req, _ := http.NewRequest("GET", "/api/v1/sets", nil)
	req.AddCookie(authenticate("user1", "password1"))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetNonExistentSet(t *testing.T) {
	clearTables()
	createTestUsers()

	req, _ := http.NewRequest("GET", "/api/v1/sets/11", nil)
	req.AddCookie(authenticate("user1", "password1"))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Set not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Set not found'. Got '%s'", m["error"])
	}
}

func TestGetSet(t *testing.T) {
	clearTables()
	userIDs := createTestUsers()
	addSets(userIDs)

	req, _ := http.NewRequest("GET", "/api/v1/sets/1", nil)
	req.AddCookie(authenticate("user1", "password1"))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestCreateSet(t *testing.T) {
	clearTables()
	createTestUsers()

	var jsonStr = []byte(`{"weight": 111.22, "exercise":"squat", "repetitions":10}`)
	req, _ := http.NewRequest("POST", "/api/v1/sets", bytes.NewBuffer(jsonStr))
	req.AddCookie(authenticate("user1", "password1"))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["weight"] != 111.22 {
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

func TestUpdateSet(t *testing.T) {
	clearTables()
	userIDs := createTestUsers()
	addSets(userIDs)

	req, _ := http.NewRequest("GET", "/api/v1/sets/1", nil)
	req.AddCookie(authenticate("user1", "password1"))
	response := executeRequest(req)
	var originalSet map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalSet)

	var jsonStr = []byte(`{"weight": 222.22, "exercise":"bench", "repetitions":15}`)
	req, _ = http.NewRequest("PUT", "/api/v1/sets/1", bytes.NewBuffer(jsonStr))
	req.AddCookie(authenticate("user1", "password1"))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

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

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("Expected set ID to be '1'. Got '%v'", m["id"])
	}
}

func TestDeleteSet(t *testing.T) {
	clearTables()
	userIDs := createTestUsers()
	addSets(userIDs)

	req, _ := http.NewRequest("GET", "/api/v1/sets/1", nil)
	req.AddCookie(authenticate("user1", "password1"))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/api/v1/sets/1", nil)
	req.AddCookie(authenticate("user1", "password1"))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/api/v1/sets/1", nil)
	req.AddCookie(authenticate("user1", "password1"))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func addSets(userIDs []string) {
	if len(userIDs) == 0 {
		log.Fatal("No userIDs available")
		return
	}

	for _, userID := range userIDs {
		testServer.DB.Exec("INSERT INTO sets(user_id, weight, exercise, repetitions) VALUES($1, $2, $3, $4)", userID, (rand.Intn(5)+1.0)*10, "squat", rand.Intn(5)*2)
	}
}
