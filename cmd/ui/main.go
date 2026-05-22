package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("UI_PORT")
	if port == "" {
		port = "3000"
	}

	transformURL := os.Getenv("TRANSFORM_SERVICE_URL")
	if transformURL == "" {
		transformURL = "http://localhost:8080"
	}

	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "debug"
	}
	gin.SetMode(ginMode)

	target, err := url.Parse(transformURL)
	if err != nil {
		log.Fatalf("invalid TRANSFORM_SERVICE_URL: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Serve static web files
	r.Static("/static", "./web/static")
	r.StaticFile("/", "./web/index.html")

	// Proxy /api/* to transform service
	r.Any("/api/*path", func(c *gin.Context) {
		c.Request.URL.Path = c.Param("path")
		proxy.ServeHTTP(c.Writer, c.Request)
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "ui"})
	})

	log.Printf("UI Service starting on :%s (proxying API to %s)", port, transformURL)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
