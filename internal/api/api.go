package api

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sxwebdev/donejournal/internal/config"
	"github.com/sxwebdev/donejournal/internal/tmanager"
	"github.com/tkcrm/mx/logger"
)

type IngestRequest struct {
	UserID string `json:"user_id"`
	Text   string `json:"text"`
}

type API struct {
	logger logger.Logger
	config *config.Config

	app *fiber.App

	taskManager *tmanager.Manager
}

func New(log logger.Logger, conf *config.Config, taskManager *tmanager.Manager) *API {
	s := &API{
		logger:      log,
		config:      conf,
		taskManager: taskManager,
		app: fiber.New(fiber.Config{
			DisableStartupMessage: true,
		}),
	}

	s.setupRoutes()

	return s
}

// Name returns the service name
func (s *API) Name() string {
	return "api"
}

// Start starts the API service
func (s *API) Start(ctx context.Context) error {
	var url string
	if strings.HasPrefix(s.config.Server.Addr, ":") {
		url = "http://localhost" + s.config.Server.Addr
	}

	errCh := make(chan error, 1)
	go func() {
		s.logger.Infof("Starting server on %s", url)
		errCh <- s.app.Listen(s.config.Server.Addr)
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

// Stop stops the API service
func (s *API) Stop(ctx context.Context) error {
	if s.app != nil {
		return s.app.ShutdownWithContext(ctx)
	}
	return nil
}

// setupRoutes sets up the API routes
func (s *API) setupRoutes() {
	s.app.Post("/ingest", func(c *fiber.Ctx) error {
		var req IngestRequest
		if err := c.BodyParser(&req); err != nil {
			return errorMessage(c, fiber.StatusBadRequest, err)
		}

		if req.Text == "" {
			return errorMessage(c, fiber.StatusBadRequest, errors.New("text is required"))
		}

		if req.UserID == "" {
			req.UserID = "anonymous"
		}

		s.logger.Debugf("received: user_id=%s, text=%s", req.UserID, req.Text)

		if err := s.taskManager.AddProcessorTask(c.Context(), tmanager.ProcessorWorkerArgs{
			Data:   req.Text,
			UserID: req.UserID,
		}); err != nil {
			return errorMessage(c, fiber.StatusInternalServerError, fmt.Errorf("failed to add processor task: %w", err))
		}

		return successMessage(c, "ok")
	})
}
