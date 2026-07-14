package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	GinMode       string
	WorkspaceRoot string
}

func Load() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Failed to load env variables : %v", err)
		return nil, err
	}
	//setting up config
	config := Config{
		Port:          os.Getenv("PORT"),
		GinMode:       os.Getenv("GIN_MODE"),
		WorkspaceRoot: os.Getenv("WORKSPACE_ROOT"),
	}

	return &config, err
}
