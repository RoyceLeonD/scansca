package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/royceleond/antska/internal/aml/registry"
)

// ExecuteQueryRequest represents the request body for executing a query
type ExecuteQueryRequest struct {
	Database string `json:"database" binding:"required"`
	Query    string `json:"query" binding:"required"`
}

// ExecuteQueryHandler handles executing queries against a database
func ExecuteQueryHandler(registry *registry.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request ExecuteQueryRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get the connector for the specified database
		connector, err := registry.GetConnector(request.Database)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Execute the query
		result, err := connector.ExecuteQuery(c.Request.Context(), request.Query)
		if err != nil {
			log.Error().Err(err).
				Str("database", request.Database).
				Str("query", request.Query).
				Msg("Failed to execute query")
			
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"result": result, // Include partial results if available
			})
			return
		}

		// If the result has an error field, it's a partial result
		if result.Error != "" {
			c.JSON(http.StatusPartialContent, gin.H{
				"database": request.Database,
				"result":   result,
				"message":  "Query executed with errors",
			})
			return
		}

		// Return the successful result
		c.JSON(http.StatusOK, gin.H{
			"database": request.Database,
			"result":   result,
		})
	}
}

// HealthHandler handles health check requests
func HealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}
}