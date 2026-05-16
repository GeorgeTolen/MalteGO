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
	return newWithClientFactory(cfg, registry, greynoise.NewClient)
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
	s.router.GET("/", s.handleIndex)
	s.router.POST("/run/:name", s.handleTransform)
	s.router.POST("/run/:name/", s.handleTransform)
}

func (s *Server) handleIndex(c *gin.Context) {
	names := s.registry.Names()
	sort.Strings(names)
	c.JSON(http.StatusOK, gin.H{
		"service":    "MalteGO — GreyNoise Maltego Transform Server",
		"transforms": names,
		"count":      len(names),
	})
}

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
	if apiKey == "" {
		xmlErr, _ := maltego.ErrorResponse("GreyNoise API key not configured. Set GREYNOISE_API_KEY or pass greynoise.api.key in TransformFields.")
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
