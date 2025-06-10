package main

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	_ "modernc.org/sqlite"

	"github.com/andy6309/go-auth/internal/config"
	"github.com/andy6309/go-auth/internal/handlers"
	"github.com/andy6309/go-auth/internal/middleware"
	"github.com/andy6309/go-auth/internal/models"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create tables if they don't exist
	if err := models.Migrate(db); err != nil {
		logger.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize repositories and handlers
	userRepo := models.NewUserRepository(db)
	authHandler := handlers.NewAuthHandler(userRepo, cfg.JWTSecret, cfg.JWTExpiry)

	// Create router
	router := mux.NewRouter()

	// Configure CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // Allow Next.js frontend
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// API routes
	api := router.PathPrefix("/api").Subrouter()
	api.Use(corsHandler.Handler)
	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/register", authHandler.Register).Methods("POST", "OPTIONS")
	auth.HandleFunc("/login", authHandler.Login).Methods("POST", "OPTIONS")

	// Protected routes
	api.Handle("/profile", middleware.JWTAuth(
		http.HandlerFunc(authHandler.Profile),
		cfg.JWTSecret,
	)).Methods("GET")

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// Serve index.html for all other routes
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	// Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Add CORS middleware to the router
	handler := c.Handler(router)

	// Start server
	addr := ":" + cfg.ServerPort
	logger.Infof("Server starting on port %s", cfg.ServerPort)
	if err := http.ListenAndServe(addr, handler); err != nil {
		logger.Fatalf("Server failed to start: %v", err)
	}
}
