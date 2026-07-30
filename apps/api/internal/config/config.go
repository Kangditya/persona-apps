package config

import (
	"errors"
	"os"
)

type Config struct {
	HTTPAddress string
	DatabaseURL string
}

func Load() (Config, error) {
	address := os.Getenv("HTTP_ADDRESS")
	if address == "" {
		address = ":8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	return Config{HTTPAddress: address, DatabaseURL: databaseURL}, nil
}
