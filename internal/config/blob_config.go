package config

import (
	"log"
	"os"
)

// Config structure to hold application configuration
type BlobConfig struct {
	Port               string `json:"port"`
	AzureAccountName   string `json:"azureAccountName"`
	AzureAccountKey    string `json:"azureAccountKey"`
	AzureContainerName string `json:"azureContainerName"`
	AzureServiceURL    string `json:"azureServiceURL"`
}

// LoadConfig loads configuration from a JSON file
func LoadBlobConfig() (*BlobConfig, error) {
	// // Open the config file
	// file, err := os.Open(filePath)
	// if err != nil {
	// 	return nil, err
	// }
	// defer file.Close()

	// // Parse the config file
	// var config BlobConfig
	// decoder := json.NewDecoder(file)
	// err = decoder.Decode(&config)
	// if err != nil {
	// 	return nil, err
	// }
	var accEnvName string
	var accEnvKey string
	var accEnvURL string
	var config BlobConfig

	switch environment := os.Getenv("APP_ENV"); environment {
	case "production":
		accEnvName = "AZURE_STORAGE_ACCOUNT_NAME"
		accEnvKey = "AZURE_STORAGE_ACCOUNT_KEY"
		accEnvURL = "AZURE_BLOB_SERVICE_URL"
	case "development":
		accEnvName = "LOCAL_AZURE_STORAGE_ACCOUNT_NAME"
		accEnvKey = "LOCAL_AZURE_STORAGE_ACCOUNT_KEY"
		accEnvURL = "LOCAL_AZURE_BLOB_SERVICE_URL"
	case "legacy":
		log.Fatal("Error: Legacy environment is no longer functional")
	default:
		log.Fatal("Error: APP_ENV must be set to either production or development")
	}
	// Load from environment variables if available
	if os.Getenv("PORT") != "" {
		config.Port = os.Getenv("PORT")
	}
	if os.Getenv(accEnvName) != "" {
		config.AzureAccountName = os.Getenv(accEnvName)
	}
	if os.Getenv(accEnvKey) != "" {
		config.AzureAccountKey = os.Getenv(accEnvKey)
	}
	if os.Getenv(accEnvURL) != "" {
		config.AzureServiceURL = os.Getenv(accEnvURL)
	}
	if os.Getenv("AZURE_BLOB_CONTAINER_NAME") != "" {
		config.AzureContainerName = os.Getenv("AZURE_BLOB_CONTAINER_NAME")
	}

	// Set default port if not specified
	if config.Port == "" {
		config.Port = "8080"
	}

	return &config, nil
}
