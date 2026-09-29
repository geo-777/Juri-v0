package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds the runtime settings used by the application.
type Config struct {
	Port          string
	GinMode       string
	WorkspaceRoot string
	RedisAddress  string
}

// Load reads configuration values from the environment and dotenv file.
func Load() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Failed to load env variables : %v", err)
		return nil, err
	}
	// Build a typed config object from the loaded environment values.
	config := Config{
		Port:          os.Getenv("PORT"),
		GinMode:       os.Getenv("GIN_MODE"),
		WorkspaceRoot: os.Getenv("WORKSPACE_ROOT"),
		RedisAddress:  os.Getenv("REDIS_ADDRESS"),
	}

	return &config, err
}
