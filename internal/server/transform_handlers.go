package server

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

func (s *Server) handleIndex(c *gin.Context) {
	names := s.registry.Names()
	sort.Strings(names)
	c.JSON(http.StatusOK, gin.H{
		"service":    "MalteGO — GreyNoise Transform Service",
		"transforms": names,
		"count":      len(names),
	})
}

func (s *Server) handleAPITransforms(c *gin.Context) {
	names := s.registry.Names()
	sort.Strings(names)
	c.JSON(http.StatusOK, gin.H{"transforms": names})
}

func (s *Server) handleAPIRun(c *gin.Context) {
	name := c.Param("name")
	if _, ok := s.registry.Get(name); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("transform %q not found", name)})
		return
	}

	var body struct {
		Value      string `json:"value"`
		EntityType string `json:"entity_type"`
		APIKey     string `json:"api_key"`
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

	apiKey := body.APIKey
	if apiKey == "" {
		apiKey = s.cfg.GreyNoiseAPIKey
	}
	client := s.newClient(apiKey, s.cfg.RequestTimeout)

	resp, err := s.registry.Run(c.Request.Context(), name, client, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp.ToJSON())
}

// handleBulkRun runs one transform against multiple IPs concurrently.
// POST /api/bulk  {"ips":["1.2.3.4","5.6.7.8"],"transform":"...","api_key":"..."}
func (s *Server) handleBulkRun(c *gin.Context) {
	var body struct {
		IPs       []string `json:"ips"`
		Transform string   `json:"transform"`
		APIKey    string   `json:"api_key"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	if len(body.IPs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ips is required"})
		return
	}
	if len(body.IPs) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "maximum 100 IPs per request"})
		return
	}
	if body.Transform == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transform is required"})
		return
	}
	if _, ok := s.registry.Get(body.Transform); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("transform %q not found", body.Transform)})
		return
	}

	apiKey := body.APIKey
	if apiKey == "" {
		apiKey = s.cfg.GreyNoiseAPIKey
	}

	type result struct {
		IP       string      `json:"ip"`
		Entities interface{} `json:"entities"`
		Messages interface{} `json:"messages"`
		Error    string      `json:"error,omitempty"`
	}

	results := make([]result, len(body.IPs))
	sem := make(chan struct{}, 10) // max 10 concurrent
	var wg sync.WaitGroup

	ctx := c.Request.Context()
	for i, ip := range body.IPs {
		wg.Add(1)
		go func(idx int, ipAddr string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			req := &maltego.Request{
				Value:      ipAddr,
				EntityType: maltego.EntityIPv4Address,
				Properties: map[string]string{},
				Settings:   map[string]string{},
				SoftLimit:  12,
				HardLimit:  12,
			}
			client := s.newClient(apiKey, s.cfg.RequestTimeout)
			resp, err := s.registry.Run(ctx, body.Transform, client, req)
			if err != nil {
				results[idx] = result{IP: ipAddr, Error: err.Error()}
				return
			}
			j := resp.ToJSON()
			results[idx] = result{IP: ipAddr, Entities: j.Entities, Messages: j.Messages}
		}(i, ip)
	}
	wg.Wait()

	c.JSON(http.StatusOK, gin.H{"results": results})
}

// handleTransform handles Maltego TRX XML requests (backward compatible).
func (s *Server) handleTransform(c *gin.Context) {
	name := c.Param("name")

	if _, ok := s.registry.Get(name); !ok {
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
		xmlErr, _ := maltego.ErrorResponse("GreyNoise API key not configured.")
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
