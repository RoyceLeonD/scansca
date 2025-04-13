package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/royceleond/antska/internal/aml/registry"
	"github.com/royceleond/antska/internal/connectors"
)

// RegisterDatabaseRequest represents the request body for registering a database
type RegisterDatabaseRequest struct {
	Name             string                 `json:"name" binding:"required"`
	Type             string                 `json:"type" binding:"required"`
	ConnectionString string                 `json:"connection_string" binding:"required"`
	Options          map[string]interface{} `json:"options,omitempty"`
}

// RegisterDatabaseHandler handles registering a new database
func RegisterDatabaseHandler(registry *registry.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request RegisterDatabaseRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		config := connectors.DatabaseConfig{
			Name:             request.Name,
			Type:             request.Type,
			ConnectionString: request.ConnectionString,
			Options:          request.Options,
		}

		err := registry.RegisterDatabase(c.Request.Context(), config)
		if err != nil {
			log.Error().Err(err).
				Str("name", request.Name).
				Str("type", request.Type).
				Msg("Failed to register database")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		log.Info().
			Str("name", request.Name).
			Str("type", request.Type).
			Msg("Database registered successfully")

		c.JSON(http.StatusCreated, gin.H{
			"message": "Database registered successfully",
			"name":    request.Name,
		})
	}
}

// ListDatabasesHandler handles listing all registered databases
func ListDatabasesHandler(registry *registry.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		databases := registry.ListDatabases()
		c.JSON(http.StatusOK, gin.H{
			"databases": databases,
		})
	}
}

// UnregisterDatabaseHandler handles removing a database registration
func UnregisterDatabaseHandler(registry *registry.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Database name is required"})
			return
		}

		err := registry.UnregisterDatabase(c.Request.Context(), name)
		if err != nil {
			log.Error().Err(err).
				Str("name", name).
				Msg("Failed to unregister database")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		log.Info().
			Str("name", name).
			Msg("Database unregistered successfully")

		c.JSON(http.StatusOK, gin.H{
			"message": "Database unregistered successfully",
		})
	}
}

// ListSchemasHandler handles listing schemas for a database
func ListSchemasHandler(registry *registry.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Database name is required"})
			return
		}

		connector, err := registry.GetConnector(name)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		schemas, err := connector.ListSchemas(c.Request.Context())
		if err != nil {
			log.Error().Err(err).
				Str("database", name).
				Msg("Failed to list schemas")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"database": name,
			"schemas":  schemas,
		})
	}
}

// ListTablesHandler handles listing tables in a schema
func ListTablesHandler(registry *registry.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		schema := c.Param("schema")
		if name == "" || schema == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Database name and schema are required"})
			return
		}

		connector, err := registry.GetConnector(name)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		tables, err := connector.ListTables(c.Request.Context(), schema)
		if err != nil {
			log.Error().Err(err).
				Str("database", name).
				Str("schema", schema).
				Msg("Failed to list tables")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"database": name,
			"schema":   schema,
			"tables":   tables,
		})
	}
}

// GetTableColumnsHandler handles retrieving column information for a table
func GetTableColumnsHandler(registry *registry.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		schema := c.Param("schema")
		table := c.Param("table")
		if name == "" || schema == "" || table == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Database name, schema, and table are required"})
			return
		}

		connector, err := registry.GetConnector(name)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		columns, err := connector.GetTableColumns(c.Request.Context(), schema, table)
		if err != nil {
			log.Error().Err(err).
				Str("database", name).
				Str("schema", schema).
				Str("table", table).
				Msg("Failed to get table columns")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"database": name,
			"schema":   schema,
			"table":    table,
			"columns":  columns,
		})
	}
}