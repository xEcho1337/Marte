package api

import (
	"backend/middleware"
	"net/http"
)

type ProfileData struct {
	Username string `json:"username"`
}

func Profile(server *middleware.Server) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		token := middleware.ExtractAuthToken(req)
		username := server.AuthManager.GetSession(token)

		writeJSON(res, http.StatusOK, ProfileData{Username: username})
	}
}
