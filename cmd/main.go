package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/royceleond/scansca/internal/db"
	"github.com/royceleond/scansca/internal/db/connectors/postgresql"
	"github.com/royceleond/scansca/internal/server"
)

// Version information (set by build flag)
var (
	version = "dev"
)

func main() {
	// Setup logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	
	// Log version information
	log.Info().Str("version", version).Msg("Starting Scansca MCP Server")
	
	// Load configuration
	loadConfig()
	
	// Initialize database registry
	dbRegistry := db.InitializeRegistry()
	
	// Register database connectors
	// Import the connector packages here to avoid import cycles
	dbRegistry.Register("postgresql", func() db.Connector {
		return &postgresql.PostgresConnector{
			Config: db.DefaultConnectionPoolConfig(),
		}
	})
	dbRegistry.Register("postgres", func() db.Connector {
		return &postgresql.PostgresConnector{
			Config: db.DefaultConnectionPoolConfig(),
		}
	}) // Alias
	
	// Setup connection manager
	connMgr := server.NewConnectorManager(dbRegistry)
	
	// Create and initialize server
	srv := server.New(server.DefaultConfig(), connMgr)
	srv.Initialize()
	
	// Configure from environment
	if port := viper.GetInt("server.port"); port > 0 {
		log.Info().Int("port", port).Msg("Using configured port")
		srv.SetPort(port)
	}
	
	// Start server in goroutine
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()
	
	log.Info().Msg("Scansca MCP Server started successfully")
	
	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Info().Msg("Shutting down server...")
	
	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}
	
	log.Info().Msg("Server exited gracefully")
}

// loadConfig loads configuration from config files and environment variables
func loadConfig() {
	viper.SetConfigName("scansca")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")
	
	// Set defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	
	// Environment variables
	viper.SetEnvPrefix("SCANSCA")
	viper.AutomaticEnv()
	
	// Read configuration
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Warn().Msg("Config file not found, using defaults and environment variables")
		} else {
			log.Error().Err(err).Msg("Error reading config file")
		}
	} else {
		log.Info().Str("file", viper.ConfigFileUsed()).Msg("Config loaded")
	}
}