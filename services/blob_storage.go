package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/google/uuid"
	"github.com/littleeinsteinchildcare/beast/models"
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
	URL, _ := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/%s", accountName, containerName))

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
				err = nil
			}
		}
		if err != nil {
			return nil, err
		}
	}

	return &BlobStorageService{
		containerURL: containerURL,
	}, nil
}

func (s *BlobStorageService) UploadImage(ctx context.Context, fileName string, contentType string, data []byte) (*models.Image, error) {
	// Generate a unique ID for the image
	imageID := uuid.New().String()

	// Create a unique blob name
	blobName := fmt.Sprintf("%s/%s", imageID, fileName)

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
			"id": imageID,
		},
	}

	_, err := azblob.UploadBufferToBlockBlob(ctx, data, blobURL, uploadOptions)
	if err != nil {
		return nil, err
	}

	// Get the URL for the uploaded blob
	blobURLString := blobURL.URL()

	now := time.Now().Format(time.RFC3339)

	// Create and return image info
	image := &models.Image{
		ID:          imageID,
		Name:        fileName,
		URL:         blobURLString.String(),
		ContentType: contentType,
		Size:        int64(len(data)),
		UploadedAt:  now,
	}

	return image, nil
}

func (s *BlobStorageService) GetImage(ctx context.Context, imageID, fileName string) ([]byte, string, error) {
	// Construct the blob name from the image ID and file name
	blobName := fmt.Sprintf("%s/%s", imageID, fileName)

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