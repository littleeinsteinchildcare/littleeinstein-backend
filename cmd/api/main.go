package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"

	"littleeinsteinchildcare/backend/internal/api/middleware"
	"littleeinsteinchildcare/backend/internal/api/routes"
	"littleeinsteinchildcare/backend/internal/config"

	"littleeinsteinchildcare/backend/firebase"
)

func main() {
	//----Environment setup------

	// Load .env file, ignoring any errors
	_ = godotenv.Load()

	log.Print("\n   ------ VERSION: 2 ------   \n ")

	// Check APP_ENV after potentially loading it from .env
	fmt.Print("App Environment Configuration: ")
	switch environment := os.Getenv("APP_ENV"); environment {
	case "production":
		fmt.Println("Production")
	case "development":
		fmt.Println("Development")
	case "legacy":
		fmt.Println("Legacy")
	default:
		log.Fatal("Error: APP_ENV must be set to either production, development, or legacy")
	}
	fmt.Println("Note: Variables must be configured properly prior to execution")
	fmt.Println("Starting API server...")

	app := firebase.Init()
	// Always sync admin claims from Firestore
	if err := firebase.SyncAdminClaims(app); err != nil {
		log.Fatalf("Error syncing admin claims: %v", err)
	}
	// Load configuration
	cfg, _ := config.LoadServerConfig()

	// Set up router with all routes
	privateRouter := routes.SetupPrivateRouter()
	publicRouter := routes.SetupPublicRouter()

	protected := middleware.FirebaseAuthMiddleware(privateRouter)
	//TODO: Wrap router in Auth middleware
	
	mainRouter := http.NewServeMux()
	mainRouter.Handle("/", publicRouter) // Public routes
	mainRouter.Handle("/api/",protected) // Protected routes



	// Wrap with CORS
	corsHandler := middleware.CorsMiddleware(mainRouter)

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
