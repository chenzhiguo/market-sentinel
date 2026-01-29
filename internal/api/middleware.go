package api

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const tokenContextKey contextKey = "token"

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authentication token")
			return
		}

		if !s.isValidToken(token) {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authentication token")
			return
		}

		ctx := context.WithValue(r.Context(), tokenContextKey, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractToken(r *http.Request) string {
	// Check Authorization header
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// Check query parameter
	if token := r.URL.Query().Get("token"); token != "" {
		return token
	}

	return ""
}

func (s *Server) isValidToken(token string) bool {
	for _, t := range s.cfg.Auth.Tokens {
		if t == token {
			return true
		}
	}
	return false
}
