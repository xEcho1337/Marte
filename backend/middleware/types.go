package middleware

import (
	"backend/auth"
	"backend/config"
	"backend/submitter"
	"net/http"
)

type Server struct {
	AuthManager *auth.AuthManager
	FlagManager *submitter.FlagManager
	Config      *config.Config
}

type Middleware func(http.HandlerFunc) http.HandlerFunc
