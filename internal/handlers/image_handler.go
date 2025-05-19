package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"littleeinsteinchildcare/backend/internal/models"
	"littleeinsteinchildcare/backend/internal/services"
	"littleeinsteinchildcare/backend/internal/utils"
)

const (
	// Maximum upload size of 10MB
	MaxUploadSize = 10 << 20
)

type BlobService interface {
	UploadImage(ctx context.Context, fileName string, contentType string, data []byte, userID string) (*models.Image, error)
	GetImage(ctx context.Context, userID, fileName string) ([]byte, string, error)
	DeleteImage(ctx context.Context, userID, fileName string) error
}

type StatisticsService interface {
	TrackUploadedImage(size int64)
	ValidateImageSize(size int64) *models.SizeValidationResult
	GetStatistics() *models.ImageStatistics
	GetSizeLimit() int64
}

type ImageHandler struct {
	blobService       BlobService
	statisticsService *services.StatisticsService
}

func NewImageController(blobService BlobService, statisticsService *services.StatisticsService) *ImageHandler {
	return &ImageHandler{
		blobService:       blobService,
		statisticsService: statisticsService,
	}
}

func getUserIDFromAuth(r *http.Request) (string, error) {
	//TODO! - Implement real auth grab (remove r from arguments, pass in context
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		return "", errors.New("Request Header is missing required field: X-User-ID")
	}
	return userID, nil
}

func (h *ImageHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	// Set appropriate headers
	w.Header().Set("Content-Type", "application/json")

	// Parse the multipart form data with a max memory allocation
	err := r.ParseMultipartForm(MaxUploadSize)
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Get the userID from the request
	userID, err := getUserIDFromAuth(r)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Failed to retrieve userID", err)
		return
	}

	// Get the file from the request
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Invalid file upload", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file size using statistics service
	sizeValidation := h.statisticsService.ValidateImageSize(header.Size)
	if !sizeValidation.Valid {
		// Return a structured error response
		response := struct {
			Success   bool                         `json:"success"`
			Message   string                       `json:"message"`
			Violation *models.SizeValidationResult `json:"violation"`
		}{
			Success:   false,
			Message:   "Image size exceeds the maximum allowed limit",
			Violation: sizeValidation,
		}

		jsonResponse, _ := json.Marshal(response)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonResponse)
		return
	}

	// Read the file content
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, file); err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Get the file name and content type
	fileName := filepath.Base(header.Filename)
	contentType := header.Header.Get("Content-Type")

	// Check if the content type is an image
	if !strings.HasPrefix(contentType, "image/") {
		http.Error(w, "File is not an image", http.StatusBadRequest)
		return
	}

	// Upload to Azure Blob Storage
	ctx := context.Background()
	image, err := h.blobService.UploadImage(ctx, fileName, contentType, buffer.Bytes(), userID)
	if err != nil {
		http.Error(w, "Failed to upload image", http.StatusInternalServerError)
		return
	}

	// Track the image in statistics
	h.statisticsService.TrackUploadedImage(header.Size)

	// Return success response
	response := models.ImageUploadResponse{
		Success: true,
		Message: "Image uploaded successfully",
		Image:   image,
	}

	// Marshal the response to JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResponse)
}

func (c *ImageHandler) GetImage(w http.ResponseWriter, r *http.Request) {

	// Get the userID from the request
	userID, err := getUserIDFromAuth(r)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Failed to retrieve userID", err)
		return
	}

	fileName := r.PathValue("fileName")

	if userID == "" || fileName == "" {
		http.Error(w, "Image ID and file name are required", http.StatusBadRequest)
		return
	}

	// Get the image from Azure Blob Storage
	ctx := context.Background()
	data, contentType, err := c.blobService.GetImage(ctx, userID, fileName)
	if err != nil {
		http.Error(w, "Failed to get image", http.StatusNotFound)
		return
	}

	// Set the content type header
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "inline; filename="+fileName)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (h *ImageHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	// Set appropriate headers
	w.Header().Set("Content-Type", "application/json")

	userID, err := getUserIDFromAuth(r)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Failed to retrieve userID", err)
		return
	}

	fileName := r.PathValue("fileName")

	if userID == "" || fileName == "" {
		http.Error(w, "Image ID and file name are required", http.StatusBadRequest)
		return
	}

	// Delete the image from Azure Blob Storage
	ctx := context.Background()
	err2 := h.blobService.DeleteImage(ctx, userID, fileName)
	if err2 != nil {
		http.Error(w, "Failed to delete image", http.StatusNotFound)
		return
	}

	// Return success response
	response := struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}{
		Success: true,
		Message: "Image deleted successfully",
	}

	// Marshal the response to JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func (h *ImageHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	// Set appropriate headers
	w.Header().Set("Content-Type", "application/json")

	// Get statistics
	stats := h.statisticsService.GetStatistics()

	// Return success response
	response := struct {
		Success    bool                    `json:"success"`
		Message    string                  `json:"message"`
		Statistics *models.ImageStatistics `json:"statistics"`
		SizeLimit  int64                   `json:"sizeLimit"`
	}{
		Success:    true,
		Message:    "Statistics retrieved successfully",
		Statistics: stats,
		SizeLimit:  h.statisticsService.GetSizeLimit(),
	}

	// Marshal the response to JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
