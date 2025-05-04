package main

import (
	"fmt"
	"log"
	"net/http"

	"littleeinsteinchildcare/backend/internal/api/routes"
	"littleeinsteinchildcare/backend/internal/config"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Unable to find .env file")
	}
	fmt.Println("Starting API Server...")

	// Load configuration
	cfg := config.LoadServerConfig()

	// Set up router with all routes
	router := routes.SetupRouter()

	// Server configuration with security timeouts
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
		// Add timeouts later as needed
	}

	log.Printf("API Server running on http://localhost:%d", cfg.Port)

	// Server initialization with fatal error handling
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
