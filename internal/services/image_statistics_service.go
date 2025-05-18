package services

import (
	"sync"

	"littleeinsteinchildcare/backend/internal/models"
)

// StatisticsService tracks statistics about uploaded images
type StatisticsService struct {
	totalImages   int64
	totalSize     int64
	largestImage  int64
	smallestImage int64
	mutex         sync.RWMutex
	sizeLimit     int64
}

// NewStatisticsService creates a new StatisticsService
func NewStatisticsService(sizeLimit int64) *StatisticsService {
	return &StatisticsService{
		totalImages:   0,
		totalSize:     0,
		largestImage:  0,
		smallestImage: 0,
		sizeLimit:     sizeLimit,
	}
}

// TrackUploadedImage adds a new image to the statistics
func (s *StatisticsService) TrackUploadedImage(size int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.totalImages++
	s.totalSize += size

	// If this is the first image or smaller than the current smallest
	if s.totalImages == 1 || size < s.smallestImage {
		s.smallestImage = size
	}

	// If this is the first image or larger than the current largest
	if s.totalImages == 1 || size > s.largestImage {
		s.largestImage = size
	}
}

// ValidateImageSize checks if an image size is within the allowed limit
func (s *StatisticsService) ValidateImageSize(size int64) *models.SizeValidationResult {
	if size > s.sizeLimit {
		return &models.SizeValidationResult{
			Valid:     false,
			Message:   "Image size exceeds the maximum allowed limit",
			SizeLimit: s.sizeLimit,
			FileSize:  size,
		}
	}

	return &models.SizeValidationResult{
		Valid:     true,
		Message:   "Image size is within the allowed limit",
		SizeLimit: s.sizeLimit,
		FileSize:  size,
	}
}

// GetStatistics returns current image statistics
func (s *StatisticsService) GetStatistics() *models.ImageStatistics {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var avgSize float64 = 0
	if s.totalImages > 0 {
		avgSize = float64(s.totalSize) / float64(s.totalImages)
	}

	return &models.ImageStatistics{
		TotalImages:   s.totalImages,
		TotalSize:     s.totalSize,
		AverageSize:   avgSize,
		LargestImage:  s.largestImage,
		SmallestImage: s.smallestImage,
	}
}

// GetSizeLimit returns the configured maximum image size
func (s *StatisticsService) GetSizeLimit() int64 {
	return s.sizeLimit
}
