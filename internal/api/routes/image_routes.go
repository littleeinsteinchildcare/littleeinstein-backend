package routes

import (
	"net/http"

	"littleeinsteinchildcare/backend/internal/handlers"
)

// SetupRoutes configures all routes for the application
func RegisterBlobRoutes(r *http.ServeMux, imageHandler *handlers.ImageHandler) {
	// Image routes
	r.HandleFunc("GET /api/image/{id}/{fileName}", imageHandler.GetImage)
	r.HandleFunc("GET /api/images", imageHandler.GetAllImages)

	r.HandleFunc("POST /api/image", imageHandler.UploadImage)

	r.HandleFunc("DELETE /api/image/{id}/{fileName}", imageHandler.DeleteImage)

	// Statistics route
	r.HandleFunc("GET /api/images/statistics", imageHandler.GetStatistics)
}
