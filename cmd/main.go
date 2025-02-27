package main

import (
	"github.com/blog/conversational-agent/internal/agents"
	"github.com/blog/conversational-agent/internal/config"
	"github.com/blog/conversational-agent/internal/handlers"
	"github.com/blog/conversational-agent/internal/logger"
	"github.com/blog/conversational-agent/internal/middleware"
	"github.com/blog/conversational-agent/internal/router"
	"github.com/labstack/echo/v4"
)

func main() {
	// Initialize Echo server
	e := echo.New()

	log := logger.GetLogger()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Load middleware
	e.Use(middleware.LoggingMiddleware())

	// Create AgentManager
	agentManager, err := agents.NewAgentManager(
		cfg.OpenAIAPIKey,
		cfg.OpenAIModel,
		cfg.WeaviateHost,
		cfg.WeaviateAPIKey,
		cfg.WeaviateIndexName,
		5, // maxBufferMessages
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize agent manager")
	}
	// Initialize handlers
	agentHandler := &handlers.AgentHandler{
		AgentManager: agentManager,
	}

	// Register routes
	router.RegisterRoutes(e, agentHandler)

	// Start the server
	log.Info().Msg("Server is starting on port 8080")
	if err := e.Start(":8080"); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}
