package todos

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dromara/carbon/v2"
	"github.com/sxwebdev/donejournal/internal/mcp"
	"github.com/sxwebdev/donejournal/internal/models"
	"github.com/sxwebdev/donejournal/internal/store/repos"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_todos"
	"github.com/sxwebdev/donejournal/internal/store/storecmn"
	"github.com/sxwebdev/donejournal/pkg/utils"
)

// Create a new todo
func (s *Service) Create(ctx context.Context, tx *sql.Tx, userID string, entry mcp.ParsedEntry) (*models.Todo, error) {
	req := repo_todos.CreateParams{
		ID:          utils.GenerateULID(),
		UserID:      userID,
		Title:       entry.Title,
		Description: entry.Description,
		// RequestID:   &request.ID,
	}

	date := carbon.Parse(entry.Date).StdTime()

	if date.IsZero() {
		return nil, fmt.Errorf("invalid date: %s", entry.Date)
	}

	req.PlannedDate = date

	switch entry.Kind {
	case mcp.EntryKindTodo:
		req.Status = models.TodoStatusPending
	case mcp.EntryKindDone:
		req.Status = models.TodoStatusCompleted
	default:
		return nil, fmt.Errorf("invalid entry kind: %s", entry.Kind)
	}

	// update reequest status
	// err := s.store.Requests(repos.WithTx(tx)).UpdateStatus(ctx, repo_requests.UpdateStatusParams{
	// 	ID:     request.ID,
	// 	Status: models.RequestStatusCompleted,
	// })
	// if err != nil {
	// 	return nil, err
	// }

	// create the todo
	todo, err := s.store.Todos(repos.WithTx(tx)).Create(ctx, req)
	if err != nil {
		return nil, err
	}

	return todo, nil
}

// BatchCreate creates todos in batch
func (s *Service) BatchCreate(ctx context.Context, userID string, parsedResponse *mcp.ParsedResponse) error {
	if len(parsedResponse.Entries) == 0 {
		return nil
	}

	err := storecmn.WrapTx(ctx, s.store.SQLite(), func(tx *sql.Tx) error {
		for _, entry := range parsedResponse.Entries {
			if _, err := s.Create(ctx, tx, userID, entry); err != nil {
				return fmt.Errorf("create todo for entry '%s': %w", entry.Title, err)
			}
		}

		return nil
	})

	return err
}
