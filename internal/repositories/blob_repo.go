package repositories

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"littleeinsteinchildcare/backend/internal/config"
	"littleeinsteinchildcare/backend/internal/models"
	"littleeinsteinchildcare/backend/internal/services"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

type BlobStorageService struct {
	containerURL azblob.ContainerURL
}

func NewBlobStorageService(accountName, accountKey, containerName string) (*BlobStorageService, error) {
	// Create a default request pipeline using your storage account name and account key.
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return nil, err
	}

	// Create a request pipeline that is used to process HTTP(S) requests
	pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	// From the Azure portal, get your storage account blob service URL endpoint.
	var containerURLStr string
	//fmt.Printf("CHECKING APP ENV\n")
	switch os.Getenv("APP_ENV") {
	case "production":

		//fmt.Printf("PRODUCTION\n")
		containerURLStr = fmt.Sprintf("https://%s.blob.core.windows.net/%s", accountName, containerName)
	case "development":
		//fmt.Printf("DEVELOPMENT\n")
		// Use environment variable if set, otherwise default to localhost
		blobServiceURL := os.Getenv("AZURE_BLOB_SERVICE_URL")
		if blobServiceURL == "" {
			blobServiceURL = "http://127.0.0.1:10000"
		}
		containerURLStr = fmt.Sprintf("%s/%s/%s", blobServiceURL, accountName, containerName)
	}

	URL, _ := url.Parse(containerURLStr)

	// Create a ContainerURL object that wraps the container URL and a request
	// pipeline to make requests.
	containerURL := azblob.NewContainerURL(*URL, pipeline)

	// Create the container if it doesn't exist
	ctx := context.Background()
	_, err = containerURL.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
	if err != nil {
		// If the container already exists, continue
		if serr, ok := err.(azblob.StorageError); ok {
			if serr.ServiceCode() == azblob.ServiceCodeContainerAlreadyExists {
				log.Printf("Container '%s' already exists, continuing...", containerName)
				err = nil
			}
		}
		// Also check for HTTP 409 status (Conflict) and error message patterns
		if err != nil {
			errStr := err.Error()
			if serr, ok := err.(azblob.StorageError); ok {
				if serr.Response().StatusCode == 409 {
					log.Printf("Container '%s' already exists (HTTP 409), continuing...", containerName)
					err = nil
				}
			} else if strings.Contains(errStr, "400 Bad Request") || 
					  strings.Contains(errStr, "ContainerAlreadyExists") {
				log.Printf("Container '%s' creation returned 400/already exists error, continuing...", containerName)
				err = nil
			}
		}
		if err != nil {
			return nil, fmt.Errorf("failed to create or access container '%s': %w", containerName, err)
		}
	} else {
		log.Printf("Successfully created container '%s'", containerName)
	}

	return &BlobStorageService{
		containerURL: containerURL,
	}, nil
}

func (s *BlobStorageService) UploadImage(ctx context.Context, fileName string, contentType string, data []byte, userID string) (*models.Image, error) {

	cfg, err := config.LoadAzTableConfig()
	if err != nil {
		return &models.Image{}, fmt.Errorf("BlobRepo.UploadImage: Failed to Load Aztable config %w", err)
	}
	userRepo, err := NewUserRepo(*cfg)
	if err != nil {
		return &models.Image{}, fmt.Errorf("BlobRepo.UploadImage: Failed to initialize User Repo %w", err)
	}

	user, err := userRepo.GetUser(services.USERSTABLE, userID)
	if err != nil {
		return &models.Image{}, fmt.Errorf("BlobRepo.UploadImage: Failed to Get User from UserRepo")
	}
	user.Images = append(user.Images, fileName)

	_, err = userRepo.UpdateUser(services.USERSTABLE, user)
	if err != nil {
		return &models.Image{}, fmt.Errorf("BlobRepo.UploadImage: Failed to add filename")
	}

	// Create a unique blob name
	blobName := fmt.Sprintf("%s/%s", userID, fileName)

	// Get a reference to a blob
	blobURL := s.containerURL.NewBlockBlobURL(blobName)

	// Upload the blob
	uploadOptions := azblob.UploadToBlockBlobOptions{
		BlockSize:   4 * 1024 * 1024,
		Parallelism: 16,
		BlobHTTPHeaders: azblob.BlobHTTPHeaders{
			ContentType: contentType,
		},
		Metadata: azblob.Metadata{
			"id": userID,
		},
	}

	_, err2 := azblob.UploadBufferToBlockBlob(ctx, data, blobURL, uploadOptions)
	if err2 != nil {
		return nil, err
	}

	// Get the URL for the uploaded blob
	blobURLString := blobURL.URL()

	now := time.Now().Format(time.RFC3339)

	// Create and return image info
	image := &models.Image{
		ID:          userID,
		Name:        fileName,
		URL:         blobURLString.String(),
		ContentType: contentType,
		Size:        int64(len(data)),
		UploadedAt:  now,
	}

	return image, nil
}

