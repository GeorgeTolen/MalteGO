package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/greynoise-maltego/maltego-go/internal/config"
	"github.com/greynoise-maltego/maltego-go/internal/gnapi"
	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/joho/godotenv"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	if !cfg.MockMode && cfg.GreyNoiseAPIKey == "" {
		slog.Error("GREYNOISE_API_KEY is required (or set MOCK_MODE=true for demo)")
		os.Exit(1)
	}

	var client greynoise.Client
	if !cfg.MockMode {
		client = greynoise.NewClient(cfg.GreyNoiseAPIKey, cfg.RequestTimeout)
	}

	srv := gnapi.New(client, cfg.MockMode, cfg.GinMode, cfg.RequestTimeout)
	httpSrv := srv.HTTPServer(":"+cfg.GNAPIPort, cfg.RequestTimeout)

	go func() {
		if cfg.MockMode {
			slog.Info("GreyNoise API Service starting", "port", cfg.GNAPIPort, "mode", "mock")
		} else {
			slog.Info("GreyNoise API Service starting", "port", cfg.GNAPIPort)
		}
		if err := httpSrv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	slog.Info("shutting down gracefully")
	_ = httpSrv.Shutdown(ctx)
}
