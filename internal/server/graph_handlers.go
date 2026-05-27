package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/greynoise-maltego/maltego-go/internal/storage"
)

func (s *Server) handleListGraphs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	graphs, total, err := s.store.ListGraphs(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if graphs == nil {
		graphs = []storage.Graph{}
	}
	c.JSON(http.StatusOK, gin.H{"graphs": graphs, "total": total, "limit": limit, "offset": offset})
}

func (s *Server) handleSaveGraph(c *gin.Context) {
	var body struct {
		Name string `json:"name" binding:"required"`
		Data string `json:"data" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and data are required"})
		return
	}
	g, err := s.store.SaveGraph(c.Request.Context(), body.Name, body.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, g)
}

func (s *Server) handleGetGraph(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	g, err := s.store.GetGraph(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if g == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "graph not found"})
		return
	}
	c.JSON(http.StatusOK, g)
}

func (s *Server) handleExportGraph(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	g, err := s.store.GetGraph(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if g == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "graph not found"})
		return
	}
	filename := fmt.Sprintf("maltego-graph-%d.json", g.ID)
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(g.Data))
}

func (s *Server) handleImportGraph(c *gin.Context) {
	var body struct {
		Name string `json:"name"`
		Data string `json:"data" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data is required"})
		return
	}
	if body.Name == "" {
		body.Name = "Imported graph"
	}
	g, err := s.store.SaveGraph(c.Request.Context(), body.Name, body.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, g)
}

func (s *Server) handleUpdateGraph(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var body struct {
		Name string `json:"name" binding:"required"`
		Data string `json:"data" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and data are required"})
		return
	}
	g, err := s.store.UpdateGraph(c.Request.Context(), id, body.Name, body.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if g == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "graph not found"})
		return
	}
	c.JSON(http.StatusOK, g)
}

func (s *Server) handleRenameGraph(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var body struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	g, err := s.store.RenameGraph(c.Request.Context(), id, body.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if g == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "graph not found"})
		return
	}
	c.JSON(http.StatusOK, g)
}

func (s *Server) handleDeleteGraph(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := s.store.DeleteGraph(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": id})
}
