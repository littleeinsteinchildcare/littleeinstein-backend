package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"littleeinsteinchildcare/backend/internal/api/middleware"
	"littleeinsteinchildcare/backend/internal/api/routes"
	"littleeinsteinchildcare/backend/internal/config"
)

func main() {
	//Environment setup

	fmt.Print("App Environment Configuration: ")
	switch environment := os.Getenv("APP_ENV"); environment {
	case "production":
		fmt.Println("Production")
	case "development":
		fmt.Println("Development")
	default:
		log.Fatal("Error: APP_ENV must be set to either production or development")
	}

	fmt.Println("Note: Variables must be configured properly prior to execution")
	fmt.Println("Starting API server...")

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
