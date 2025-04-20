package config

// Config holds application configuration
type Config struct {
	Port int
}

// Load returns the application configuration
func Load() *Config {
	return &Config{
		Port: 8080,
	}
}
