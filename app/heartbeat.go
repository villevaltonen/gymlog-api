package app

import (
	"fmt"
	"net/http"
)

func (s *Server) handleHeartbeat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := validateToken(w, r)
		if err != nil {
			return
		}

		// Finally return welcome to the user along with the username
		w.Write([]byte(fmt.Sprintf("OK")))
	}
}
