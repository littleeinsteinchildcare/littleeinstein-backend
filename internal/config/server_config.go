package config

import (
	"log"
	"os"
	"strconv"
)

// ServerConfig holds application configuration
type ServerConfig struct {
	Port        int
	Host        string
	Environment string
}

// LoadServerConfig returns the application configuration based on environment
func LoadServerConfig() *ServerConfig {
	environment := getEnv("APP_ENV", "development")

	config := &ServerConfig{
		Environment: environment,
		Port:        getPort(environment),
		Host:        getHost(environment),
	}

	log.Printf("Config loaded - Environment: %s, Host: %s, Port: %d",
		config.Environment, config.Host, config.Port)

	return config
}

// getPort determines the appropriate port based on environment
func getPort(environment string) int {
	// Always check for PORT environment variable first (required for Azure)
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		if port, err := strconv.Atoi(portEnv); err == nil {
			return port
		} else {
			log.Printf("Warning: Invalid PORT environment variable '%s'", portEnv)
		}
	}

	// Environment-specific defaults
	switch environment {
	case "production":
		return 8080 // Azure App Service default
	case "development":
		return 3000 // Common development port
	case "legacy":
		return 8080
	default:
		return 8080
	}
}

// getHost determines the appropriate host based on environment
func getHost(environment string) string {
	// Check for custom HOST environment variable
	if host := os.Getenv("HOST"); host != "" {
		return host
	}

	// Environment-specific defaults
	switch environment {
	case "production":
		return "0.0.0.0" // Required for Azure App Service/containers
	case "development":
		return "0.0.0.0" // Use 0.0.0.0 to allow external connections in dev
	case "legacy":
		return "localhost" // Legacy behavior
	default:
		return "0.0.0.0"
	}
}

// Helper function to get environment variable with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
