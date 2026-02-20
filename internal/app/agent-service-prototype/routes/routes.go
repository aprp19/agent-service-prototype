package routes

import (
	"database/sql"

	"agent-service-prototype/internal/app/agent-service-prototype/controller"
	"agent-service-prototype/internal/app/agent-service-prototype/repository"
	"agent-service-prototype/internal/app/agent-service-prototype/service"
	"agent-service-prototype/internal/config"
	"agent-service-prototype/pkg/logger"

	"github.com/labstack/echo/v4"
)

// RegisterSetupRoutes registers POST /setup/installation and GET /setup/status
// on the root Echo instance (not under /api/v1).
func RegisterSetupRoutes(e *echo.Echo, cfg *config.Config, db *sql.DB) {
	repo := repository.NewRepository(db)
	svc := service.NewService(cfg, repo)
	ctrl := controller.NewController(svc)

	g := e.Group("/setup")
	g.POST("/installation", ctrl.Installation)
	g.GET("/status", ctrl.Status)

	logger.Info().Msg("setup routes registered")
}

// RegisterRoutes registers all HTTP routes for the agent-service-prototype module
// Currently commented out — Ent code generation has not been run yet.
// func RegisterRoutes(api *echo.Group, protectedAPI *echo.Group, client *ent.Client, cfg *config.Config, db *sql.DB) {
// 	if client == nil {
// 		logger.Error().Msg("⚠️ Database client not available for agent-service-prototype module")
// 		return
// 	}
//
// 	logger.Info().Msg("✅ agent-service-prototype database client available")
//
// 	repo := repository.NewRepository(client, db)
// 	svc := service.NewService(repo)
// 	ctrl := controller.NewController(svc)
//
// 	exampleGroup := api.Group("/example")
// 	{
// 		exampleGroup.GET("/:id", ctrl.ExampleGetByID)
// 	}
//
// 	protectedExampleGroup := protectedAPI.Group("/example")
// 	{
// 		_ = protectedExampleGroup
// 	}
//
// 	logger.Info().Msg("🔗 agent-service-prototype routes registered")
// }

// RegisterGRPCServices registers all gRPC services for the module
// Currently commented out — Ent code generation has not been run yet.
// func RegisterGRPCServices(grpcServer *grpc.Server, client *ent.Client, cfg *config.Config) {
// 	if client == nil {
// 		logger.Error().Msg("⚠️ Database client not available for gRPC module")
// 		return
// 	}
//
// 	logger.Info().Msg("✅ gRPC database client available")
// 	logger.Info().Msg("🔗 gRPC services registered")
// }
