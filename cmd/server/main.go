package main

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"

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
	db.SetMaxOpenConns(1)

	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		logger.Fatalf("Failed to create database tables: %v", err)
	}

	// Initialize repositories and handlers
	userRepo := models.NewUserRepository(db)
	authHandler := handlers.NewAuthHandler(userRepo, cfg.JWTSecret, int(cfg.JWTExpiry.Hours()))

	// Create router
	r := mux.NewRouter()

	// Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug:            true,
	})

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// Auth routes
	authRouter := api.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/register", authHandler.Register).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/login", authHandler.Login).Methods("POST", "OPTIONS")

	// Protected routes
	protectedRouter := api.PathPrefix("").Subrouter()
	protectedRouter.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	protectedRouter.HandleFunc("/profile", authHandler.Profile).Methods("GET")

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// Serve index.html for all other routes
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	// Create HTTP server with CORS middleware
	handler := c.Handler(r)
	
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	logger.Printf("Server starting on port %s", cfg.ServerPort)
	if err := server.ListenAndServe(); err != nil {
		logger.Fatalf("Server failed to start: %v", err)
	}
}

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			created_at DATETIME NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
	`)
	return err
}