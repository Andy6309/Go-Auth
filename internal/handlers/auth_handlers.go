package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/andy6309/go-auth/internal/auth"
	"github.com/andy6309/go-auth/internal/middleware"
	"github.com/andy6309/go-auth/internal/models"
)

type AuthHandler struct {
	userRepo    *models.UserRepository
	jwtSecret   string
	jwtExpiry   time.Duration
}

func NewAuthHandler(userRepo *models.UserRepository, jwtSecret string, jwtExpiry time.Duration) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		jwtSecret:   jwtSecret,
		jwtExpiry:   jwtExpiry,
	}
}

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse represents the response for authentication endpoints
type AuthResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
	TokenType string `json:"token_type"`
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ErrorResponse(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		middleware.ErrorResponse(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	_, err := h.userRepo.GetUserByUsername(req.Username)
	if err == nil {
		middleware.ErrorResponse(w, "Username already exists", http.StatusConflict)
		return
	}

	// Create new user
	user := &models.User{
		Username: req.Username,
		Password: req.Password,
	}

	if err := h.userRepo.CreateUser(user); err != nil {
		middleware.ErrorResponse(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	middleware.JSONResponse(w, map[string]string{"message": "User registered successfully"}, http.StatusCreated)
}

// Login handles user login and returns a JWT token
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ErrorResponse(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		middleware.ErrorResponse(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Get user from database
	user, err := h.userRepo.GetUserByUsername(req.Username)
	if err != nil {
		middleware.ErrorResponse(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Check password
	if err := user.CheckPassword(req.Password); err != nil {
		middleware.ErrorResponse(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	tokenString, err := auth.GenerateToken(user.ID, user.Username, h.jwtSecret, h.jwtExpiry)
	if err != nil {
		middleware.ErrorResponse(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return token in response
	middleware.JSONResponse(w, AuthResponse{
		Token:     tokenString,
		ExpiresIn: int(h.jwtExpiry.Seconds()),
		TokenType: "bearer",
	}, http.StatusOK)
}

// Profile returns the authenticated user's profile
func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		middleware.ErrorResponse(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	username, ok := r.Context().Value(middleware.UsernameKey).(string)
	if !ok {
		middleware.ErrorResponse(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	middleware.JSONResponse(w, map[string]interface{}{
		"id":       userID,
		"username": username,
	}, http.StatusOK)
}
