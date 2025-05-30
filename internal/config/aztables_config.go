package config

import (
	"errors"
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

	var config AzTableConfig
	//if os.Getenv("AZURE_STORAGE_ACCOUNT_NAME") == "" || os.Getenv("AZURE_STORAGE_ACCOUNT_KEY") == "" || os.Getenv("AZURE_TABLE_SERVICE_URL") == "" {
	//	return nil, errors.New("Missing environment variables for AzTableConfig")
	//}
	//if os.Getenv("AZURE_STORAGE_ACCOUNT_NAME") != "" {
	//	config.AzureAccountName = os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	//}
	//if os.Getenv("AZURE_STORAGE_ACCOUNT_KEY") != "" {
	//	config.AzureAccountKey = os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
	//}
	//if os.Getenv("AZURE_TABLE_SERVICE_URL") != "" {
	//	config.AzureContainerName = os.Getenv("AZURE_TABLE_SERVICE_URL")
	//}

	appEnv := os.Getenv("APP_ENV")

	if appEnv == "development" {
		config.AzureAccountName = os.Getenv("LOCAL_AZURE_STORAGE_ACCOUNT_NAME")
		config.AzureAccountKey = os.Getenv("LOCAL_AZURE_STORAGE_ACCOUNT_KEY")
		config.AzureContainerName = os.Getenv("LOCAL_AZURE_TABLE_SERVICE_URL")
	} else {
		config.AzureAccountName = os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
		config.AzureAccountKey = os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
		config.AzureContainerName = os.Getenv("AZURE_TABLE_SERVICE_URL")
	}

	if config.AzureAccountName == "" || config.AzureAccountKey == "" || config.AzureContainerName == "" {
		return nil, errors.New("Missing environment variables for AzTableConfig")
	}
	return &config, nil
}
