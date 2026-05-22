package server

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/greynoise-maltego/maltego-go/internal/config"
	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
	"github.com/greynoise-maltego/maltego-go/internal/transforms"
)

type Server struct {
	cfg       *config.Config
	registry  *transforms.Registry
	router    *gin.Engine
	newClient func(apiKey string, timeout time.Duration) greynoise.Client
}

func New(cfg *config.Config, registry *transforms.Registry) *Server {
	factory := greynoise.NewClient
	// If GREYNOISE_API_URL is set, route through the greynoise-api microservice.
	if cfg.GreyNoiseAPIURL != "" {
		serviceURL := cfg.GreyNoiseAPIURL
		factory = func(_ string, timeout time.Duration) greynoise.Client {
			return greynoise.NewServiceClient(serviceURL, timeout)
		}
	}
	return newWithClientFactory(cfg, registry, factory)
}

func newWithClientFactory(cfg *config.Config, registry *transforms.Registry, factory func(string, time.Duration) greynoise.Client) *Server {
	gin.SetMode(cfg.GinMode)

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	s := &Server{cfg: cfg, registry: registry, router: r, newClient: factory}
	s.setupRoutes()
	return s
}

func (s *Server) Run() error {
	return s.router.Run(":" + s.cfg.Port)
}

func (s *Server) setupRoutes() {
	// Maltego TRX XML endpoints (backward compatible)
	s.router.GET("/", s.handleIndex)
	s.router.POST("/run/:name", s.handleTransform)
	s.router.POST("/run/:name/", s.handleTransform)

	// JSON API for Web UI
	s.router.GET("/api/transforms", s.handleAPITransforms)
	s.router.POST("/api/run/:name", s.handleAPIRun)
}

func (s *Server) handleIndex(c *gin.Context) {
	names := s.registry.Names()
	sort.Strings(names)
	c.JSON(http.StatusOK, gin.H{
		"service":    "MalteGO — GreyNoise Transform Service",
		"transforms": names,
		"count":      len(names),
	})
}

// handleAPITransforms returns list of transforms as JSON for the Web UI.
func (s *Server) handleAPITransforms(c *gin.Context) {
	names := s.registry.Names()
	sort.Strings(names)
	c.JSON(http.StatusOK, gin.H{"transforms": names})
}

// handleAPIRun runs a transform and returns JSON (used by Web UI).
func (s *Server) handleAPIRun(c *gin.Context) {
	name := c.Param("name")

	if _, ok := s.registry.Get(name); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("transform %q not found", name)})
		return
	}

	var body struct {
		Value      string `json:"value"`
		EntityType string `json:"entity_type"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	if strings.TrimSpace(body.Value) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "value is required"})
		return
	}
	if body.EntityType == "" {
		body.EntityType = maltego.EntityIPv4Address
	}

	req := &maltego.Request{
		Value:      body.Value,
		EntityType: body.EntityType,
		Properties: map[string]string{},
		Settings:   map[string]string{},
		SoftLimit:  12,
		HardLimit:  12,
	}

	client := s.newClient(s.cfg.GreyNoiseAPIKey, s.cfg.RequestTimeout)

	resp, err := s.registry.Run(c.Request.Context(), name, client, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp.ToJSON())
}

// handleTransform handles Maltego TRX XML requests.
func (s *Server) handleTransform(c *gin.Context) {
	name := c.Param("name")

	_, ok := s.registry.Get(name)
	if !ok {
		xmlErr, _ := maltego.ErrorResponse(fmt.Sprintf("Transform %q not found", name))
		c.Data(http.StatusOK, "text/xml; charset=utf-8", xmlErr)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		xmlErr, _ := maltego.ErrorResponse("Failed to read request body")
		c.Data(http.StatusOK, "text/xml; charset=utf-8", xmlErr)
		return
	}

	req, err := maltego.ParseRequest(body)
	if err != nil {
		xmlErr, _ := maltego.ErrorResponse(fmt.Sprintf("Invalid Maltego XML: %v", err))
		c.Data(http.StatusOK, "text/xml; charset=utf-8", xmlErr)
		return
	}

	apiKey := req.APIKey(s.cfg.GreyNoiseAPIKey)
	if apiKey == "" && s.cfg.GreyNoiseAPIURL == "" {
		xmlErr, _ := maltego.ErrorResponse("GreyNoise API key not configured. Set GREYNOISE_API_KEY or GREYNOISE_API_URL.")
		c.Data(http.StatusOK, "text/xml; charset=utf-8", xmlErr)
		return
	}

	if strings.TrimSpace(req.Value) == "" {
		xmlErr, _ := maltego.ErrorResponse("Entity value is empty")
		c.Data(http.StatusOK, "text/xml; charset=utf-8", xmlErr)
		return
	}

	client := s.newClient(apiKey, s.cfg.RequestTimeout)

	resp, err := s.registry.Run(c.Request.Context(), name, client, req)
	if err != nil {
		xmlErr, _ := maltego.ErrorResponse(fmt.Sprintf("Transform error: %v", err))
		c.Data(http.StatusOK, "text/xml; charset=utf-8", xmlErr)
		return
	}

	xmlData, err := resp.ToXML()
	if err != nil {
		xmlErr, _ := maltego.ErrorResponse(fmt.Sprintf("XML serialisation error: %v", err))
		c.Data(http.StatusOK, "text/xml; charset=utf-8", xmlErr)
		return
	}

	c.Data(http.StatusOK, "text/xml; charset=utf-8", xmlData)
}
