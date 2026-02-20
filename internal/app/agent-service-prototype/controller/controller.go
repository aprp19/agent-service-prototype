package controller

import (
	"net/http"
	"time"

	"agent-service-prototype/internal/app/agent-service-prototype/dto"
	"agent-service-prototype/internal/app/agent-service-prototype/service"
	"agent-service-prototype/pkg/setup"

	"github.com/labstack/echo/v4"
)

type Controller struct {
	service *service.Service
}

func NewController(svc *service.Service) *Controller {
	return &Controller{service: svc}
}

// Installation handles POST /setup/installation
func (c *Controller) Installation(ctx echo.Context) error {
	if !c.service.TryStart() {
		s := c.service.GetStatus()
		return ctx.JSON(http.StatusConflict, setup.NewStatusPayload(s.Status, s.Step, s.Error, s.StartedAt, s.FinishedAt))
	}

	result := c.service.RunInstallation(ctx.Request().Context())

	if result.Success {
		return ctx.JSON(http.StatusOK, dto.InstallationSuccess{
			Status:          "SUCCESS",
			SchemaVersion:   result.SchemaVersion,
			DurationSeconds: result.Duration.Seconds(),
		})
	}

	return ctx.JSON(http.StatusOK, dto.InstallationFailed{
		Status: "FAILED",
		Step:   result.Step,
		Error:  result.Error,
	})
}

// Status handles GET /setup/status
func (c *Controller) Status(ctx echo.Context) error {
	s := c.service.GetStatus()
	if s == nil {
		return ctx.JSON(http.StatusOK, setup.NewStatusPayload("idle", "", "", time.Time{}, time.Time{}))
	}
	return ctx.JSON(http.StatusOK, setup.NewStatusPayload(s.Status, s.Step, s.Error, s.StartedAt, s.FinishedAt))
}
