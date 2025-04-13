package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/royceleond/antska/internal/aml/registry"
	"github.com/royceleond/antska/internal/aml/scheduler"
	"github.com/royceleond/antska/internal/aml/state"
	"github.com/royceleond/antska/internal/mci"
	"github.com/royceleond/antska/pkg/connectors"
)

func main() {
	// Setup logging
	setupLogging()

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Create state manager
	stateManager, err := state.NewStateManager(config.ConfigDir, "antska-state.json")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create state manager")
	}

	// Create connector factory
	connectorFactory := connectors.NewConnectorFactory()

	// Create registry
	dbRegistry := registry.NewRegistry(connectorFactory)

	// Create scheduler
	jobScheduler := scheduler.NewScheduler()
	jobScheduler.Start()

	// Create and start MCI server
	mciServer := mci.NewServer(dbRegistry, jobScheduler)
	go func() {
		if err := mciServer.Start(fmt.Sprintf(":%d", config.HttpPort)); err != nil {
			log.Fatal().Err(err).Msg("Failed to start MCI server")
		}
	}()

	log.Info().
		Str("version", "0.1.0").
		Int("http_port", config.HttpPort).
		Msg("Antska MCP server started")

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Perform cleanup
	ctx := context.Background()
	
	// Stop the MCI server
	if err := mciServer.Stop(); err != nil {
		log.Error().Err(err).Msg("Error stopping MCI server")
	}

	// Stop the scheduler
	jobScheduler.Stop()

	// Close all database connections
	dbRegistry.CloseAll(ctx)

	// Save state
	if err := stateManager.Save(); err != nil {
		log.Error().Err(err).Msg("Error saving state")
	}

	log.Info().Msg("Server stopped")
}

// Config holds the application configuration
type Config struct {
	HttpPort  int
	ConfigDir string
}

// loadConfig loads the application configuration
func loadConfig() (Config, error) {
	// Set default values
	config := Config{
		HttpPort:  8080,
		ConfigDir: "./config",
	}

	// Initialize viper
	v := viper.New()
	
	// Set config defaults
	v.SetDefault("http.port", config.HttpPort)
	v.SetDefault("config.dir", config.ConfigDir)

	// Look for config file
	v.SetConfigName("antska")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath("$HOME/.antska")
	v.AddConfigPath("/etc/antska")

	// Read environment variables
	v.AutomaticEnv()
	v.SetEnvPrefix("ANTSKA")

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		// It's okay if there's no config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return config, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Assign values from config or environment
	config.HttpPort = v.GetInt("http.port")
	config.ConfigDir = v.GetString("config.dir")

	return config, nil
}

// setupLogging configures the logging system
func setupLogging() {
	// Configure zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Check if we're in a TTY
	if isTerminal() {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Set log level from environment
	logLevel := os.Getenv("ANTSKA_LOG_LEVEL")
	if logLevel != "" {
		level, err := zerolog.ParseLevel(logLevel)
		if err != nil {
			log.Warn().Str("level", logLevel).Msg("Invalid log level, using default")
		} else {
			zerolog.SetGlobalLevel(level)
			log.Info().Str("level", level.String()).Msg("Log level set")
		}
	}
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}