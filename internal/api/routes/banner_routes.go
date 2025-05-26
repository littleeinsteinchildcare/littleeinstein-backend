package routes

import (
	"littleeinsteinchildcare/backend/internal/handlers"
	"net/http"
)

// Method sets up all banner-related routes
func RegisterBannerRoutes(router *http.ServeMux, bannerHandler *handlers.BannerHandler) {

	router.HandleFunc("GET /api/banner", bannerHandler.GetBanner)
	router.HandleFunc("POST /api/banner", bannerHandler.CreateOrUpdateBanner)
	router.HandleFunc("DELETE /api/banner", bannerHandler.DeleteBanner)
}
