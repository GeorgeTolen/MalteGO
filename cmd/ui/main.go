package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/greynoise-maltego/maltego-go/internal/config"
	"github.com/greynoise-maltego/maltego-go/internal/server"
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

	target, err := url.Parse(cfg.TransformURL)
	if err != nil {
		slog.Error("invalid TRANSFORM_SERVICE_URL", "err", err)
		os.Exit(1)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)

	gin.SetMode(cfg.GinMode)
	r := gin.New()
	r.Use(server.SlogMiddleware(), gin.Recovery())

	r.Static("/static", "./web/static")
	r.StaticFile("/", "./web/index.html")

	r.Any("/api/*path", func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "ui"})
	})

	httpSrv := &http.Server{
		Addr:    ":" + cfg.UIPort,
		Handler: r,
	}

	go func() {
		slog.Info("UI Service starting", "port", cfg.UIPort, "proxy", cfg.TransformURL)
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
