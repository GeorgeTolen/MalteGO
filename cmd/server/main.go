package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/greynoise-maltego/maltego-go/internal/app"
	"github.com/greynoise-maltego/maltego-go/internal/config"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
	"github.com/greynoise-maltego/maltego-go/internal/server"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	registry := app.NewRegistry()

	if len(os.Args) > 1 && os.Args[1] == "local" {
		if err := app.RunLocal(cfg, registry, os.Args[2:]); err != nil {
			xmlErr, _ := maltego.ErrorResponse(err.Error())
			fmt.Print(string(xmlErr))
			os.Exit(1)
		}
		return
	}

	srv := server.New(cfg, registry, nil)
	httpSrv := srv.HTTPServer()

	go func() {
		slog.Info("MalteGO starting", "port", cfg.Port)
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
