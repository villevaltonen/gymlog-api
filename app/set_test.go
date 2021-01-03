package app

import (
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	"bytes"
	"encoding/json"
	"net/http"
)

func TestEmptyTable(t *testing.T) {
	clearTables()
	createTestUsers()

	req, _ := http.NewRequest("GET", "/api/v1/sets", nil)
	req.AddCookie(authenticate("user1@localhost.com", "password1"))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	// Numbers are compared to floats because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["results"] != 0.0 {
		t.Errorf("Expected results to be '0'. Got '%v'", m["results"])
	}

	if m["skip"] != 0.0 {
		t.Errorf("Expected skip to be '0'. Got '%v'", m["skip"])
	}

	if m["limit"] != 10.0 {
		t.Errorf("Expected limit to be '10'. Got '%v'", m["limit"])
	}

	sets := fmt.Sprintf("%v", m["sets"])
	if sets != "[]" {
		t.Errorf("Expected an empty array. Got %s", sets)
	}
}

func TestGetNonExistentSet(t *testing.T) {
	clearTables()
	createTestUsers()

	req, _ := http.NewRequest("GET", "/api/v1/sets/11", nil)
	req.AddCookie(authenticate("user1@localhost.com", "password1"))
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
	req.AddCookie(authenticate("user1@localhost.com", "password1"))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestGetSets(t *testing.T) {
	clearTables()
	userIDs := createTestUsers()
	addSets(userIDs)

	req, _ := http.NewRequest("GET", "/api/v1/sets", nil)
	req.AddCookie(authenticate("user1@localhost.com", "password1"))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestCreateSet(t *testing.T) {
	clearTables()
	createTestUsers()

	// Valid request
	var jsonStr1 = []byte(`{"weight": 111.22, "exercise":"squat", "repetitions":10}`)
	req, _ := http.NewRequest("POST", "/api/v1/sets", bytes.NewBuffer(jsonStr1))
	req.AddCookie(authenticate("user1@localhost.com", "password1"))
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

	// Numbers are compared to floats because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["repetitions"] != 10.0 {
		t.Errorf("Expected repetitions to be '10'. Got '%v'", m["repetitions"])
	}

	if m["id"] != 1.0 {
		t.Errorf("Expected set ID to be '1'. Got '%v'", m["id"])
	}

	// Invalid request
	var jsonStr2 = []byte(`{"exercise":"squat", "repetitions":10}`)
	req, _ = http.NewRequest("POST", "/api/v1/sets", bytes.NewBuffer(jsonStr2))
	req.AddCookie(authenticate("user1@localhost.com", "password1"))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestUpdateSet(t *testing.T) {
	clearTables()
	userIDs := createTestUsers()
	addSets(userIDs)

	req, _ := http.NewRequest("GET", "/api/v1/sets/1", nil)
	req.AddCookie(authenticate("user1@localhost.com", "password1"))
	response := executeRequest(req)
	var originalSet map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalSet)

	var jsonStr = []byte(`{"weight": 222.22, "exercise":"bench", "repetitions":15}`)
	req, _ = http.NewRequest("PUT", "/api/v1/sets/1", bytes.NewBuffer(jsonStr))
	req.AddCookie(authenticate("user1@localhost.com", "password1"))
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

	// Numbers are compared to floats because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("Expected set ID to be '1'. Got '%v'", m["id"])
	}

	// Invalid request
	var jsonStr2 = []byte(`{"exercise":"squat", "repetitions":10}`)
	req, _ = http.NewRequest("PUT", "/api/v1/sets/1", bytes.NewBuffer(jsonStr2))
	req.AddCookie(authenticate("user1@localhost.com", "password1"))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestDeleteSet(t *testing.T) {
	clearTables()
	userIDs := createTestUsers()
	addSets(userIDs)

	// Check that a set with id 1 exists
	req, _ := http.NewRequest("GET", "/api/v1/sets/1", nil)
	req.AddCookie(authenticate("user1@localhost.com", "password1"))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Try to delete set with other user
	req, _ = http.NewRequest("DELETE", "/api/v1/sets/1", nil)
	req.AddCookie(authenticate("user2@localhost.com", "password2"))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)

	// Check that a set with id 1 still exists
	req, _ = http.NewRequest("GET", "/api/v1/sets/1", nil)
	req.AddCookie(authenticate("user1@localhost.com", "password1"))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Delete set with correct user
	req, _ = http.NewRequest("DELETE", "/api/v1/sets/1", nil)
	req.AddCookie(authenticate("user1@localhost.com", "password1"))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Check that set has been deleted
	req, _ = http.NewRequest("GET", "/api/v1/sets/1", nil)
	req.AddCookie(authenticate("user1@localhost.com", "password1"))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func addSets(userIDs []string) {
	if len(userIDs) == 0 {
		log.Fatal("No userIDs available")
		return
	}

	current := time.Now()
	for _, userID := range userIDs {
		_, err := testServer.DB.Exec("INSERT INTO sets(user_id, weight, exercise, repetitions, created, modified) VALUES($1, $2, $3, $4, $5, $6)", userID, (rand.Intn(5)+1.0)*10, "squat", rand.Intn(5)*2, current, current)
		if err != nil {
			log.Fatal(err.Error())
			break
		}
	}
}
