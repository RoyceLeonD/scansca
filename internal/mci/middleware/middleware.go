package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Logger returns a middleware that logs request information
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Calculate request duration
		duration := time.Since(start)
		status := c.Writer.Status()

		// Log the request
		logger := log.Info()
		if status >= 400 {
			logger = log.Error()
		}

		logger.
			Str("method", method).
			Str("path", path).
			Int("status", status).
			Dur("duration", duration).
			Int("size", c.Writer.Size()).
			Str("client_ip", c.ClientIP()).
			Msg("HTTP request")
	}
}

// Recovery returns a middleware that recovers from panics
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the error
				log.Error().
					Interface("error", err).
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.Path).
					Msg("Panic recovered")

				// Return internal server error
				c.AbortWithStatusJSON(500, gin.H{
					"error": "Internal server error",
				})
			}
		}()

		c.Next()
	}
}

// CORS returns a middleware that handles Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Authentication returns a middleware that handles JWT authentication
// This is a placeholder - actual implementation would verify JWT tokens
func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the auth header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		// TODO: Validate the JWT token
		// For now, we'll just continue

		c.Next()
	}
}