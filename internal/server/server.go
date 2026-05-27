package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/greynoise-maltego/maltego-go/internal/config"
	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/storage"
	"github.com/greynoise-maltego/maltego-go/internal/transforms"
)

// Server is the main transform + graph HTTP server.
type Server struct {
	cfg       *config.Config
	registry  *transforms.Registry
	router    *gin.Engine
	newClient func(apiKey string, timeout time.Duration) greynoise.Client
	store     storage.Store
}

func New(cfg *config.Config, registry *transforms.Registry, store storage.Store) *Server {
	factory := greynoise.NewClient
	if cfg.GreyNoiseAPIURL != "" {
		serviceURL := cfg.GreyNoiseAPIURL
		factory = func(_ string, timeout time.Duration) greynoise.Client {
			return greynoise.NewServiceClient(serviceURL, timeout)
		}
	}
	return newWithClientFactory(cfg, registry, factory, store)
}

func newWithClientFactory(cfg *config.Config, registry *transforms.Registry, factory func(string, time.Duration) greynoise.Client, store storage.Store) *Server {
	gin.SetMode(cfg.GinMode)

	r := gin.New()
	r.Use(SlogMiddleware())
	r.Use(gin.Recovery())

	s := &Server{cfg: cfg, registry: registry, router: r, newClient: factory, store: store}
	s.setupRoutes()
	return s
}

// HTTPServer returns a configured *http.Server for use with graceful shutdown.
func (s *Server) HTTPServer() *http.Server {
	return &http.Server{
		Addr:    ":" + s.cfg.Port,
		Handler: s.router,
	}
}

func (s *Server) setupRoutes() {
	// Maltego TRX XML (backward compatible with Maltego Desktop)
	s.router.GET("/", s.handleIndex)
	s.router.POST("/run/:name", s.handleTransform)
	s.router.POST("/run/:name/", s.handleTransform)

	// JSON API for Web UI
	s.router.GET("/api/transforms", s.handleAPITransforms)
	s.router.POST("/api/run/:name", s.handleAPIRun)
	s.router.POST("/api/bulk", s.handleBulkRun)

	// Graph persistence (only when database is configured)
	if s.store != nil {
		s.router.GET("/api/graphs", s.handleListGraphs)
		s.router.POST("/api/graphs", s.handleSaveGraph)
		s.router.POST("/api/graphs/import", s.handleImportGraph)
		s.router.GET("/api/graphs/:id", s.handleGetGraph)
		s.router.GET("/api/graphs/:id/export", s.handleExportGraph)
		s.router.PUT("/api/graphs/:id", s.handleUpdateGraph)
		s.router.PATCH("/api/graphs/:id/rename", s.handleRenameGraph)
		s.router.DELETE("/api/graphs/:id", s.handleDeleteGraph)
	}
}
