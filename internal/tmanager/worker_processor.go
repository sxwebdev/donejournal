package tmanager

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/riverqueue/river"
	"github.com/sxwebdev/donejournal/internal/bot"
	"github.com/sxwebdev/donejournal/internal/processor"
)

type processorWorkerArgs struct {
	Data   string `json:"data"`
	UserID int64  `json:"user_id"`
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
	botService       *bot.Bot
}

func (w *processorWorker) Timeout(*river.Job[processorWorkerArgs]) time.Duration {
	return 120 * time.Second
}

func (w *processorWorker) Work(ctx context.Context, job *river.Job[processorWorkerArgs]) error {
	// Send typing indicator so user sees the bot is processing
	_ = w.botService.SendChatAction(ctx, job.Args.UserID, "typing")

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
