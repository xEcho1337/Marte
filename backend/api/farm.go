package api

import (
	"backend/middleware"
	"net/http"
)

type FarmResponse struct {
	SubmitterPort int           `json:"submitter_port"`
	NOPTeamID     int           `json:"nop_team_id"`
	TeamID        int           `json:"team_id"`
	TeamIPFormat  string        `json:"team_ip_format"`
	UrlFlagIDs    string        `json:"url_flag_ids"`
	TeamToken     string        `json:"team_token"`
	FlagIDFormat  string        `json:"flag_id_format"`
	FlagRegex     string        `json:"flag_regex"`
	TeamCount     int           `json:"team_count"`
	TeamIPs       []string      `json:"team_ips"`
	Services      []ServiceData `json:"services"`
}

func Farm(server *middleware.Server) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var services []ServiceData
		for name, port := range server.Config.Services {
			services = append(services, ServiceData{
				Name: name,
				Port: port,
			})
		}

		writeJSON(res, http.StatusOK, FarmResponse{
			SubmitterPort: server.Config.SubmitterPort,
			NOPTeamID:     server.Config.NOPTeamId,
			TeamID:        server.Config.TeamId,
			TeamIPFormat:  server.Config.TeamIpFormat,
			UrlFlagIDs:    server.Config.UrlFlagIds,
			TeamToken:     server.Config.TeamToken,
			FlagIDFormat:  server.Config.FlagIdFormat,
			FlagRegex:     server.Config.FlagRegex,
			TeamCount:     server.Config.TeamCount,
			TeamIPs:       server.Config.TeamIPs,
			Services:      services,
		})
	}
}
