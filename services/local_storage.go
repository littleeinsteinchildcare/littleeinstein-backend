package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/littleeinsteinchildcare/beast/models"
)

// LocalStorageService provides an in-memory implementation for blob storage
// Used for local development and testing when Azure credentials are not available
type LocalStorageService struct {
	images map[string][]byte
	meta   map[string]imageMeta
	mutex  sync.RWMutex
}

type imageMeta struct {
	fileName    string
	contentType string
	size        int64
	url         string
	uploadedAt  string
}

// NewLocalStorageService creates a new in-memory storage service
func NewLocalStorageService() *LocalStorageService {
	return &LocalStorageService{
		images: make(map[string][]byte),
		meta:   make(map[string]imageMeta),
	}
}

// UploadImage stores an image in memory
func (s *LocalStorageService) UploadImage(ctx context.Context, fileName string, contentType string, data []byte) (*models.Image, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Generate a unique ID for the image
	imageID := uuid.New().String()

	// Create a unique blob name
	blobName := fmt.Sprintf("%s/%s", imageID, fileName)

	// Mock URL for the uploaded blob
	blobURL := fmt.Sprintf("http://localhost:8080/api/images/%s/%s", imageID, fileName)

	// Store the image data
	s.images[blobName] = data

	// Store metadata
	now := time.Now().Format(time.RFC3339)
	s.meta[blobName] = imageMeta{
		fileName:    fileName,
		contentType: contentType,
		size:        int64(len(data)),
		url:         blobURL,
		uploadedAt:  now,
	}

	// Create and return image info
	image := &models.Image{
		ID:          imageID,
		Name:        fileName,
		URL:         blobURL,
		ContentType: contentType,
		Size:        int64(len(data)),
		UploadedAt:  now,
	}

	return image, nil
}

// GetImage retrieves an image from memory
func (s *LocalStorageService) GetImage(ctx context.Context, imageID, fileName string) ([]byte, string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Construct the blob name from the image ID and file name
	blobName := fmt.Sprintf("%s/%s", imageID, fileName)

	// Check if the image exists
	data, exists := s.images[blobName]
	if !exists {
		return nil, "", errors.New("image not found")
	}

	// Get content type
	meta, exists := s.meta[blobName]
	if !exists {
		return nil, "application/octet-stream", nil
	}

	return data, meta.contentType, nil
}

// DeleteImage removes an image from memory
func (s *LocalStorageService) DeleteImage(ctx context.Context, imageID, fileName string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Construct the blob name from the image ID and file name
	blobName := fmt.Sprintf("%s/%s", imageID, fileName)

	// Check if the image exists
	_, exists := s.images[blobName]
	if !exists {
		return errors.New("image not found")
	}

	// Delete the image and its metadata
	delete(s.images, blobName)
	delete(s.meta, blobName)

	return nil
}