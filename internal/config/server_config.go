package config

// Config holds application configuration
type ServerConfig struct {
	Port int
}

// Load returns the application configuration
func LoadServerConfig() *ServerConfig {
	return &ServerConfig{
		Port: 8080,
	}
}
