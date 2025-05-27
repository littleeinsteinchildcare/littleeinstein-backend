package config

import (
	"fmt"
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
func LoadServerConfig() (*ServerConfig, error) {
	environment, err := getEnv("APP_ENV")
	if err != nil {
		return nil, err
	}

	config := &ServerConfig{
		Environment: environment,
		Port:        getPort(environment),
		Host:        getHost(environment),
	}

	log.Printf("Config loaded - Environment: %s, Host: %s, Port: %d",
		config.Environment, config.Host, config.Port)

	return config, nil
}

// getPort determines the appropriate port based on environment
func getPort(environment string) int {
	// Always check for PORT environment variable first (required for Azure)
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		if port, err := strconv.Atoi(portEnv); err == nil { //converse env var that is implicit string to int
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
		return 8080 // Common development port
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
		return "localhost"
	case "legacy":
		return "localhost"
	default:
		return "localhost"
	}
}

// getEnv returns environment variable value or error if not set
func getEnv(key string) (string, error) {
	if value := os.Getenv(key); value != "" {
		return value, nil
	}
	return "", fmt.Errorf("environment variable %s must be set to either production, development, or legacy", key)
}
