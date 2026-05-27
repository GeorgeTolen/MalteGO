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

	mockMode := os.Getenv("MOCK_MODE") == "true"

	port := os.Getenv("GREYNOISE_API_PORT")
	if port == "" {
		port = "8090"
	}

	if !mockMode && cfg.GreyNoiseAPIKey == "" {
		log.Fatal("GREYNOISE_API_KEY is required (or set MOCK_MODE=true for demo)")
	}

	var client greynoise.Client
	if !mockMode {
		client = greynoise.NewClient(cfg.GreyNoiseAPIKey, cfg.RequestTimeout)
	}

	gin.SetMode(cfg.GinMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		mode := "live"
		if mockMode {
			mode = "mock"
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "greynoise-api", "mode": mode})
	})

	r.GET("/community/:ip", func(c *gin.Context) {
		ip := c.Param("ip")
		if mockMode {
			c.JSON(http.StatusOK, greynoise.MockCommunity(ip))
			return
		}
		resp, err := client.CommunityIP(c.Request.Context(), ip)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	r.GET("/context/:ip", func(c *gin.Context) {
		ip := c.Param("ip")
		if mockMode {
			c.JSON(http.StatusOK, greynoise.MockContext(ip))
			return
		}
		resp, err := client.ContextIP(c.Request.Context(), ip)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	r.GET("/riot/:ip", func(c *gin.Context) {
		ip := c.Param("ip")
		if mockMode {
			c.JSON(http.StatusOK, greynoise.MockRIOT(ip))
			return
		}
		resp, err := client.RIOT(c.Request.Context(), ip)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	r.GET("/similar/:ip", func(c *gin.Context) {
		ip := c.Param("ip")
		if mockMode {
			c.JSON(http.StatusOK, greynoise.MockSimilar(ip))
			return
		}
		resp, err := client.SimilarIPs(c.Request.Context(), ip, 0, 0)
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
		if mockMode {
			c.JSON(http.StatusOK, greynoise.MockGNQL(query))
			return
		}
		resp, err := client.GNQL(c.Request.Context(), query, 50)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	if mockMode {
		log.Printf("GreyNoise API Service starting on :%s [MOCK MODE — demo data only]", port)
	} else {
		log.Printf("GreyNoise API Service starting on :%s", port)
	}

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
