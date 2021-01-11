package app

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestLogin(t *testing.T) {
	clearTables()
	createTestUsers()

	// Correct credentials
	var jsonStr1 = []byte(`{"username":"user1@localhost.com", "password": "password1"}`)
	req, _ := http.NewRequest("POST", "/api/users/login", bytes.NewBuffer(jsonStr1))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Incorrect password
	var jsonStr2 = []byte(`{"username":"user1@localhost.com", "password": "passwordnotcorrect"}`)
	req, _ = http.NewRequest("POST", "/api/users/login", bytes.NewBuffer(jsonStr2))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusUnauthorized, response.Code)

	// Username not found
	var jsonStr3 = []byte(`{"username":"usernotfound@localhost.com", "password": "password1"}`)
	req, _ = http.NewRequest("POST", "/api/users/login", bytes.NewBuffer(jsonStr3))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)

	// Incorrect username
	var jsonStr4 = []byte(`{"username":"invalidemail", "password": "password1"}`)
	req, _ = http.NewRequest("POST", "/api/users/login", bytes.NewBuffer(jsonStr4))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestRegister(t *testing.T) {
	clearTables()
	createTestUsers()

	// New user
	var jsonStr1 = []byte(`{"username":"user3@localhost.com", "password": "password3"}`)
	req, _ := http.NewRequest("POST", "/api/users/register", bytes.NewBuffer(jsonStr1))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	// Duplicate user
	var jsonStr2 = []byte(`{"username":"user1@localhost.com", "password": "password1"}`)
	req, _ = http.NewRequest("POST", "/api/users/register", bytes.NewBuffer(jsonStr2))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, response.Code)

	// Incorrect username
	var jsonStr3 = []byte(`{"username":"user3", "password": "password3"}`)
	req, _ = http.NewRequest("POST", "/api/users/register", bytes.NewBuffer(jsonStr3))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, response.Code)

	// Missing field
	var jsonStr4 = []byte(`{"username":"user4"}`)
	req, _ = http.NewRequest("POST", "/api/users/register", bytes.NewBuffer(jsonStr4))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestMethodNotAllowed(t *testing.T) {
	clearTables()

	// GET to /api/users/register
	req, _ := http.NewRequest("GET", "/api/users/register", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusMethodNotAllowed, response.Code)
}

func TestCORS(t *testing.T) {
	clearTables()

	req, _ := http.NewRequest("OPTIONS", "/api/users/register", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	headers := response.HeaderMap
	fmt.Println(len(headers))
	for k, v := range headers {
		fmt.Println(k, "value is", v)
	}

	if headers.Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Errorf("Expected results to be 'http://localhost:3000'. Got '%v'", headers["Access-Control-Allow-Origin"])
	}

	if headers.Get("Access-Control-Allow-Methods") != string("GET, POST, PUT, DELETE, OPTIONS") {
		t.Errorf("Expected results to be 'GET, POST, PUT, DELETE, OPTIONS'. Got '%v'", headers["Access-Control-Allow-Methods"])
	}

	if headers.Get("Access-Control-Allow-Headers") != string("Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, token") {
		t.Errorf("Expected results to be 'Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization'. Got '%v'", headers["Access-Control-Allow-Headers"])
	}

	if headers.Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("Expected results to be 'true'. Got '%v'", headers["Access-Control-Allow-Credentials"])
	}

	var jsonStr1 = []byte(`{"username":"user3@localhost.com", "password": "password3"}`)
	req, _ = http.NewRequest("POST", "/api/users/register", bytes.NewBuffer(jsonStr1))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	headers = response.HeaderMap
	fmt.Println(len(headers))
	for k, v := range headers {
		fmt.Println(k, "value is", v)
	}

	if headers.Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Errorf("Expected results to be 'http://localhost:3000'. Got '%v'", headers["Access-Control-Allow-Origin"])
	}

	if headers.Get("Access-Control-Allow-Methods") != string("GET, POST, PUT, DELETE, OPTIONS") {
		t.Errorf("Expected results to be 'GET, POST, PUT, DELETE, OPTIONS'. Got '%v'", headers["Access-Control-Allow-Methods"])
	}

	if headers.Get("Access-Control-Allow-Headers") != string("Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, token") {
		t.Errorf("Expected results to be 'Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization'. Got '%v'", headers["Access-Control-Allow-Headers"])
	}

}

func authenticate(username, password string) *http.Cookie {
	var jsonStr = []byte(fmt.Sprintf(`{"username":"%s", "password": "%s"}`, username, password))
	req, _ := http.NewRequest("POST", "/api/users/login", bytes.NewBuffer(jsonStr))
	response := executeRequest(req)
	cookie := response.Result().Cookies()[0]
	return cookie
}

func createTestUsers() []string {
	var creds []user
	creds = append(creds, user{Username: "user1@localhost.com", Password: "password1"})
	creds = append(creds, user{Username: "user2@localhost.com", Password: "password2"})

	var userIDs []string

	for _, credential := range creds {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(credential.Password), 8)
		if err != nil {
			log.Fatal(err.Error())
		}

		userID, err := uuid.NewRandom()
		if err != nil {
			log.Fatal(err.Error())
		}
		current := time.Now()
		_, err = testServer.DB.Exec("INSERT INTO users(user_id, username, password, enabled, created, modified) VALUES($1, $2, $3, 1, $4, $5)", userID.String(), credential.Username, hashedPassword, current, current)
		if err != nil {
			log.Fatal(err.Error())
			break
		}
		userIDs = append(userIDs, userID.String())
	}

	return userIDs
}
