package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/andy6309/go-auth/internal/auth"
)

// AuthMiddleware verifies the JWT token in the Authorization header
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip middleware for login and register routes
			if r.URL.Path == "/api/auth/login" || r.URL.Path == "/api/auth/register" {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				ErrorResponse(w, "Authorization header is required", http.StatusUnauthorized)
				return
			}

			// Extract the token from the Authorization header
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
				ErrorResponse(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := tokenParts[1]

			// Validate the token
			claims, err := auth.ValidateToken(tokenString, jwtSecret)
			if err != nil {
				ErrorResponse(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Add user info to the request context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "username", claims.Username)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ErrorResponse is a helper function to send JSON error responses
func ErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// JSONResponse is a helper function to send JSON responses
func JSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}