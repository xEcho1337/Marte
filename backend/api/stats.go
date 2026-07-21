package api

import (
	"backend/middleware"
	"backend/submitter"
	"net/http"
)

type StatsResponse struct {
	Summary      *FlagSummaryInfo             `json:"summary"`
	Timeline     []submitter.TimelinePoint    `json:"timeline"`
	TopAttackers []submitter.TopAttackerEntry `json:"top_attackers"`
}

func Stats(server *middleware.Server) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		summary := server.FlagManager.GetSummary()
		timeline := server.FlagManager.GetTimeline()
		attackers := server.FlagManager.GetTopAttackers(0)

		writeJSON(res, http.StatusOK, StatsResponse{
			Summary: &FlagSummaryInfo{
				Total:    summary.Total,
				Accepted: summary.Accepted,
				Pending:  summary.Pending,
				Rejected: summary.Rejected,
			},
			Timeline:     timeline,
			TopAttackers: attackers,
		})
	}
}
