package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	GinMode               string
	GreyNoiseAPIKey       string
	RequestTimeout        time.Duration
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
	if apiKey == "" {
		fmt.Println("[warn] GREYNOISE_API_KEY not set; transforms will use key from request TransformFields")
	}

	return &Config{
		Port:           port,
		GinMode:        ginMode,
		GreyNoiseAPIKey: apiKey,
		RequestTimeout: time.Duration(timeout) * time.Second,
	}, nil
}
