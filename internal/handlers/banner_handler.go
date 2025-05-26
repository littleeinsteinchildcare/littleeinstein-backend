package handlers

import (
	"encoding/json"
	"fmt"
	"littleeinsteinchildcare/backend/internal/models"
	"littleeinsteinchildcare/backend/internal/utils"
	"net/http"
	"time"
)

// BannerService interface
type BannerService interface {
	GetCurrentBanner() (*models.Banner, error)
	CreateOrUpdateBanner(banner *models.Banner) error
	DeleteBanner() error
	IsTimerRunning() bool
}

// BannerHandler handles HTTP requests related to banners
type BannerHandler struct {
	bannerService BannerService
}

// NewBannerHandler creates a new banner handler
func NewBannerHandler(s BannerService) *BannerHandler {
	return &BannerHandler{
		bannerService: s,
	}
}

// GetBanner handles GET requests to retrieve the current banner
func (h *BannerHandler) GetBanner(w http.ResponseWriter, r *http.Request) {
	banner, err := h.bannerService.GetCurrentBanner()
	if err != nil {
		utils.WriteJSONError(w, http.StatusNotFound, "No active banner found", err)
		return
	}

	response := buildBannerResponse(banner)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateOrUpdateBanner handles POST requests to create or update the current banner
func (h *BannerHandler) CreateOrUpdateBanner(w http.ResponseWriter, r *http.Request) {
	bannerData, err := utils.DecodeJSONRequest(r)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Failed to decode JSON request", err)
		return
	}

	// Validate required fields
	if bannerData["type"] == nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Banner type is required", nil)
		return
	}

	bannerType := bannerData["type"].(string)
	message := ""
	if bannerData["message"] != nil {
		message = bannerData["message"].(string)
	}

	// Validate banner type
	if bannerType != models.BannerTypeWeather && bannerType != models.BannerTypeClosure && bannerType != models.BannerTypeCustom {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid banner type: must be weather, closure, or custom", nil)
		return
	}

	// Validate message is provided for custom type
	if bannerType == models.BannerTypeCustom && (message == "") {
		utils.WriteJSONError(w, http.StatusBadRequest, "Message is required for custom banner type", nil)
		return
	}

	// Parse expiration time
	var expiresAt time.Time
	if bannerData["expiresAt"] == nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Expiration time is required", nil)
		return
	}

	expiresAtStr := bannerData["expiresAt"].(string)
	expiresAt, err = time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid expiration time format, use ISO 8601", err)
		return
	}

	// Validate expiration time is in the future
	if expiresAt.Before(time.Now()) {
		utils.WriteJSONError(w, http.StatusBadRequest, "Expiration time must be in the future", nil)
		return
	}

	// Create a banner object
	banner, err := models.NewBanner(bannerType, message, expiresAt)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("Invalid banner data: %v", err), nil)
		return
	}

	// Save the banner
	err = h.bannerService.CreateOrUpdateBanner(banner)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to save banner", err)
		return
	}

	response := buildBannerResponse(banner)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteBanner handles DELETE requests to remove the current banner
func (h *BannerHandler) DeleteBanner(w http.ResponseWriter, r *http.Request) {
	err := h.bannerService.DeleteBanner()
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to delete banner", err)
		return
	}

	response := map[string]string{
		"message": "banner cleared",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Helper function to build the banner response
func buildBannerResponse(banner *models.Banner) map[string]interface{} {
	return map[string]interface{}{
		"type":      banner.Type,
		"message":   banner.Message,
		"expiresAt": banner.ExpiresAt.Format(time.RFC3339),
	}
}
