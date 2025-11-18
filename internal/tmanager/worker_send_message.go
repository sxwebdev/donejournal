package tmanager

import (
	"context"
	"time"

	"github.com/riverqueue/river"
	"github.com/sxwebdev/donejournal/internal/bot"
)

type sendMessageWorkerArgs struct {
	Data   string
	UserID int64
}

func (sendMessageWorkerArgs) Kind() string { return "send_message" }

type sendMessageWorker struct {
	river.WorkerDefaults[sendMessageWorkerArgs]

	botService *bot.Bot
}

func (w *sendMessageWorker) Timeout(*river.Job[sendMessageWorkerArgs]) time.Duration {
	return time.Second * 30
}

func (w *sendMessageWorker) Work(ctx context.Context, job *river.Job[sendMessageWorkerArgs]) error {
	return w.botService.SendMessage(ctx, job.Args.UserID, job.Args.Data)
}
