package routes

import (
	"net/http"

	"littleeinsteinchildcare/backend/internal/handlers"
)

// RegisterUserRoutes sets up all user-related routes
func RegisterUserRoutes(router *http.ServeMux, userHandler *handlers.UserHandler) {
	// User routes using Go 1.22+ path pattern syntax
	router.HandleFunc("GET /api/user/{id}", userHandler.GetUser)
	router.HandleFunc("GET /api/users", userHandler.GetAllUsers)
	router.HandleFunc("PUT /api/user/{id}", userHandler.UpdateUser)
	router.HandleFunc("DELETE /api/user/{id}", userHandler.DeleteUser)
	router.HandleFunc("POST /api/user", userHandler.CreateUser)

}
