package tmanager

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riversqlite"
	"github.com/sxwebdev/donejournal/internal/mcp"
	"github.com/sxwebdev/donejournal/internal/services/baseservices"
	"github.com/sxwebdev/donejournal/pkg/sqlite"
)

type Manager struct {
	riverClient *river.Client[*sql.Tx]
}

func New(
	sqliteDB *sqlite.SQLite,
	baseService *baseservices.BaseServices,
	mcpService *mcp.MCP,
) (*Manager, error) {
	workers := river.NewWorkers()
	river.AddWorker(workers, &processorWorker{
		baseServices: baseService,
		mcpService:   mcpService,
	})

	riverClient, err := river.NewClient(riversqlite.New(sqliteDB.DB), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
		Workers: workers,
	})
	if err != nil {
		return nil, err
	}

	return &Manager{
		riverClient: riverClient,
	}, nil
}

// Name returns the name of the task manager
func (m *Manager) Name() string {
	return "tmanager"
}

// Start starts the task manager
func (m *Manager) Start(ctx context.Context) error {
	return m.riverClient.Start(ctx)
}

// Stop stops the task manager
func (m *Manager) Stop(ctx context.Context) error {
	return m.riverClient.Stop(ctx)
}

// AddProcessorTask adds a new processor task to the task manager
func (m *Manager) AddProcessorTask(ctx context.Context, params ProcessorWorkerArgs) error {
	if err := params.Validate(); err != nil {
		return err
	}

	_, err := m.riverClient.Insert(ctx, &params, nil)
	return err
}
