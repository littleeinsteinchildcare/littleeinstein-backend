package models

type Image struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	ContentType string `json:"contentType"`
	Size        int64  `json:"size"`
	UploadedAt  string `json:"uploadedAt"`
}

type ImageUploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Image   *Image `json:"image,omitempty"`
}