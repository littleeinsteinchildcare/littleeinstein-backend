package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/littleeinsteinchildcare/beast/controllers"
	"github.com/littleeinsteinchildcare/beast/middleware"
	"github.com/littleeinsteinchildcare/beast/routes"
	"github.com/littleeinsteinchildcare/beast/services"
	"github.com/littleeinsteinchildcare/beast/utils"
)

func main() {
	// Determine config file path
	configPath := "configs/config.json"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	// Load configuration
	config, err := utils.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Try to initialize the Azure Blob Storage service
	var blobService services.BlobStorageInterface
	azureBlobService, err := services.NewBlobStorageService(
		config.AzureAccountName,
		config.AzureAccountKey,
		config.AzureContainerName,
	)
	
	if err != nil {
		log.Printf("Warning: Failed to initialize Azure Blob Storage service: %v", err)
		log.Printf("Using local in-memory storage instead. This is for testing only.")
		blobService = services.NewLocalStorageService()
	} else {
		blobService = azureBlobService
	}

	// Initialize the Statistics service with a 10MB limit
	statisticsService := services.NewStatisticsService(10 << 20) // 10MB limit

	// Create controllers
	imageController := controllers.NewImageController(blobService, statisticsService)

	// Create router
	r := mux.NewRouter()

	// Add CORS middleware
	corsRouter := middleware.CORS(r)

	// Setup routes
	routes.SetupRoutes(r, imageController)

	// Start server
	port := fmt.Sprintf(":%s", config.Port)
	log.Printf("Server starting on port %s", config.Port)
	err = http.ListenAndServe(port, corsRouter)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
