package tmanager

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite"

	"github.com/mymmrac/telego"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riversqlite"
	"github.com/sxwebdev/donejournal/internal/bot"
	"github.com/sxwebdev/donejournal/internal/processor"
	"github.com/sxwebdev/donejournal/pkg/sqlite"
	"github.com/tkcrm/mx/logger"
)

type Manager struct {
	logger           logger.Logger
	riverClient      *river.Client[*sql.Tx]
	botService       *bot.Bot
	processorService *processor.Processor
}

func New(
	l logger.Logger,
	sqliteDB *sqlite.SQLite,
	processorService *processor.Processor,
	botService *bot.Bot,
) (*Manager, error) {
	workers := river.NewWorkers()

	// Add send message worker
	river.AddWorker(workers, &sendMessageWorker{
		botService: botService,
	})

	// Add processor worker
	pWorker := &processorWorker{
		processorService: processorService,
	}

	river.AddWorker(workers, pWorker)

	// Create river client
	riverClient, err := river.NewClient(riversqlite.New(sqliteDB.DB), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
		Workers: workers,
	})
	if err != nil {
		return nil, err
	}

	pWorker.riverClient = riverClient

	return &Manager{
		logger:           l,
		riverClient:      riverClient,
		botService:       botService,
		processorService: processorService,
	}, nil
}

// Name returns the name of the task manager
func (m *Manager) Name() string {
	return "tmanager"
}

// Start starts the task manager
func (m *Manager) Start(ctx context.Context) error {
	if err := m.riverClient.Start(ctx); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case update := <-m.botService.OnUpdate():
				m.handleUpdate(ctx, update)
			}
		}
	}()

	return nil
}

// Stop stops the task manager
func (m *Manager) Stop(ctx context.Context) error {
	return m.riverClient.Stop(ctx)
}

// AddProcessorTask adds a new processor task to the task manager
func (m *Manager) AddProcessorTask(ctx context.Context, userID int64, data string) error {
	params := processorWorkerArgs{
		UserID: userID,
		Data:   data,
	}

	if err := params.Validate(); err != nil {
		return err
	}

	_, err := m.riverClient.Insert(ctx, &params, nil)
	return err
}

// handleUpdate handles incoming updates from the bot.
func (m *Manager) handleUpdate(ctx context.Context, update telego.Update) {
	// Handle callback queries
	if update.CallbackQuery != nil {
		if err := m.processorService.HandleCallbackQuery(ctx, update.CallbackQuery); err != nil {
			m.logger.Errorf("failed to handle callback query: %v", err)
		}
		return
	}

	// Handle messages
	if update.Message != nil {
		// Handle commands (messages starting with /)
		if len(update.Message.Text) > 0 && update.Message.Text[0] == '/' {
			if err := m.processorService.HandleCommand(ctx, update.Message.Chat.ID, update.Message.Text); err != nil {
				m.logger.Errorf("failed to handle command: %v", err)
			}
			return
		}

		// Handle regular messages - add to processor task
		err := m.AddProcessorTask(ctx, update.Message.From.ID, update.Message.Text)
		if err != nil {
			m.logger.Errorf("failed to enqueue processor task: %v", err)
		}
	}
}
