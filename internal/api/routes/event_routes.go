package routes

import (
	"net/http"

	"littleeinsteinchildcare/backend/internal/handlers"
)

// RegisterEventRoutes sets up all event-related routes
func RegisterEventRoutes(router *http.ServeMux, eventHandler *handlers.EventHandler) {
	// Event routes - authentication handled at router level
	router.Handle("GET /api/event/{id}", http.HandlerFunc(eventHandler.GetEvent))
	router.Handle("GET /api/events", http.HandlerFunc(eventHandler.GetAllEvents))
	router.Handle("GET /api/events/user/{userId}", http.HandlerFunc(eventHandler.GetEventsByUser))
	router.Handle("DELETE /api/event/{id}", http.HandlerFunc(eventHandler.DeleteEvent))
	router.Handle("POST /api/event", http.HandlerFunc(eventHandler.CreateEvent))
	router.Handle("PUT /api/event/{id}", http.HandlerFunc(eventHandler.UpdateEvent))
	router.Handle("GET /test", http.HandlerFunc(eventHandler.TestConnection))
}
