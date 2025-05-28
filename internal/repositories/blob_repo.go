package repositories

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
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
	log.Printf("BLOB REPO - NEW BLOB STORAGE SERVICE : CALLED WITH ACCNAME=%s ACCKEY=%s CONTAINERNAME=%s\n", accountName, accountKey, containerName)
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		log.Printf("BLOB REPO - NEW BLOB STORAGE SERVICE: FAILED: ERROR=%v\n", err)
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
		// containerURLStr = fmt.Sprintf("http://127.0.0.1:10000/%s/%s", accountName, containerName)
		containerURLStr = fmt.Sprintf("http://host.docker.internal:10000/%s/%s", accountName, containerName)
	}

	URL, _ := url.Parse(containerURLStr)

	// Create a ContainerURL object that wraps the container URL and a request
	// pipeline to make requests.
	containerURL := azblob.NewContainerURL(*URL, pipeline)

	// Create the container if it doesn't exist
	ctx := context.Background()
	_, err = containerURL.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
	if err != nil {
		log.Printf("BLOB REPO - NEW BLOB STORAGE SERVICE: FAILED: ERROR=%v\n", err)
		// If the container already exists, continue
		if serr, ok := err.(azblob.StorageError); ok {
			if serr.ServiceCode() == azblob.ServiceCodeContainerAlreadyExists {
				err = nil
			}
		}
		if err != nil {
			log.Printf("BLOB REPO - NEW BLOB STORAGE SERVICE: FAILED: ERROR=%v\n", err)
			return nil, err
		}
	}

	log.Printf("BLOB REPO - NEW BLOB STORAGE SERVICE : SUCCESS: containerURL=%v\n", containerURL)
	return &BlobStorageService{
		containerURL: containerURL,
	}, nil
}

func (s *BlobStorageService) UploadImage(ctx context.Context, fileName string, contentType string, data []byte, userID string) (*models.Image, error) {

	cfg, err := config.LoadAzTableConfig()
	if err != nil {
		log.Printf("BLOB REPO - UPLOAD IMAGE - LOAD AZTABLE CONFIG: FAILED: ERROR=%v\n", err)
		return &models.Image{}, fmt.Errorf("BlobRepo.UploadImage: Failed to Load Aztable config %w", err)
	}
	userRepo, err := NewUserRepo(*cfg)
	if err != nil {
		log.Printf("BLOB REPO - UPLOAD IMAGE - NEW USER REPO: FAILED: ERROR=%v\n", err)
		return &models.Image{}, fmt.Errorf("BlobRepo.UploadImage: Failed to initialize User Repo %w", err)
	}

	user, err := userRepo.GetUser(services.USERSTABLE, userID)
	if err != nil {
		log.Printf("BLOB REPO - UPLOAD IMAGE - USER REPO.GET USER FAILED: ERROR=%v\n", err)
		return &models.Image{}, fmt.Errorf("BlobRepo.UploadImage: Failed to Get User from UserRepo")
	}
	user.Images = append(user.Images, fileName)

	_, err = userRepo.UpdateUser(services.USERSTABLE, user)
	if err != nil {
		log.Printf("BLOB REPO - UPLOAD IMAGE - USER REPO.UPDATE USER FAILED: ERROR=%v\n", err)
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
		log.Printf("BLOB REPO - UPLOAD IMAGE - UPLOAD BLOB FAILED: ERROR=%v\n", err2)
		return nil, err2
	}

	// Get the URL for the uploaded blob
	blobURLString := blobURL.URL()

	now := time.Now().Format(time.RFC3339)

	log.Printf("BLOB REPO - UPLOAD IMAGE - PACKAGING IMAGE: USERID=%s FILENAME=%s URLSTRING=%s\n", userID, fileName, blobURLString.String())
	// Create and return image info
	image := &models.Image{
		ID:          userID,
		Name:        fileName,
		URL:         blobURLString.String(),
		ContentType: contentType,
		Size:        int64(len(data)),
		UploadedAt:  now,
	}
	log.Printf("BLOB REPO - UPLOAD IMAGE: SUCCESS: IMAGE=%v\n", image)

	return image, nil
}

func (s *BlobStorageService) GetImage(ctx context.Context, userID, fileName string) ([]byte, string, error) {
	log.Printf("BLOB REPO - GET IMAGE: CALLED WITH userID=%s filename=%s\n", userID, fileName)
	// Construct the blob name from the image ID and file name
	blobName := fmt.Sprintf("%s/%s", userID, fileName)

	// Get a reference to the blob
	blobURL := s.containerURL.NewBlockBlobURL(blobName)

	// Download the blob
	downloadResponse, err := blobURL.Download(ctx, 0, azblob.CountToEnd, azblob.BlobAccessConditions{}, false, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		log.Printf("BLOB REPO - GET IMAGE - blobURL.Download : FAILED: ERROR=%v\n", err)
		return nil, "", err
	}

	// Read the blob content
	bodyStream := downloadResponse.Body(azblob.RetryReaderOptions{})
	defer bodyStream.Close()

	// Read the entire blob into a buffer
	buffer := new(bytes.Buffer)
	_, err = io.Copy(buffer, bodyStream)
	if err != nil {
		log.Printf("BLOB REPO - GET IMAGE - io.Copy : FAILED: ERROR=%v\n", err)
		return nil, "", err
	}

	// Get content type
	contentType := downloadResponse.ContentType()

	log.Printf("BLOB REPO - GET IMAGE: SUCCESS\n")
	return buffer.Bytes(), contentType, nil
}

