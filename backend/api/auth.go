package api

import (
	"backend/middleware"
	"encoding/json"
	"net/http"
	"strings"
)

type LoginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AuthToken string `json:"authorization"`
}

func Login(server *middleware.Server) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var data LoginData

		err := json.NewDecoder(req.Body).Decode(&data)

		if err != nil {
			http.Error(res, "Bad request", http.StatusBadRequest)
			return
		}

		if strings.TrimSpace(data.Username) == "" {
			http.Error(res, "Username cannot be empty", http.StatusBadRequest)
			return
		}

		if data.Password != server.AuthManager.GetPassword() {
			http.Error(res, "Invalid password", http.StatusUnauthorized)
			return
		}

		token := server.AuthManager.Login(data.Username)

		writeJSON(res, http.StatusOK, LoginResponse{AuthToken: token})
	}
}
