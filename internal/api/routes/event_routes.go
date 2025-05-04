package routes

import (
	"net/http"

	"littleeinsteinchildcare/backend/internal/handlers"
)

// RegisterUserRoutes sets up all user-related routes
func RegisterEventRoutes(router *http.ServeMux, eventHandler *handlers.EventHandler) {
	// User routes using Go 1.22+ path pattern syntax
	router.HandleFunc("GET /events/{id}", eventHandler.GetEvent)
	router.HandleFunc("DELETE /events/{id}", eventHandler.DeleteEvent)
	router.HandleFunc("POST /events", eventHandler.CreateEvent)
	// Add other routes as needed
}
