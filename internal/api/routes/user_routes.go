package routes

import (
	"net/http"

	"littleeinsteinchildcare/backend/internal/handlers"
)

// RegisterUserRoutes sets up all user-related routes
func RegisterUserRoutes(router *http.ServeMux, userHandler *handlers.UserHandler) {
	// User routes using Go 1.22+ path pattern syntax
	router.HandleFunc("GET /users/{id}", userHandler.GetUser)
	router.HandleFunc("POST /users", userHandler.CreateUser)
	// Add other routes as needed
}
