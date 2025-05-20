package routes

import (
	"encoding/json"
	"fmt"
	"littleeinsteinchildcare/backend/internal/config"
	"littleeinsteinchildcare/backend/internal/handlers"
	"littleeinsteinchildcare/backend/internal/repositories"
	"littleeinsteinchildcare/backend/internal/services"
	"log"
	"net/http"
	"strings"
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

	// ---------- EVENT MODULE SETUP ----------
	// Initialize event repository with the same Azure Table configuration
	// Events are stored in a separate table but same storage account
	eventRepo, err := repositories.NewEventRepo(*azTableCfg)
	if err != nil {
		log.Fatalf("Router.SetupRouter: Failed to create event repository: %v", err)
	}

	blobConfig, err := config.LoadBlobConfig()
	blobRepo, err := repositories.NewBlobStorageService(blobConfig.AzureAccountName, blobConfig.AzureAccountKey, blobConfig.AzureContainerName)
	if err != nil {
		fmt.Printf("ERROR CREATING BLOB SERVICE: %v\n", err)
		fmt.Printf("ACCOUNT NAME: %s\n", blobConfig.AzureAccountName)
		fmt.Printf("ACCOUNT KEY: %s\n", blobConfig.AzureAccountKey)
		fmt.Printf("ACCOUNT KEY: %s\n", blobConfig.AzureContainerName)
	}
	blobService := services.NewBlobService(blobRepo)
	// Create user service with repository dependency
	// This service will handle business logic for user operations
	userService := services.NewUserService(userRepo, eventRepo, blobRepo)

	// Initialize user handler with service dependency
	// This handler will process HTTP requests and use the service layer
	userHandler := handlers.NewUserHandler(userService)

	// Register all user-related routes (create, get, update, delete)
	RegisterUserRoutes(router, userHandler)

	// Create event service with repository dependency
	// This service will handle business logic for event operations
	eventService := services.NewEventService(eventRepo, *userService)

	// Initialize event handler with event service and user service dependencies
	// The handler needs user service to validate user relationships with events
	eventHandler := handlers.NewEventHandler(eventService, userService)

	// Register all event-related routes (create, get, update, delete)
	RegisterEventRoutes(router, eventHandler)

	// statService := services.StatisticsService{}
	statService := services.NewStatisticsService(handlers.MaxUploadSize)
	imageHandler := handlers.NewImageController(blobService, statService)
	RegisterBlobRoutes(router, imageHandler)

	// Register Azure B2C auth endpoint

	// ---------- BANNER MODULE SETUP ----------
	// Create banner service (no repository needed)
	bannerService := services.NewBannerService()

	// Initialize banner handler with service
	bannerHandler := handlers.NewBannerHandler(bannerService)

	// Register all banner-related routes
	RegisterBannerRoutes(router, bannerHandler)

	// ----------Register Azure B2C Auth Endpoint ----------

	registerAzureB2CEndpoint(router)

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

// registerAzureB2CEndpoint adds the Azure B2C token endpoint
// This is a separate function so it can be easily removed later
func registerAzureB2CEndpoint(router *http.ServeMux) {
	router.HandleFunc("/auth/azure-b2c/{id}", func(w http.ResponseWriter, r *http.Request) {

		// 3. Now check for the actual GET method (frontend is using GET with the callSecureApi function)
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Only log ID for GET requests
		id := r.PathValue("id")
		log.Printf("Received request with ID: %s", id)

		// 4. Extract the token from the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Printf("Error: No Authorization header provided")
			http.Error(w, "No Authorization header provided", http.StatusUnauthorized)
			return
		}

		// 5. The token is expected in the format "Bearer <token>"
		bearerToken := strings.TrimPrefix(authHeader, "Bearer ")
		if bearerToken == authHeader { // If no change, then "Bearer " prefix wasn't there
			log.Printf("Error: Invalid Authorization header format, expected 'Bearer <token>'")
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		log.Printf("Received Azure B2C token: %s\n", bearerToken)

		// For this example, we're just checking if the token exists
		if bearerToken == "" {
			response := struct {
				Status  string `json:"status"`
				Message string `json:"message"`
			}{
				Status:  "error",
				Message: "Invalid token",
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 6. Return a simple successful verification response
		response := struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}{
			Status:  "success",
			Message: "Token successfully verified",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})
}
