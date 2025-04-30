package routes

import (
	"net/http"
	"os"

	"littleeinsteinchildcare/backend/internal/handlers"
	"littleeinsteinchildcare/backend/internal/repositories"
	"littleeinsteinchildcare/backend/internal/services"
)

// SetupRouter configures and returns the main router
func SetupRouter() *http.ServeMux {

	router := http.NewServeMux()

	// Initialize repositories, services, and handlers
	userRepoCfg := repositories.NewUserRepoConfig(os.Getenv("AZURE_STORAGE_ACCOUNT_NAME"), os.Getenv("AZURE_STORAGE_ACCOUNT_KEY"), os.Getenv("AZURE_STORAGE_SERVICE_URL"))
	userRepo := repositories.NewUserRepo(*userRepoCfg)
	userService := services.NewUserService(userRepo)

	// Create a handler without actual dependencies for now
	userHandler := handlers.NewUserHandler(userService)

	// Register routes
	RegisterUserRoutes(router, userHandler)

	eventRepoCfg := repositories.NewEventRepoConfig(os.Getenv("AZURE_STORAGE_ACCOUNT_NAME"), os.Getenv("AZURE_STORAGE_ACCOUNT_KEY"), os.Getenv("AZURE_STORAGE_SERVICE_URL"))
	eventRepo := repositories.NewEventRepo(*eventRepoCfg)
	eventService := services.NewEventService(eventRepo)

	eventHandler := handlers.NewEventHandler(eventService, userService)

	RegisterEventRoutes(router, eventHandler)

	// API information endpoint
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Welcome to the Little Einstein Childcare API", "version": "1.0"}`))
	})

	return router
}