func (s *BlobStorageService) GetImage(ctx context.Context, userID, fileName string) ([]byte, string, error) {
	// Construct the blob name from the image ID and file name
	blobName := fmt.Sprintf("%s/%s", userID, fileName)

	// Get a reference to the blob
	blobURL := s.containerURL.NewBlockBlobURL(blobName)

	// Download the blob
	downloadResponse, err := blobURL.Download(ctx, 0, azblob.CountToEnd, azblob.BlobAccessConditions{}, false, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		return nil, "", err
	}

	// Read the blob content
	bodyStream := downloadResponse.Body(azblob.RetryReaderOptions{})
	defer bodyStream.Close()

	// Read the entire blob into a buffer
	buffer := new(bytes.Buffer)
	_, err = io.Copy(buffer, bodyStream)
	if err != nil {
		return nil, "", err
	}

	// Get content type
	contentType := downloadResponse.ContentType()

	return buffer.Bytes(), contentType, nil
}

func (s *BlobStorageService) GetAllImages(ctx context.Context) ([]string, error) {
	var imgNames []string

	for marker := (azblob.Marker{}); marker.NotDone(); {
		listBlob, err := s.containerURL.ListBlobsFlatSegment(context.Background(), marker, azblob.ListBlobsSegmentOptions{})
		if err != nil {
			return nil, fmt.Errorf("BlobRepo.GetAllImages: Failed to list blobs: %w", err)
		}

		for _, blob := range listBlob.Segment.BlobItems {
			imgNames = append(imgNames, blob.Name)
		}
		marker = listBlob.NextMarker
	}
	return imgNames, nil
}

func removeImage(images []string, fileName string) []string {
	var removed []string
	removed = append(removed, "")
	for _, img := range images {
		if img != fileName {
			removed = append(removed, img)
		}
	}
	return removed
}
func (s *BlobStorageService) DeleteImage(ctx context.Context, userID, fileName string) error {

	cfg, err := config.LoadAzTableConfig()
	if err != nil {
		return fmt.Errorf("BlobRepo.DeleteImage: Failed to Load Aztable config %w", err)
	}
	userRepo, err := NewUserRepo(*cfg)
	if err != nil {
		return fmt.Errorf("BlobRepo.DeleteImage: Failed to initialize User Repo %w", err)
	}

	user, err := userRepo.GetUser(services.USERSTABLE, userID)
	if err != nil {
		return fmt.Errorf("BlobRepo.DeleteImage: Failed to Get User from UserRepo")
	}

	user.Images = removeImage(user.Images, fileName)

	_, err = userRepo.UpdateUser(services.USERSTABLE, user)
	if err != nil {
		return fmt.Errorf("BlobRepo.DeleteImage: Failed to remove filename")
	}

	// Construct the blob name from the image ID and file name
	blobName := fmt.Sprintf("%s/%s", userID, fileName)

	// Get a reference to the blob
	blobURL := s.containerURL.NewBlockBlobURL(blobName)
	// Delete the blob
	_, err = blobURL.Delete(ctx, azblob.DeleteSnapshotsOptionNone, azblob.BlobAccessConditions{})
	return err
}

func (s *BlobStorageService) DeleteAllImages(userID string) error {
	userImagesFolder := fmt.Sprintf("%s/", userID)
	for marker := (azblob.Marker{}); marker.NotDone(); {
		listBlob, err := s.containerURL.ListBlobsFlatSegment(context.Background(), marker, azblob.ListBlobsSegmentOptions{Prefix: userImagesFolder})
		if err != nil {
			return fmt.Errorf("Failed to list blobs for User %s: %w", userID, err)
		}
		marker = listBlob.NextMarker

		for _, blob := range listBlob.Segment.BlobItems {
			blobURL := s.containerURL.NewBlockBlobURL(blob.Name)
			_, err := blobURL.Delete(context.Background(), azblob.DeleteSnapshotsOptionNone, azblob.BlobAccessConditions{})
			if err != nil {
				log.Printf("Failed to delete blob %s: %v", blob.Name, err)
			}
		}
	}
	return nil
}
