// Package gnapi provides the GreyNoise API proxy HTTP server.
package gnapi

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/server"
)

// Server wraps the GreyNoise API proxy as a proper HTTP server.
type Server struct {
	client   greynoise.Client
	mockMode bool
	router   *gin.Engine
}

// New creates a configured gnapi Server. client may be nil when mockMode is true.
func New(client greynoise.Client, mockMode bool, ginMode string, requestTimeout time.Duration) *Server {
	gin.SetMode(ginMode)
	r := gin.New()
	r.Use(server.SlogMiddleware(), gin.Recovery())

	s := &Server{client: client, mockMode: mockMode, router: r}

	r.GET("/health", s.handleHealth)
	r.GET("/community/:ip", s.handleCommunity)
	r.GET("/context/:ip", s.handleContext)
	r.GET("/riot/:ip", s.handleRIOT)
	r.GET("/similar/:ip", s.handleSimilar)
	r.GET("/gnql", s.handleGNQL)

	return s
}

// HTTPServer returns a configured *http.Server for graceful shutdown.
func (s *Server) HTTPServer(addr string, timeout time.Duration) *http.Server {
	return &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  timeout + 5*time.Second,
		WriteTimeout: timeout + 5*time.Second,
	}
}

func (s *Server) handleHealth(c *gin.Context) {
	mode := "live"
	if s.mockMode {
		mode = "mock"
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "greynoise-api", "mode": mode})
}

func (s *Server) handleCommunity(c *gin.Context) {
	ip := c.Param("ip")
	if s.mockMode {
		c.JSON(http.StatusOK, greynoise.MockCommunity(ip))
		return
	}
	resp, err := s.client.CommunityIP(c.Request.Context(), ip)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) handleContext(c *gin.Context) {
	ip := c.Param("ip")
	if s.mockMode {
		c.JSON(http.StatusOK, greynoise.MockContext(ip))
		return
	}
	resp, err := s.client.ContextIP(c.Request.Context(), ip)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) handleRIOT(c *gin.Context) {
	ip := c.Param("ip")
	if s.mockMode {
		c.JSON(http.StatusOK, greynoise.MockRIOT(ip))
		return
	}
	resp, err := s.client.RIOT(c.Request.Context(), ip)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) handleSimilar(c *gin.Context) {
	ip := c.Param("ip")
	if s.mockMode {
		c.JSON(http.StatusOK, greynoise.MockSimilar(ip))
		return
	}
	resp, err := s.client.SimilarIPs(c.Request.Context(), ip, 0, 0)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) handleGNQL(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter required"})
		return
	}
	if s.mockMode {
		c.JSON(http.StatusOK, greynoise.MockGNQL(query))
		return
	}
	resp, err := s.client.GNQL(c.Request.Context(), query, 50)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
