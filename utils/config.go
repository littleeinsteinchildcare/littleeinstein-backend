package utils

import (
	"encoding/json"
	"os"
)

// Config structure to hold application configuration
type Config struct {
	Port            string `json:"port"`
	AzureAccountName string `json:"azureAccountName"`
	AzureAccountKey string `json:"azureAccountKey"`
	AzureContainerName string `json:"azureContainerName"`
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(filePath string) (*Config, error) {
	// Open the config file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Parse the config file
	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	// Load from environment variables if available
	if os.Getenv("PORT") != "" {
		config.Port = os.Getenv("PORT")
	}
	if os.Getenv("AZURE_ACCOUNT_NAME") != "" {
		config.AzureAccountName = os.Getenv("AZURE_ACCOUNT_NAME")
	}
	if os.Getenv("AZURE_ACCOUNT_KEY") != "" {
		config.AzureAccountKey = os.Getenv("AZURE_ACCOUNT_KEY")
	}
	if os.Getenv("AZURE_CONTAINER_NAME") != "" {
		config.AzureContainerName = os.Getenv("AZURE_CONTAINER_NAME")
	}

	// Set default port if not specified
	if config.Port == "" {
		config.Port = "8080"
	}

	return &config, nil
}