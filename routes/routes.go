package routes

import (
	"github.com/gorilla/mux"
	"github.com/littleeinsteinchildcare/beast/controllers"
)

// SetupRoutes configures all routes for the application
func SetupRoutes(r *mux.Router, imageController *controllers.ImageController) {
	// Image routes
	r.HandleFunc("/api/images", imageController.UploadImage).Methods("POST")
	r.HandleFunc("/api/images/{id}/{fileName}", imageController.GetImage).Methods("GET")
}