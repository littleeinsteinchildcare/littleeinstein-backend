package routes

import (
	"net/http"

	"littleeinsteinchildcare/backend/internal/handlers"
)

// SetupRoutes configures all routes for the application
func RegisterBlobRoutes(r *http.ServeMux, imageHandler *handlers.ImageHandler) {
	// Image routes
	// r.HandleFunc("/api/images/{id}/{fileName}", imageController.GetImage).Methods("GET")
	r.HandleFunc("GET /images/{id}/{fileName}", imageHandler.GetImage)

	// r.HandleFunc("/api/images", imageController.UploadImage).Methods("POST")
	r.HandleFunc("POST /images", imageHandler.UploadImage)

	// r.HandleFunc("/api/images/{id}/{fileName}", imageController.DeleteImage).Methods("DELETE")
	r.HandleFunc("DELETE /images/{id}/{fileName}", imageHandler.DeleteImage)

	// Statistics route
	// r.HandleFunc("/api/images/statistics", imageController.GetStatistics).Methods("GET")
	r.HandleFunc("GET /images/statstics", imageHandler.GetStatistics)
}
