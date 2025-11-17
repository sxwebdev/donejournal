package tmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/riverqueue/river"
	"github.com/sxwebdev/donejournal/internal/mcp"
	"github.com/sxwebdev/donejournal/internal/services/baseservices"
)

type ProcessorWorkerArgs struct {
	Data   string
	UserID string
}

func (ProcessorWorkerArgs) Kind() string { return "processor" }

// Validate validates the worker arguments
func (args *ProcessorWorkerArgs) Validate() error {
	if args.Data == "" {
		return fmt.Errorf("data is required")
	}
	if args.UserID == "" {
		return fmt.Errorf("userID is required")
	}
	return nil
}

type processorWorker struct {
	river.WorkerDefaults[ProcessorWorkerArgs]

	baseServices *baseservices.BaseServices
	mcpService   *mcp.MCP
}

func (w *processorWorker) Timeout(*river.Job[ProcessorWorkerArgs]) time.Duration {
	return time.Second * 30
}

func (w *processorWorker) Work(ctx context.Context, job *river.Job[ProcessorWorkerArgs]) error {
	resp, err := w.mcpService.ParseMessage(ctx, job.Args.UserID, job.Args.Data)
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	err = w.baseServices.Todos().BatchCreate(ctx, job.Args.UserID, resp)
	if err != nil {
		return fmt.Errorf("failed to batch create todos: %w", err)
	}
	return nil
}
