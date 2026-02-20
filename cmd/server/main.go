package main

import (
	"agent-service-prototype/internal/server"
	"agent-service-prototype/internal/config"
	"agent-service-prototype/pkg/logger"
)

func main() {
	// Initialize logger
	logger.Init()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load config")
	}

	serverManager, err := server.NewServerManager(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create server manager")
	}
	serverManager.Run()
}
