package tmanager

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/riverqueue/river"
	"github.com/sxwebdev/donejournal/internal/bot"
	"github.com/sxwebdev/donejournal/internal/stt"
)

type voiceWorkerArgs struct {
	UserID int64
	FileID string
}

func (voiceWorkerArgs) Kind() string { return "voice" }

type voiceWorker struct {
	river.WorkerDefaults[voiceWorkerArgs]

	riverClient *river.Client[*sql.Tx]
	botService  *bot.Bot
	sttService  *stt.Service
}

func (w *voiceWorker) Timeout(*river.Job[voiceWorkerArgs]) time.Duration {
	return 5 * time.Minute
}

func (w *voiceWorker) Work(ctx context.Context, job *river.Job[voiceWorkerArgs]) error {
	data, err := w.botService.DownloadFile(ctx, job.Args.FileID)
	if err != nil {
		return fmt.Errorf("failed to download voice file: %w", err)
	}

	text, err := w.sttService.Transcribe(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to transcribe voice: %w", err)
	}

	if text == "" {
		return w.botService.SendMessage(ctx, job.Args.UserID, "⚠️ Не удалось распознать речь в голосовом сообщении.")
	}

	_, err = w.riverClient.Insert(ctx, &processorWorkerArgs{
		UserID: job.Args.UserID,
		Data:   text,
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to enqueue processor task: %w", err)
	}

	return nil
}
