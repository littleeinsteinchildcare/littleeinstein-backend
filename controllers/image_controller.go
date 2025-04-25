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
	blobService *services.BlobStorageService
}

func NewImageController(blobService *services.BlobStorageService) *ImageController {
	return &ImageController{
		blobService: blobService,
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

	// Check file size
	if header.Size > MaxUploadSize {
		http.Error(w, "File too large", http.StatusBadRequest)
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