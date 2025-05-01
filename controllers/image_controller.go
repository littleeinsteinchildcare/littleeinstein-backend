package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/littleeinsteinchildcare/beast/models"
	"github.com/littleeinsteinchildcare/beast/services"
)

const (
	// Maximum upload size of 10MB
	MaxUploadSize = 10 << 20
)

type ImageController struct {
	blobService       services.BlobStorageInterface
	statisticsService *services.StatisticsService
}

func NewImageController(blobService services.BlobStorageInterface, statisticsService *services.StatisticsService) *ImageController {
	return &ImageController{
		blobService:       blobService,
		statisticsService: statisticsService,
	}
}

func (c *ImageController) UploadImage(w http.ResponseWriter, r *http.Request) {
	// Set appropriate headers
	w.Header().Set("Content-Type", "application/json")

	// Parse the multipart form data with a max memory allocation
	err := r.ParseMultipartForm(MaxUploadSize)
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
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
	sizeValidation := c.statisticsService.ValidateImageSize(header.Size)
	if !sizeValidation.Valid {
		// Return a structured error response
		response := struct {
			Success   bool                       `json:"success"`
			Message   string                     `json:"message"`
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
	image, err := c.blobService.UploadImage(ctx, fileName, contentType, buffer.Bytes())
	if err != nil {
		http.Error(w, "Failed to upload image", http.StatusInternalServerError)
		return
	}

	// Track the image in statistics
	c.statisticsService.TrackUploadedImage(header.Size)

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

func (c *ImageController) GetImage(w http.ResponseWriter, r *http.Request) {
	// Get the image ID and file name from the URL parameters
	vars := mux.Vars(r)
	imageID := vars["id"]
	fileName := vars["fileName"]

	if imageID == "" || fileName == "" {
		http.Error(w, "Image ID and file name are required", http.StatusBadRequest)
		return
	}

	// Get the image from Azure Blob Storage
	ctx := context.Background()
	data, contentType, err := c.blobService.GetImage(ctx, imageID, fileName)
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

func (c *ImageController) DeleteImage(w http.ResponseWriter, r *http.Request) {
	// Set appropriate headers
	w.Header().Set("Content-Type", "application/json")

	// Get the image ID and file name from the URL parameters
	vars := mux.Vars(r)
	imageID := vars["id"]
	fileName := vars["fileName"]

	if imageID == "" || fileName == "" {
		http.Error(w, "Image ID and file name are required", http.StatusBadRequest)
		return
	}

	// Delete the image from Azure Blob Storage
	ctx := context.Background()
	err := c.blobService.DeleteImage(ctx, imageID, fileName)
	if err != nil {
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

func (c *ImageController) GetStatistics(w http.ResponseWriter, r *http.Request) {
	// Set appropriate headers
	w.Header().Set("Content-Type", "application/json")

	// Get statistics
	stats := c.statisticsService.GetStatistics()

	// Return success response
	response := struct {
		Success    bool                  `json:"success"`
		Message    string                `json:"message"`
		Statistics *models.ImageStatistics `json:"statistics"`
		SizeLimit  int64                 `json:"sizeLimit"`
	}{
		Success:    true,
		Message:    "Statistics retrieved successfully",
		Statistics: stats,
		SizeLimit:  c.statisticsService.GetSizeLimit(),
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