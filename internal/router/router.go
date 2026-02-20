package router

import (
	"database/sql"

	"agent-service-prototype/internal/app/agent-service-prototype/routes"
	"agent-service-prototype/internal/config"
	"agent-service-prototype/pkg/logger"

	"github.com/labstack/echo/v4"
	// "google.golang.org/grpc"
)

// InitRouter initializes all HTTP routes by calling RegisterRoutes for each module
func InitRouter(e *echo.Echo, cfg *config.Config, db *sql.DB) {
	logger.Info().Msg("🔗 Initializing HTTP router...")

	// Register setup routes (root-level, no auth)
	routes.RegisterSetupRoutes(e, cfg, db)

	// RegisterRoutes commented out — Ent code generation has not been run yet.
	// routes.RegisterRoutes(api, protectedAPI, client, cfg, db)

	logger.Info().Msg("✅ HTTP router initialization completed")
}