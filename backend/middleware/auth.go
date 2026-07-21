package middleware

import (
	"net/http"
	"strings"
)

func RequireAuth(srv *Server) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(res http.ResponseWriter, req *http.Request) {
			token := ExtractAuthToken(req)

			if token == "" {
				http.Error(res, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			if srv.AuthManager.GetSession(token) == "" {
				http.Error(res, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			next(res, req)
		}
	}
}

func ExtractAuthToken(req *http.Request) string {
	header := req.Header.Get("Authorization")
	token := strings.TrimPrefix(header, "Bearer ")
	return token
}
