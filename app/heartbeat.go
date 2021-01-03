package app

import (
	"fmt"
	"net/http"
)

func (s *Server) handleHeartbeat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf("OK")))
	}
}
