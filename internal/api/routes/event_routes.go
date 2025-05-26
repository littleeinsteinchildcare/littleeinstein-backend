package routes

import (
	"net/http"

	"littleeinsteinchildcare/backend/internal/handlers"
)

// RegisterUserRoutes sets up all user-related routes
func RegisterEventRoutes(router *http.ServeMux, eventHandler *handlers.EventHandler) {
	// User routes using Go 1.22+ path pattern syntax
	router.HandleFunc("GET /api/event/{id}", eventHandler.GetEvent)
	router.HandleFunc("GET /api/events", eventHandler.GetAllEvents)
	router.HandleFunc("DELETE /api/event/{id}", eventHandler.DeleteEvent)
	router.HandleFunc("POST /api/event", eventHandler.CreateEvent)
	router.HandleFunc("PUT /api/event/{id}", eventHandler.UpdateEvent)
	router.HandleFunc("GET /test", eventHandler.TestConnection)

}
