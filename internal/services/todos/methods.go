package todos

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dromara/carbon/v2"
	"github.com/sxwebdev/donejournal/internal/mcp"
	"github.com/sxwebdev/donejournal/internal/models"
	"github.com/sxwebdev/donejournal/internal/store/repos"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_todos"
	"github.com/sxwebdev/donejournal/internal/store/storecmn"
	"github.com/sxwebdev/donejournal/pkg/utils"
)

// Create a new todo from a parsed MCP entry
func (s *Service) Create(ctx context.Context, tx *sql.Tx, userID int64, entry mcp.ParsedEntry) (*models.Todo, error) {
	req := repo_todos.CreateParams{
		ID:          utils.GenerateULID(),
		UserID:      userID,
		Title:       entry.Title,
		Description: entry.Description,
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
		req.CompletedAt = &req.PlannedDate
	default:
		return nil, fmt.Errorf("invalid entry kind: %s", entry.Kind)
	}

	todo, err := s.store.Todos(repos.WithTx(tx)).Create(ctx, req)
	if err != nil {
		return nil, err
	}

	return todo, nil
}

// BatchCreate creates todos in batch
func (s *Service) BatchCreate(ctx context.Context, userID int64, parsedResponse *mcp.ParsedResponse) error {
	if len(parsedResponse.Entries) == 0 {
		return nil
	}

	if err := storecmn.WrapTx(ctx, s.store.SQLite(), func(tx *sql.Tx) error {
		for _, entry := range parsedResponse.Entries {
			if _, err := s.Create(ctx, tx, userID, entry); err != nil {
				return fmt.Errorf("create todo for entry '%s': %w", entry.Title, err)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	s.broker.Publish(TodoEvent{UserID: userID})
	return nil
}

// Find todos by params
func (s *Service) Find(ctx context.Context, params repo_todos.FindParams) (*storecmn.FindResponseWithCount[*models.Todo], error) {
	return s.store.Todos().Find(ctx, params)
}

// CreateFromAPI creates a new todo from API request (without MCP)
func (s *Service) CreateFromAPI(ctx context.Context, userID int64, title, description string, plannedDate time.Time) (*models.Todo, error) {
	todo, err := s.store.Todos().Create(ctx, repo_todos.CreateParams{
		ID:          utils.GenerateULID(),
		UserID:      userID,
		Title:       title,
		Description: description,
		Status:      models.TodoStatusPending,
		PlannedDate: plannedDate,
	})
	if err != nil {
		return nil, err
	}
	s.broker.Publish(TodoEvent{UserID: userID})
	return todo, nil
}

// GetByID returns a single todo
func (s *Service) GetByID(ctx context.Context, id string) (*models.Todo, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}
	return s.store.Todos().GetByID(ctx, id)
}

// UpdateParams contains optional fields for partial update
type UpdateParams struct {
	Title       *string
	Description *string
	Status      *models.TodoStatusType
	PlannedDate *time.Time
}

// Update performs a partial update on a todo
func (s *Service) Update(ctx context.Context, userID int64, id string, params UpdateParams) (*models.Todo, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}

	sets := []string{}
	args := []any{}

	if params.Title != nil {
		sets = append(sets, "title = ?")
		args = append(args, *params.Title)
	}
	if params.Description != nil {
		sets = append(sets, "description = ?")
		args = append(args, *params.Description)
	}
	if params.Status != nil {
		sets = append(sets, "status = ?")
		args = append(args, string(*params.Status))
	}
	if params.PlannedDate != nil {
		sets = append(sets, "planned_date = ?")
		args = append(args, *params.PlannedDate)
	}

	if len(sets) == 0 {
		return s.store.Todos().GetByID(ctx, id)
	}

	query := "UPDATE todos SET "
	for i, set := range sets {
		if i > 0 {
			query += ", "
		}
		query += set
	}
	query += ", updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	args = append(args, id)

	_, err := s.store.SQLite().ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	todo, err := s.store.Todos().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.broker.Publish(TodoEvent{UserID: userID})
	return todo, nil
}

// Complete marks a todo as completed
func (s *Service) Complete(ctx context.Context, userID int64, id string) (*models.Todo, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}

	now := time.Now()
	_, err := s.store.SQLite().ExecContext(ctx,
		"UPDATE todos SET status = ?, completed_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		models.TodoStatusCompleted, now, id,
	)
	if err != nil {
		return nil, err
	}

	todo, err := s.store.Todos().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.broker.Publish(TodoEvent{UserID: userID})
	return todo, nil
}

// Delete deletes a todo
func (s *Service) Delete(ctx context.Context, userID int64, id string) error {
	if id == "" {
		return storecmn.ErrEmptyID
	}
	if err := s.store.Todos().Delete(ctx, id); err != nil {
		return err
	}
	s.broker.Publish(TodoEvent{UserID: userID})
	return nil
}
