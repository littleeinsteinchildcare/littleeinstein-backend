package routes

import (
	"littleeinsteinchildcare/backend/internal/config"
	"littleeinsteinchildcare/backend/internal/handlers"
	"littleeinsteinchildcare/backend/internal/repositories"
	"littleeinsteinchildcare/backend/internal/services"
	"log"
	"net/http"
)

// SetupRouter configures and returns the main HTTP router for the application.
// It initializes all necessary Azure Table Storage dependencies, registers API routes,
// and configures error handling for the Little Einstein Childcare API.
func SetupRouter() *http.ServeMux {
	// Create a new HTTP router instance
	router := http.NewServeMux()

	// ---------- AZURE TABLE STORAGE CONFIGURATION ----------
	// Load Azure Table Storage configuration (account name, key, container URL)
	// from environment variables or configuration files
	azTableCfg, err := config.LoadAzTableConfig()
	if err != nil {
		log.Fatalf("Router.SetupRouter: Failed to load Azure Table config: %v", err)
	}

	// ---------- USER MODULE SETUP ----------
	// Initialize user repository with Azure Table credentials
	// This creates the shared key credential and service client for Azure Tables
	userRepo, err := repositories.NewUserRepo(*azTableCfg)
	if err != nil {
		log.Fatalf("Router.SetupRouter: Failed to create user repository: %v", err)
	}

	// Create user service with repository dependency
	// This service will handle business logic for user operations
	userService := services.NewUserService(userRepo)

	// Initialize user handler with service dependency
	// This handler will process HTTP requests and use the service layer
	userHandler := handlers.NewUserHandler(userService)

	// Register all user-related routes (create, get, update, delete)
	RegisterUserRoutes(router, userHandler)

	// ---------- EVENT MODULE SETUP ----------
	// Initialize event repository with the same Azure Table configuration
	// Events are stored in a separate table but same storage account
	eventRepo := repositories.NewEventRepo(*azTableCfg)

	// Create event service with repository dependency
	// This service will handle business logic for event operations
	eventService := services.NewEventService(eventRepo)

	// Initialize event handler with event service and user service dependencies
	// The handler needs user service to validate user relationships with events
	eventHandler := handlers.NewEventHandler(eventService, userService)

	// Register all event-related routes (create, get, update, delete)
	RegisterEventRoutes(router, eventHandler)

	// ---------- API INFORMATION ENDPOINT ----------
	// Root endpoint that provides basic API information
	// Acts as a health check and API documentation entry point
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Return 404 if path is not exactly "/"
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		// Return API information as JSON
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Welcome to the Little Einstein Childcare API", "version": "1.0"}`))
	})

	return router
}
