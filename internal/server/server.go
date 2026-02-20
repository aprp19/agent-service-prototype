package server

import (
	"bufio"
	"context"
	"database/sql"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"agent-service-prototype/internal/bootstrap"
	"agent-service-prototype/internal/config"
	"agent-service-prototype/internal/router"
	"agent-service-prototype/pkg/logger"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

type ServerManager struct {
	HTTP         *echo.Echo
	RawDB        *sql.DB
	cfg          *config.Config
}

func NewServerManager(cfg *config.Config) (*ServerManager, error) {
	rawDB := bootstrap.InitDatabase(cfg)

	e := echo.New()

	// Global middleware
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORS())

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// Create API groups (kept for future use when Ent is generated)
	// api := e.Group("/api/v1")
	// protectedAPI := e.Group("/api/v1")
	// protectedAPI.Use(middleware.AuthMiddleware(cfg))

	// Setup HTTP router
	router.InitRouter(e, cfg, rawDB)

	return &ServerManager{
		HTTP:         e,
		RawDB:        rawDB,
		cfg:          cfg,
	}, nil
}

func (s *ServerManager) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	shutdownConfirmed := make(chan bool, 1)

	go func() {
		for {
			<-sigChan
			logger.Warn().Msg("⚠️  Shutdown signal received (Ctrl+C)")
			logger.Info().Msg("🔄 Are you sure you want to shutdown the server? (Y/N): ")

			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				logger.Error().Err(err).Msg("Failed to read input")
				logger.Info().Msg("⏸️  Shutdown cancelled, server continues running")
				continue
			}

			response = strings.TrimSpace(strings.ToUpper(response))

			if response == "Y" || response == "YES" {
				logger.Info().Msg("✅ Shutdown confirmed")
				shutdownConfirmed <- true
				return
			} else {
				logger.Info().Msg("⏸️  Shutdown cancelled, server continues running")
			}
		}
	}()

	// Start HTTP server
	go func() {
		logger.Info().Str("port", s.cfg.HTTP.Port).Msg("🚀 Starting HTTP server")
		if err := s.HTTP.Start(":" + s.cfg.HTTP.Port); err != nil {
			logger.Error().Err(err).Msg("Failed to start HTTP server")
		}
	}()

	// Wait for confirmed shutdown
	<-shutdownConfirmed
	cancel()

	<-ctx.Done()

	if err := s.Shutdown(context.Background()); err != nil {
		logger.Error().Err(err).Msg("Failed to shutdown servers gracefully")
	}

	bootstrap.CloseDatabase(s.RawDB)
	logger.Info().Msg("🛑 Servers stopped")
	return nil
}

func (s *ServerManager) Shutdown(ctx context.Context) error {
	if err := s.HTTP.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("Failed to shutdown HTTP server")
	}

	return nil
}
