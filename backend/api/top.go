package api

import (
	"backend/middleware"
	"backend/submitter"
	"net/http"
	"strconv"
)

type TopResponse struct {
	Attackers []submitter.TopAttackerEntry `json:"attackers"`
}

func Top(server *middleware.Server) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		N := 0
		if nStr := req.URL.Query().Get("n"); nStr != "" {
			if n, err := strconv.Atoi(nStr); err == nil && n > 0 {
				N = n
			}
		}

		attackers := server.FlagManager.GetTopAttackers(N)
		writeJSON(res, http.StatusOK, TopResponse{Attackers: attackers})
	}
}
