package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	GinMode           string
	GreyNoiseAPIKey   string
	GreyNoiseAPIURL   string // URL of greynoise-api microservice
	RequestTimeout    time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	timeout := 30
	if v := os.Getenv("REQUEST_TIMEOUT_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			timeout = n
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "debug"
	}

	apiKey := os.Getenv("GREYNOISE_API_KEY")
	apiURL := os.Getenv("GREYNOISE_API_URL")

	if apiURL == "" && apiKey == "" {
		fmt.Fprintln(os.Stderr, "[warn] Neither GREYNOISE_API_KEY nor GREYNOISE_API_URL set")
	}

	return &Config{
		Port:            port,
		GinMode:         ginMode,
		GreyNoiseAPIKey: apiKey,
		GreyNoiseAPIURL: apiURL,
		RequestTimeout:  time.Duration(timeout) * time.Second,
	}, nil
}
