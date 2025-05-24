package routes

import (
	"net/http"

	"littleeinsteinchildcare/backend/internal/handlers"
	"littleeinsteinchildcare/backend/internal/api/middleware"
)

// RegisterEventRoutes sets up all event-related routes with Firebase authentication
func RegisterEventRoutes(router *http.ServeMux, eventHandler *handlers.EventHandler) {
	// Event routes with Firebase authentication middleware
	router.Handle("GET /api/event/{id}", middleware.FirebaseAuthMiddleware(http.HandlerFunc(eventHandler.GetEvent)))
	router.Handle("GET /api/events", middleware.FirebaseAuthMiddleware(http.HandlerFunc(eventHandler.GetAllEvents)))
	router.Handle("GET /api/events/user/{userId}", middleware.FirebaseAuthMiddleware(http.HandlerFunc(eventHandler.GetEventsByUser)))
	router.Handle("DELETE /api/event/{id}", middleware.FirebaseAuthMiddleware(http.HandlerFunc(eventHandler.DeleteEvent)))
	router.Handle("POST /api/event", middleware.FirebaseAuthMiddleware(http.HandlerFunc(eventHandler.CreateEvent)))
	router.Handle("PUT /api/event/{id}", middleware.FirebaseAuthMiddleware(http.HandlerFunc(eventHandler.UpdateEvent)))
	router.Handle("GET /test", middleware.FirebaseAuthMiddleware(http.HandlerFunc(eventHandler.TestConnection)))
}