func (s *BlobStorageService) GetAllImages(ctx context.Context) ([]string, error) {
	log.Printf("BLOB REPO - GET ALL IMAGES: CALLED\n")
	var imgNames []string

	for marker := (azblob.Marker{}); marker.NotDone(); {
		listBlob, err := s.containerURL.ListBlobsFlatSegment(context.Background(), marker, azblob.ListBlobsSegmentOptions{})
		if err != nil {
			log.Printf("BLOB REPO - GET ALL IMAGES: FAILED: ERROR=%v\n", err)
			return nil, fmt.Errorf("BlobRepo.GetAllImages: Failed to list blobs: %w", err)
		}

		for _, blob := range listBlob.Segment.BlobItems {
			imgNames = append(imgNames, blob.Name)
		}
		marker = listBlob.NextMarker
	}
	log.Printf("BLOB REPO - GET ALL IMAGES: SUCCESS: imgNames=%v\n", imgNames)
	return imgNames, nil
}

func removeImage(images []string, fileName string) []string {
	log.Printf("BLOB REPO - REMOVE IMAGE: CALLED\n")
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
	log.Printf("BLOB REPO - DELETE IMAGE: CALLED WITH USERID=%s FILENAME=%s\n", userID, fileName)

	cfg, err := config.LoadAzTableConfig()
	if err != nil {
		log.Printf("BLOB REPO - DELETE IMAGE - LOAD AZTABLE CONFIG: FAILED: ERROR=%v\n", err)
		return fmt.Errorf("BlobRepo.DeleteImage: Failed to Load Aztable config %w", err)
	}
	userRepo, err := NewUserRepo(*cfg)
	if err != nil {
		log.Printf("BLOB REPO - DELETE IMAGE - USER REPO.NEW USER REPO FAILED: ERROR=%v\n", err)
		return fmt.Errorf("BlobRepo.DeleteImage: Failed to initialize User Repo %w", err)
	}

	user, err := userRepo.GetUser(services.USERSTABLE, userID)
	if err != nil {
		log.Printf("BLOB REPO - DELETE IMAGE - USER REPO.GET USER FAILED: ERROR=%v\n", err)
		return fmt.Errorf("BlobRepo.DeleteImage: Failed to Get User from UserRepo")
	}

	user.Images = removeImage(user.Images, fileName)

	_, err = userRepo.UpdateUser(services.USERSTABLE, user)
	if err != nil {
		log.Printf("BLOB REPO - DELETE IMAGE - USER REPO.UPDATE USER FAILED: ERROR=%v\n", err)
		return fmt.Errorf("BlobRepo.DeleteImage: Failed to remove filename")
	}

	// Construct the blob name from the image ID and file name
	blobName := fmt.Sprintf("%s/%s", userID, fileName)

	// Get a reference to the blob
	blobURL := s.containerURL.NewBlockBlobURL(blobName)
	// Delete the blob
	_, err = blobURL.Delete(ctx, azblob.DeleteSnapshotsOptionNone, azblob.BlobAccessConditions{})
	log.Printf("BLOB REPO - DELETE IMAGE: SUCCESS\n")
	return err
}

func (s *BlobStorageService) DeleteAllImages(userID string) error {
	log.Printf("BLOB REPO - DELETE ALL IMAGES: CALLED USERID=%s \n", userID)
	userImagesFolder := fmt.Sprintf("%s/", userID)
	for marker := (azblob.Marker{}); marker.NotDone(); {
		listBlob, err := s.containerURL.ListBlobsFlatSegment(context.Background(), marker, azblob.ListBlobsSegmentOptions{Prefix: userImagesFolder})
		if err != nil {
			log.Printf("BLOB REPO - DELETE ALL IMAGES - FAILED: ERROR=%v\n", err)
			return fmt.Errorf("Failed to list blobs for User %s: %w", userID, err)
		}
		marker = listBlob.NextMarker

		for _, blob := range listBlob.Segment.BlobItems {
			blobURL := s.containerURL.NewBlockBlobURL(blob.Name)
			_, err := blobURL.Delete(context.Background(), azblob.DeleteSnapshotsOptionNone, azblob.BlobAccessConditions{})
			if err != nil {
				log.Printf("BLOB REPO - DELETE ALL IMAGES - FAILED: ERROR=%v\n", err)
				log.Printf("Failed to delete blob %s: %v", blob.Name, err)
			}
		}
	}
	return nil
}
