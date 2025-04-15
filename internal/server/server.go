package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/royceleond/scansca/internal/mcp"
)

// Config holds server configuration
type Config struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// DefaultConfig returns default server configuration
func DefaultConfig() Config {
	return Config{
		Host:            "0.0.0.0",
		Port:            8080,
		ReadTimeout:     time.Second * 15,
		WriteTimeout:    time.Second * 15,
		ShutdownTimeout: time.Second * 5,
	}
}

// Server represents the HTTP server for the application
type Server struct {
	config     Config
	router     *gin.Engine
	httpServer *http.Server
	connMgr    *ConnectorManager
	toolRegistry *mcp.ToolRegistry
}

// New creates a new server instance
func New(config Config, connMgr *ConnectorManager) *Server {
	router := gin.Default()
	
	server := &Server{
		config:     config,
		router:     router,
		connMgr:    connMgr,
	}
	
	// Create tool registry
	server.toolRegistry = mcp.NewToolRegistry(connMgr)
	
	return server
}

// Initialize sets up routes and middleware
func (s *Server) Initialize() {
	// Setup global middleware
	s.router.Use(
		gin.Recovery(),
		corsMiddleware(),
		requestLogger(),
	)

	// API routes
	api := s.router.Group("/api/v1")
	{
		// Database management
		api.POST("/databases", s.registerDatabase)
		api.GET("/databases", s.listDatabases)
		api.GET("/databases/:name", s.getDatabaseInfo)
		api.DELETE("/databases/:name", s.removeDatabase)

		// Schema introspection
		api.GET("/databases/:name/schemas", s.listSchemas)
		api.GET("/databases/:name/schemas/:schema/tables", s.listTables)
		api.GET("/databases/:name/schemas/:schema/tables/:table", s.getTableInfo)

		// Query execution
		api.POST("/query", s.executeQuery)
	}
	
	// MCP routes
	mcpAPI := s.router.Group("/mcp/v1")
	{
		// Tool listing
		mcpAPI.GET("/tools", s.listTools)
		
		// Tool execution
		mcpAPI.POST("/tools/:name/invoke", s.invokeTool)
	}

	// Health check
	s.router.GET("/health", s.healthCheck)
}

// Start begins listening for requests
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	log.Info().Str("address", addr).Msg("Starting HTTP server")
	
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.ShutdownTimeout)
	defer cancel()
	
	log.Info().Msg("Shutting down HTTP server")
	return s.httpServer.Shutdown(shutdownCtx)
}

// SetPort updates the server port
func (s *Server) SetPort(port int) {
	s.config.Port = port
}

// healthCheck handles health check requests
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": "1.0.0", // Should be pulled from app version
		"time":    time.Now().UTC(),
	})
}

// MCP handlers

// listTools returns all available MCP tools
func (s *Server) listTools(c *gin.Context) {
	tools := s.toolRegistry.ListTools()
	
	// Remove handlers from the response
	response := make([]map[string]interface{}, len(tools))
	for i, tool := range tools {
		response[i] = map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"parameters":  tool.Parameters,
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"tools": response,
	})
}

// invokeTool handles tool execution
func (s *Server) invokeTool(c *gin.Context) {
	toolName := c.Param("name")
	
	// Check if tool exists
	_, exists := s.toolRegistry.GetTool(toolName)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Tool not found: %s", toolName),
		})
		return
	}
	
	// Read raw JSON to pass to the tool handler
	rawJSON, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Failed to read request body: %v", err),
		})
		return
	}
	
	// Set timeout based on the tool
	timeout := 30 * time.Second
	if toolName == "execute_query" {
		timeout = 60 * time.Second
	}
	
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()
	
	// Execute the tool
	result, err := s.toolRegistry.ExecuteTool(ctx, toolName, rawJSON)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  err.Error(),
			"status": "error",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"result": result,
	})
}