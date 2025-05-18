package services

import (
	"context"
	"littleeinsteinchildcare/backend/internal/models"
)

type BlobRepo interface {
	UploadImage(ctx context.Context, fileName string, contentType string, data []byte, userID string) (*models.Image, error)
	GetImage(ctx context.Context, userID, fileName string) ([]byte, string, error)
	DeleteImage(ctx context.Context, userID, fileName string) error
}

type BlobService struct {
	blobRepo BlobRepo
}

func NewBlobService(b BlobRepo) *BlobService {
	return &BlobService{blobRepo: b}
}

func (s *BlobService) UploadImage(ctx context.Context, fileName string, contentType string, data []byte, userID string) (*models.Image, error) {
	img, err := s.blobRepo.UploadImage(ctx, fileName, contentType, data, userID)
	if err != nil {
		return &models.Image{}, err
	}
	return img, nil
}

func (s *BlobService) GetImage(ctx context.Context, userID, fileName string) ([]byte, string, error) {
	data, contentType, err := s.blobRepo.GetImage(ctx, userID, fileName)
	if err != nil {
		return nil, "", err
	}
	return data, contentType, nil
}

func (s *BlobService) DeleteImage(ctx context.Context, userID, fileName string) error {
	err := s.blobRepo.DeleteImage(ctx, userID, fileName)
	if err != nil {
		return err
	}
	return nil
}
