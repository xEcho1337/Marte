package api

import (
	"backend/middleware"
	"net/http"
	"strconv"
)

type FlagInfo struct {
	Value     string `json:"value"`
	Status    string `json:"status"`
	TeamId    int    `json:"team_id"`
	Submitter string `json:"submitter"`
	Timestamp int64  `json:"timestamp"`
}

type FlagSummaryInfo struct {
	Total    int `json:"total"`
	Accepted int `json:"accepted"`
	Pending  int `json:"pending"`
	Rejected int `json:"rejected"`
}

type FlagResponse struct {
	Flags   []FlagInfo       `json:"flags"`
	Total   int              `json:"total"`
	Summary *FlagSummaryInfo `json:"summary,omitempty"`
}

func Flags(server *middleware.Server) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()

		limit := 0
		if limitStr := query.Get("limit"); limitStr != "" {
			var err error
			limit, err = strconv.Atoi(limitStr)
			if err != nil || limit <= 0 {
				limit = 0
			}
		}

		offset := 0
		if offsetStr := query.Get("offset"); offsetStr != "" {
			var err error
			offset, err = strconv.Atoi(offsetStr)
			if err != nil || offset < 0 {
				offset = 0
			}
		}

		entries := server.FlagManager.GetSeenSlice(limit, offset)
		total := server.FlagManager.GetTotalCount()
		summary := server.FlagManager.GetSummary()

		infos := make([]FlagInfo, 0, len(entries))
		for _, e := range entries {
			infos = append(infos, FlagInfo{
				Value:     e.Value,
				Status:    e.Status.String(),
				TeamId:    e.TeamId,
				Submitter: e.Submitter,
				Timestamp: e.Timestamp,
			})
		}

		writeJSON(res, http.StatusOK, FlagResponse{
			Flags: infos,
			Total: total,
			Summary: &FlagSummaryInfo{
				Total:    summary.Total,
				Accepted: summary.Accepted,
				Pending:  summary.Pending,
				Rejected: summary.Rejected,
			},
		})
	}
}
