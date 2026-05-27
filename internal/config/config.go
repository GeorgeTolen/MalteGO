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
	GNAPIPort         string // env GREYNOISE_API_PORT, default "8090"
	UIPort            string // env UI_PORT, default "3000"
	TransformURL      string // env TRANSFORM_SERVICE_URL, default "http://localhost:8080"
	MockMode          bool   // env MOCK_MODE
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

	gnapiPort := os.Getenv("GREYNOISE_API_PORT")
	if gnapiPort == "" {
		gnapiPort = "8090"
	}

	uiPort := os.Getenv("UI_PORT")
	if uiPort == "" {
		uiPort = "3000"
	}

	transformURL := os.Getenv("TRANSFORM_SERVICE_URL")
	if transformURL == "" {
		transformURL = "http://localhost:8080"
	}

	return &Config{
		Port:            port,
		GinMode:         ginMode,
		GreyNoiseAPIKey: apiKey,
		GreyNoiseAPIURL: apiURL,
		RequestTimeout:  time.Duration(timeout) * time.Second,
		GNAPIPort:       gnapiPort,
		UIPort:          uiPort,
		TransformURL:    transformURL,
		MockMode:        os.Getenv("MOCK_MODE") == "true",
	}, nil
}
