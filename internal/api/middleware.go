package api

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/djengua/mqtt-ingestor/internal/auth"
)

type contextKey string

const userIDKey contextKey = "user_id"

// AuthMiddleware validates JWT token and adds user_id to context
func AuthMiddleware(authService *auth.Service, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				writeError(w, http.StatusUnauthorized, "missing authorization token")
				return
			}

			userID, err := authService.GetUserIDFromToken(token)
			if err != nil {
				logger.Debug("invalid token", slog.String("error", err.Error()))
				writeError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			// Add user_id to context
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// GetUserIDFromContext extracts the user ID from request context
func GetUserIDFromContext(r *http.Request) (int64, bool) {
	userID, ok := r.Context().Value(userIDKey).(int64)
	return userID, ok
}
