package app

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// CheckEnvVariableExists is meant as a fail-safe mechanism for required environment variables. Panics if given environment variable is not present.
func CheckEnvVariableExists(name string) {
	if key := len(os.Getenv(name)); key <= 0 {
		log.Panicf("%s not set", name)
	}
}
