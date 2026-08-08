package config

import (
	"fmt"
	"os"
	"strconv"
)

const defaultPort = "8080"

type Config struct {
	Address     string
	DatabaseURL string
}

func Load() (Config, error) {
	portValue := os.Getenv("PORT")

	if portValue == "" {
		portValue = defaultPort

	}
	port, err := strconv.Atoi(portValue)
	if err != nil {
		return Config{}, fmt.Errorf(
			"invalid PORT %q, must be a number",
			portValue,
		)
	}

	if port < 1 || port > 65535 {
		return Config{}, fmt.Errorf(

			"invalid PORt %q: must be between 1 and 65535",
			portValue,
		)
	}

	DatabaseURL := os.Getenv("DATABASE_URL")

	if DatabaseURL == "" {
		return Config{}, fmt.Errorf(
			"DATABASE_URL environment variale is required",
		)
	}

	return Config{
		Address:     fmt.Sprintf(":%d", port),
		DatabaseURL: DatabaseURL,
	}, nil
}
