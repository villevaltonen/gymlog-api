package app

import (
	"net/http"
	"testing"
)

func TestHeartbeat(t *testing.T) {
	// without JWT
	req1, _ := http.NewRequest("GET", "/api/heartbeat", nil)
	response1 := executeRequest(req1)
	checkResponseCode(t, http.StatusUnauthorized, response1.Code)

	// with JWT
	req3, _ := http.NewRequest("GET", "/api/heartbeat", nil)
	req3.AddCookie(authenticate())
	response3 := executeRequest(req3)
	checkResponseCode(t, http.StatusOK, response3.Code)
}