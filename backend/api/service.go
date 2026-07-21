package api

import (
	"backend/middleware"
	"net/http"
)

type ServiceData struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

type ServicesResponse struct {
	Services []ServiceData `json:"services"`
}

func GetService(server *middleware.Server) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var services []ServiceData

		for name, port := range server.Config.Services {
			services = append(services, ServiceData{
				Name: name,
				Port: port,
			})
		}

		resp := ServicesResponse{
			Services: services,
		}

		writeJSON(res, http.StatusOK, resp)
	}
}
