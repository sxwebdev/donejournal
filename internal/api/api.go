package api

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sxwebdev/donejournal/internal/config"
	"github.com/sxwebdev/donejournal/internal/mcp"
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

	mcpService *mcp.MCP
}

func New(log logger.Logger, conf *config.Config, mcpService *mcp.MCP) *API {
	s := &API{
		logger:     log,
		config:     conf,
		mcpService: mcpService,
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
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid json body",
			})
		}

		if req.Text == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "text is required",
			})
		}

		if req.UserID == "" {
			req.UserID = "anonymous"
		}

		s.logger.Debugf("received: user_id=%s, text=%s", req.UserID, req.Text)

		go func(userID, text string) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
			defer cancel()

			now := time.Now()
			resp, err := s.mcpService.ParseMessage(ctx, userID, text)
			if err != nil {
				s.logger.Errorf("mcp processing failed: %v", err)
				return
			}

			s.logger.Infof("processed response in %s", time.Since(now).String())

			if len(resp.Entries) == 0 {
				return
			}

			jsonData, err := json.MarshalIndent(resp.Entries, "", "  ")
			if err != nil {
				s.logger.Errorf("failed to marshal MCP response: %v", err)
				return
			}

			s.logger.Debugf("MCP parsed response for user %s: \n%s\n", userID, string(jsonData))
		}(req.UserID, req.Text)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "ok",
			"message": "accepted",
		})
	})
}
