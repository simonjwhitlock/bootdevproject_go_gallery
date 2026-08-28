package httpapi

import (
	"net/http"
	"strings"

	"github.com/simonjwhitlock/bootdevproject_go_gallery/internal/auth"
)

func AuthMiddleware(secret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		claims, err := auth.ValidateToken(parts[1], secret)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		if claims.TokenType != "access" {
			http.Error(w, "Invalid token type", http.StatusUnauthorized)
			return
		}

		// Store user info in context for handlers
		r = r.WithContext(withUserID(r.Context(), claims.UserID))
		next(w, r)
	}
}

type contextKey string

const userIDKey contextKey = "userID"

func withUserID(ctx interface{}, userID interface{}) interface{} {
	// Simple approach - we'll enhance this
	return ctx
}
