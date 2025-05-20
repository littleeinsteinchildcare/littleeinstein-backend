package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"littleeinsteinchildcare/backend/internal/api/middleware"
	"littleeinsteinchildcare/backend/internal/api/routes"
	"littleeinsteinchildcare/backend/internal/config"

	"github.com/joho/godotenv"
)

func main() {
	// Check if .env file exists before trying to load it
	if _, err := os.Stat(".env"); err == nil {
		// File exists, so load it
		if err := godotenv.Load(); err != nil {
			log.Println("Error loading .env file:", err)
		}
	} else {
		log.Println("No .env file found, using environment variables directly")
	}
	fmt.Println("Starting API Server...")

	// Load configuration
	cfg := config.LoadServerConfig()

	// Set up router with all routes
	router := routes.SetupRouter()

	// Wrap with CORS
	corsHandler := middleware.CorsMiddleware(router)

	// Server configuration with security timeouts
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: corsHandler, // CORS (Cross-Origin Resource Sharing) is a browser security feature that blocks requests between different origins (domain/port/protocol).
		// corsHandler wraps our router to add headers that allow our frontend (localhost:3000) to communicate with this backend API
		// Add timeouts later as needed
	}

	log.Printf("API Server running on http://localhost:%d", cfg.Port)

	// Server initialization with fatal error handling
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
