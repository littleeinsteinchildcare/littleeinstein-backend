package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/littleeinsteinchildcare/beast/controllers"
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

	// Initialize the Azure Blob Storage service
	blobService, err := services.NewBlobStorageService(
		config.AzureAccountName,
		config.AzureAccountKey,
		config.AzureContainerName,
	)
	if err != nil {
		log.Fatalf("Failed to initialize Azure Blob Storage service: %v", err)
	}

	// Create controllers
	imageController := controllers.NewImageController(blobService)

	// Create router
	r := mux.NewRouter()

	// Setup routes
	routes.SetupRoutes(r, imageController)

	// Start server
	port := fmt.Sprintf(":%s", config.Port)
	log.Printf("Server starting on port %s", config.Port)
	err = http.ListenAndServe(port, r)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
