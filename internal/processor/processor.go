package processor

import (
	"context"
	"time"

	"github.com/sxwebdev/donejournal/internal/mcp"
	"github.com/sxwebdev/donejournal/internal/models"
	"github.com/sxwebdev/donejournal/internal/services/baseservices"
	"github.com/sxwebdev/donejournal/pkg/loop"
	"github.com/tkcrm/mx/logger"
)

type Processor struct {
	logger      logger.Logger
	baseService *baseservices.BaseServices
	mcpService  *mcp.MCP
	looper      *loop.Loop
}

func New(
	l logger.Logger,
	baseService *baseservices.BaseServices,
	mcpService *mcp.MCP,
) *Processor {
	s := &Processor{
		logger:      l,
		baseService: baseService,
		mcpService:  mcpService,
	}

	s.looper = loop.New(
		s.initProcessor,
		loop.WithLeading(),
		loop.WithPeriod(time.Second*2),
		loop.WithContextTimeout(time.Second*30),
	)

	return s
}

// Name returns processor name
func (s *Processor) Name() string {
	return "processor"
}

// Start starts the processor
func (s *Processor) Start(ctx context.Context) error {
	s.looper.Start(ctx)
	return nil
}

// Stop stops the processor
func (s *Processor) Stop(ctx context.Context) error {
	s.looper.Stop()
	s.looper.Wait()
	return nil
}

// initProcessor initializes the processor loop
func (s *Processor) initProcessor(ctx context.Context) {
	if err := s.do(ctx); err != nil {
		s.logger.Errorf("failed to process notification history: %v", err)
	}
}

// do processes pending requests
func (s *Processor) do(ctx context.Context) error {
	// items, err := s.baseService.Requests().GetPendingRequests(ctx)
	// if err != nil {
	// 	return fmt.Errorf("get pending requests: %w", err)
	// }

	// if len(items) == 0 {
	// 	return nil
	// }

	// s.logger.Infof("found %d pending requests", len(items))

	// for _, item := range items {
	// 	if err := s.processItem(ctx, item); err != nil {
	// 		s.logger.Errorf("process item %s: %v", item.ID, err)
	// 		continue
	// 	}
	// }

	return nil
}

// processItem processes a single request item
func (s *Processor) processItem(ctx context.Context, item *models.Inbox) error {
	// resp, err := s.mcpService.ParseMessage(ctx, item.UserID, item.Data)
	// if err != nil {
	// 	return fmt.Errorf("failed to parse message: %w", err)
	// }

	// err = s.baseService.Todos().BatchCreate(ctx, item, resp)
	// if err != nil {
	// 	return fmt.Errorf("failed to batch create todos: %w", err)
	// }

	return nil
}
