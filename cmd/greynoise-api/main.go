package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/greynoise-maltego/maltego-go/internal/config"
	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if cfg.GreyNoiseAPIKey == "" {
		log.Fatal("GREYNOISE_API_KEY is required for greynoise-api service")
	}

	port := os.Getenv("GREYNOISE_API_PORT")
	if port == "" {
		port = "8090"
	}

	client := greynoise.NewClient(cfg.GreyNoiseAPIKey, cfg.RequestTimeout)

	gin.SetMode(cfg.GinMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "greynoise-api"})
	})

	r.GET("/community/:ip", func(c *gin.Context) {
		ctx := c.Request.Context()
		resp, err := client.CommunityIP(ctx, c.Param("ip"))
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	r.GET("/context/:ip", func(c *gin.Context) {
		ctx := c.Request.Context()
		resp, err := client.ContextIP(ctx, c.Param("ip"))
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	r.GET("/riot/:ip", func(c *gin.Context) {
		ctx := c.Request.Context()
		resp, err := client.RIOT(ctx, c.Param("ip"))
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	r.GET("/similar/:ip", func(c *gin.Context) {
		ctx := c.Request.Context()
		resp, err := client.SimilarIPs(ctx, c.Param("ip"), 0, 0)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	r.GET("/gnql", func(c *gin.Context) {
		query := c.Query("query")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter required"})
			return
		}
		ctx := c.Request.Context()
		resp, err := client.GNQL(ctx, query, 50)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	log.Printf("GreyNoise API Service starting on :%s", port)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  cfg.RequestTimeout + 5*time.Second,
		WriteTimeout: cfg.RequestTimeout + 5*time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
