package config

import (
	"errors"
	"os"
	"strconv"
)

type Environment string

const (
	Development Environment = "development"
	Test        Environment = "test"
	Staging     Environment = "staging"
	Production  Environment = "production"
)

type Config struct {
	HTTPAddress              string
	DatabaseURL              string
	Environment              Environment
	AllowDestructiveCommands bool
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

	runtimeEnvironment := Environment(os.Getenv("APP_ENV"))
	if runtimeEnvironment == "" {
		return Config{}, errors.New("APP_ENV is required (development, test, staging, or production)")
	}
	if !isValidEnvironment(runtimeEnvironment) {
		return Config{}, errors.New("APP_ENV must be development, test, staging, or production")
	}

	allowDestructiveCommands := false
	if raw := os.Getenv("ALLOW_DESTRUCTIVE_DB_COMMANDS"); raw != "" {
		var err error
		allowDestructiveCommands, err = strconv.ParseBool(raw)
		if err != nil {
			return Config{}, errors.New("ALLOW_DESTRUCTIVE_DB_COMMANDS must be a boolean")
		}
	}

	return Config{
		HTTPAddress:              address,
		DatabaseURL:              databaseURL,
		Environment:              runtimeEnvironment,
		AllowDestructiveCommands: allowDestructiveCommands,
	}, nil
}

func isValidEnvironment(value Environment) bool {
	switch value {
	case Development, Test, Staging, Production:
		return true
	default:
		return false
	}
}

func (e Environment) AllowsDevelopmentSeeds() bool {
	return e == Development || e == Test
}

func (e Environment) AllowsRollback(explicit bool) bool {
	return e == Development || e == Test || explicit
}
