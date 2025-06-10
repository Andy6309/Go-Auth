package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/andy6309/go-auth/internal/auth"
)

// contextKey is a type for context keys to avoid collisions
type contextKey string

const (
	// UserIDKey is the key used to store the user ID in the context
	UserIDKey contextKey = "user_id"
	// UsernameKey is the key used to store the username in the context
	UsernameKey contextKey = "username"
)

// JWTAuth is a middleware that validates JWT tokens and adds user info to the context
func JWTAuth(next http.Handler, secret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Extract the token from the header (format: "Bearer <token>")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Validate the token and get claims
		claims, err := auth.ValidateToken(tokenString, secret)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add user info to the request context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UsernameKey, claims.Username)

		// Call the next handler with the new context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
