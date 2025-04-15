package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/royceleond/scansca/internal/db"
)

// ConnectorManager manages database connections
type ConnectorManager struct {
	registry     *db.ConnectorRegistry
	connectors   map[string]db.Connector
	databaseInfo map[string]db.DatabaseConfig
	mu           sync.RWMutex
}

// NewConnectorManager creates a new connector manager
func NewConnectorManager(registry *db.ConnectorRegistry) *ConnectorManager {
	return &ConnectorManager{
		registry:     registry,
		connectors:   make(map[string]db.Connector),
		databaseInfo: make(map[string]db.DatabaseConfig),
	}
}

// RegisterDatabase adds a new database connection
func (cm *ConnectorManager) RegisterDatabase(ctx context.Context, config db.DatabaseConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if database with this name already exists
	if _, exists := cm.connectors[config.Name]; exists {
		return fmt.Errorf("database with name '%s' already exists", config.Name)
	}

	// Create connector based on database type
	connector, err := cm.registry.Create(config.Type)
	if err != nil {
		return err
	}

	// Connect to the database
	if err := connector.Connect(ctx, config.ConnectionString); err != nil {
		return err
	}

	// Store the connector and config
	cm.connectors[config.Name] = connector
	cm.databaseInfo[config.Name] = config

	return nil
}

// GetConnector returns a connector for the specified database
func (cm *ConnectorManager) GetConnector(name string) (db.Connector, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	connector, exists := cm.connectors[name]
	if !exists {
		return nil, fmt.Errorf("database '%s' not found", name)
	}

	return connector, nil
}

// ListDatabases returns information about all registered databases
func (cm *ConnectorManager) ListDatabases() []db.DatabaseConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make([]db.DatabaseConfig, 0, len(cm.databaseInfo))
	for _, config := range cm.databaseInfo {
		// Don't include connection string in response for security
		configCopy := config
		configCopy.ConnectionString = "[REDACTED]"
		result = append(result, configCopy)
	}

	return result
}

// RemoveDatabase disconnects and removes a database
func (cm *ConnectorManager) RemoveDatabase(ctx context.Context, name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	connector, exists := cm.connectors[name]
	if !exists {
		return fmt.Errorf("database '%s' not found", name)
	}

	// Disconnect from the database
	if err := connector.Disconnect(ctx); err != nil {
		return err
	}

	// Remove from maps
	delete(cm.connectors, name)
	delete(cm.databaseInfo, name)

	return nil
}

// API Handlers

// registerDatabase handles database registration
func (s *Server) registerDatabase(c *gin.Context) {
	var config db.DatabaseConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate required fields
	if config.Name == "" || config.Type == "" || config.ConnectionString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, type and connection_string are required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := s.connMgr.RegisterDatabase(ctx, config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Database registered successfully"})
}

// listDatabases returns all registered databases
func (s *Server) listDatabases(c *gin.Context) {
	databases := s.connMgr.ListDatabases()
	c.JSON(http.StatusOK, gin.H{"databases": databases})
}

// getDatabaseInfo returns information about a specific database
func (s *Server) getDatabaseInfo(c *gin.Context) {
	dbName := c.Param("name")
	
	connector, err := s.connMgr.GetConnector(dbName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	status, err := connector.GetStatus(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":   dbName,
		"type":   connector.GetConnectorType(),
		"status": status,
	})
}

// removeDatabase handles database removal
func (s *Server) removeDatabase(c *gin.Context) {
	dbName := c.Param("name")
	
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := s.connMgr.RemoveDatabase(ctx, dbName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Database removed successfully"})
}

// listSchemas returns all schemas in a database
func (s *Server) listSchemas(c *gin.Context) {
	dbName := c.Param("name")
	
	connector, err := s.connMgr.GetConnector(dbName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	schemas, err := connector.ListSchemas(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"database": dbName,
		"schemas":  schemas,
	})
}

// listTables returns all tables in a schema
func (s *Server) listTables(c *gin.Context) {
	dbName := c.Param("name")
	schema := c.Param("schema")
	
	connector, err := s.connMgr.GetConnector(dbName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	tables, err := connector.ListTables(ctx, schema)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	views, err := connector.ListViews(ctx, schema)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to list views")
		// Continue anyway, views are optional
	}

	c.JSON(http.StatusOK, gin.H{
		"database": dbName,
		"schema":   schema,
		"tables":   tables,
		"views":    views,
	})
}

// getTableInfo returns detailed information about a table
func (s *Server) getTableInfo(c *gin.Context) {
	dbName := c.Param("name")
	schema := c.Param("schema")
	table := c.Param("table")
	
	connector, err := s.connMgr.GetConnector(dbName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	tableInfo, err := connector.GetTableInfo(ctx, schema, table)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"database":  dbName,
		"schema":    schema,
		"table":     table,
		"tableInfo": tableInfo,
	})
}

// executeQuery handles query execution
func (s *Server) executeQuery(c *gin.Context) {
	var request struct {
		Database string        `json:"database" binding:"required"`
		Query    string        `json:"query" binding:"required"`
		Params   []interface{} `json:"params"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	connector, err := s.connMgr.GetConnector(request.Database)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	result, err := connector.ExecuteQuery(ctx, request.Query, request.Params...)
	if err != nil {
		// Still return the result, as it may contain partial data and error details
		result.Error = err.Error()
		c.JSON(http.StatusOK, gin.H{"result": result})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}