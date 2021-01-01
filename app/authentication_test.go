package app

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestLogin(t *testing.T) {
	clearTables()
	createTestUsers()

	// correct credentials
	var jsonStr1 = []byte(`{"username":"user1", "password": "password1"}`)
	req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonStr1))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// incorrect password
	var jsonStr2 = []byte(`{"username":"user1", "password": "passwordnotcorrect"}`)
	req, _ = http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonStr2))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusUnauthorized, response.Code)

	// incorrect username
	var jsonStr3 = []byte(`{"username":"usernotfound", "password": "password1"}`)
	req, _ = http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonStr3))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestRegister(t *testing.T) {
	// new user
	var jsonStr1 = []byte(`{"username":"user3", "password": "password3"}`)
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonStr1))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	// duplicate user
	var jsonStr2 = []byte(`{"username":"user1", "password": "password1"}`)
	req, _ = http.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonStr2))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func authenticate(username, password string) *http.Cookie {
	var jsonStr = []byte(fmt.Sprintf(`{"username":"%s", "password": "%s"}`, username, password))
	req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonStr))
	response := executeRequest(req)
	cookie := response.Result().Cookies()[0]
	return cookie
}

func createTestUsers() []string {
	var creds []credentials
	creds = append(creds, credentials{Username: "user1", Password: "password1"})
	creds = append(creds, credentials{Username: "user2", Password: "password2"})

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
		testServer.DB.Exec("INSERT INTO users(user_id, username, password, enabled) VALUES($1, $2, $3, 1)", userID.String(), credential.Username, hashedPassword)
		userIDs = append(userIDs, userID.String())
	}

	return userIDs
}
