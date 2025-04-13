package mci

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/royceleond/antska/internal/aml/registry"
	"github.com/royceleond/antska/internal/aml/scheduler"
	"github.com/royceleond/antska/internal/mci/api"
	"github.com/royceleond/antska/internal/mci/middleware"
)

// Server represents the MCI HTTP server
type Server struct {
	router    *gin.Engine
	httpServer *http.Server
	registry  *registry.Registry
	scheduler *scheduler.Scheduler
}

// NewServer creates a new MCI server
func NewServer(registry *registry.Registry, scheduler *scheduler.Scheduler) *Server {
	// Set Gin to release mode in production
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Apply global middleware
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())

	server := &Server{
		router:    router,
		registry:  registry,
		scheduler: scheduler,
	}

	// Setup routes
	server.setupRoutes()

	return server
}

// setupRoutes configures the API routes
func (s *Server) setupRoutes() {
	// API version group
	v1 := s.router.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", api.HealthHandler())

		// Database management
		databases := v1.Group("/databases")
		{
			databases.GET("", api.ListDatabasesHandler(s.registry))
			databases.POST("", api.RegisterDatabaseHandler(s.registry))
			databases.DELETE("/:name", api.UnregisterDatabaseHandler(s.registry))
			databases.GET("/:name/schemas", api.ListSchemasHandler(s.registry))
			databases.GET("/:name/schemas/:schema/tables", api.ListTablesHandler(s.registry))
			databases.GET("/:name/schemas/:schema/tables/:table/columns", api.GetTableColumnsHandler(s.registry))
		}

		// Query execution
		v1.POST("/query", api.ExecuteQueryHandler(s.registry))

		// Job management
		jobs := v1.Group("/jobs")
		{
			jobs.GET("", api.ListJobsHandler(s.scheduler))
			jobs.POST("", api.CreateJobHandler(s.registry, s.scheduler))
			jobs.GET("/:id", api.GetJobHandler(s.scheduler))
			jobs.DELETE("/:id", api.DeleteJobHandler(s.scheduler))
			jobs.POST("/:id/run", api.RunJobHandler(s.scheduler))
		}
	}
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	log.Info().Str("address", addr).Msg("Starting MCI HTTP server")
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

// Stop gracefully shuts down the HTTP server
func (s *Server) Stop() error {
	log.Info().Msg("Shutting down MCI HTTP server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	return nil
}