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
	"github.com/royceleond/scansca/internal/mcp"
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
	
	// Get server configuration
	host := viper.GetString("server.host")
	port := viper.GetInt("server.port")
	
	// Create and initialize MCP server
	mcpServer, err := mcp.NewMCPServer(host, port, connMgr)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create MCP server")
	}
	
	// Register MCP tools
	if err := mcpServer.RegisterTools(); err != nil {
		log.Fatal().Err(err).Msg("Failed to register MCP tools")
	}
	
	// Create and initialize standard HTTP server (for legacy API and monitoring)
	srv := server.New(server.DefaultConfig(), connMgr)
	srv.Initialize()
	
	// Use a different port for the standard API
	apiPort := viper.GetInt("server.api_port")
	if apiPort <= 0 {
		apiPort = port + 1 // Default to MCP port + 1
	}
	srv.SetPort(apiPort)
	
	// Start server in goroutine
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatal().Err(err).Msg("Failed to start API server")
		}
	}()
	
	// Start MCP server in goroutine
	go func() {
		log.Info().Str("host", host).Int("port", port).Msg("Starting MCP server with SSE")
		if err := mcpServer.Start(); err != nil {
			log.Fatal().Err(err).Msg("Failed to start MCP server")
		}
	}()
	
	log.Info().Msg("Scansca servers started successfully")
	
	// Wait for interrupt signal to gracefully shut down the servers
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Info().Msg("Shutting down servers...")
	
	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Shutdown both servers
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("API server forced to shutdown")
	}
	
	if err := mcpServer.Shutdown(); err != nil {
		log.Error().Err(err).Msg("MCP server forced to shutdown")
	}
	
	log.Info().Msg("Servers exited gracefully")
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
	viper.SetDefault("server.api_port", 8081)
	
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