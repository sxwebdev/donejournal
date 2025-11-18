package tmanager

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/riverqueue/river"
	"github.com/sxwebdev/donejournal/internal/processor"
)

type processorWorkerArgs struct {
	Data   string
	UserID int64
}

func (processorWorkerArgs) Kind() string { return "processor" }

// Validate validates the worker arguments
func (args *processorWorkerArgs) Validate() error {
	if args.Data == "" {
		return fmt.Errorf("data is required")
	}
	if args.UserID == 0 {
		return fmt.Errorf("userID is required")
	}
	return nil
}

type processorWorker struct {
	river.WorkerDefaults[processorWorkerArgs]

	riverClient      *river.Client[*sql.Tx]
	processorService *processor.Processor
}

func (w *processorWorker) Timeout(*river.Job[processorWorkerArgs]) time.Duration {
	return time.Second * 30
}

func (w *processorWorker) Work(ctx context.Context, job *river.Job[processorWorkerArgs]) error {
	responseText, err := w.processorService.ProcessNewRequest(ctx, job.Args.UserID, job.Args.Data)
	if err != nil {
		return fmt.Errorf("failed to process new request: %w", err)
	}

	if len(responseText) > 0 {
		_, err = w.riverClient.Insert(ctx, sendMessageWorkerArgs{
			UserID: job.Args.UserID,
			Data:   responseText,
		}, nil)
		if err != nil {
			return fmt.Errorf("failed to enqueue send message task: %w", err)
		}
	}

	return nil
}
