package config

import (
	"errors"
	"fmt"
	"log"
	"os"
)

// Config holds application configuration
type AzTableConfig struct {
	AzureAccountName   string
	AzureAccountKey    string
	AzureContainerName string
}

func LoadAzTableConfig() (*AzTableConfig, error) {
	//TODO: Add JSON config/read from file?
	var accEnvName string
	var accEnvKey string
	var accEnvURL string

	switch environment := os.Getenv("APP_ENV"); environment {
	case "production":
		fmt.Println("Production")
		accEnvName = "AZURE_STORAGE_ACCOUNT_NAME"
		accEnvKey = "AZURE_STORAGE_ACCOUNT_KEY"
		accEnvURL = "AZURE_TABLE_SERVICE_URL"
	case "development":
		fmt.Println("Development")
		accEnvName = "LOCAL_AZURE_STORAGE_ACCOUNT_NAME"
		accEnvKey = "LOCAL_AZURE_STORAGE_ACCOUNT_KEY"
		accEnvURL = "LOCAL_AZURE_TABLE_SERVICE_URL"
	case "legacy":
		fmt.Println("Legacy")
		log.Fatal("Error: Legacy environment is no longer functional")
	default:
		log.Fatal("Error: APP_ENV must be set to either production or development")
	}

	var config AzTableConfig
	if os.Getenv(accEnvName) == "" || os.Getenv(accEnvKey) == "" || os.Getenv(accEnvURL) == "" {
		return nil, errors.New("Missing environment variables for AzTableConfig")
	}
	if os.Getenv(accEnvName) != "" {
		config.AzureAccountName = os.Getenv(accEnvName)
	}
	if os.Getenv(accEnvKey) != "" {
		config.AzureAccountKey = os.Getenv(accEnvKey)
	}
	if os.Getenv(accEnvURL) != "" {
		config.AzureContainerName = os.Getenv(accEnvURL)
	}
	return &config, nil
}
