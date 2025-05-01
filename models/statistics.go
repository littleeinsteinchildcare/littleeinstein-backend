package models

// ImageStatistics stores statistics about uploaded images
type ImageStatistics struct {
	TotalImages   int64   `json:"totalImages"`
	TotalSize     int64   `json:"totalSize"`
	AverageSize   float64 `json:"averageSize"`
	LargestImage  int64   `json:"largestImage"`
	SmallestImage int64   `json:"smallestImage"`
}

// SizeValidationResult represents the result of image size validation
type SizeValidationResult struct {
	Valid     bool   `json:"valid"`
	Message   string `json:"message"`
	SizeLimit int64  `json:"sizeLimit"`
	FileSize  int64  `json:"fileSize"`
}