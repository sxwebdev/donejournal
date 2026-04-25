package todos

import (
	"context"
	"database/sql"
	"time"

	"github.com/sxwebdev/donejournal/internal/models"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_todos"
	"github.com/sxwebdev/donejournal/internal/store/storecmn"
	"github.com/sxwebdev/donejournal/pkg/utils"
)

// Count returns the number of todos matching the given filters.
func (s *Service) Count(ctx context.Context, params repo_todos.FindParams) (uint32, error) {
	return s.store.Todos().Count(ctx, params)
}

// Find todos by params
func (s *Service) Find(ctx context.Context, params repo_todos.FindParams) (*storecmn.FindResponseWithCount[*models.Todo], error) {
	return s.store.Todos().Find(ctx, params)
}

// CreateFromAPI creates a new todo from API request
func (s *Service) CreateFromAPI(ctx context.Context, userID int64, title, description string, plannedDate time.Time, workspaceID *string, priority models.TodoPriorityType, recurrenceRule *string) (*models.Todo, error) {
	if priority == "" {
		priority = models.TodoPriorityNone
	}
	todo, err := s.store.Todos().Create(ctx, repo_todos.CreateParams{
		ID:             utils.GenerateULID(),
		UserID:         userID,
		Title:          title,
		Description:    description,
		Status:         models.TodoStatusPending,
		PlannedDate:    plannedDate,
		WorkspaceID:    storecmn.PtrToNullString(workspaceID),
		Priority:       priority,
		RecurrenceRule: storecmn.PtrToNullString(recurrenceRule),
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
	WorkspaceID *string
	Priority    *models.TodoPriorityType
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
	if params.WorkspaceID != nil {
		sets = append(sets, "workspace_id = ?")
		args = append(args, *params.WorkspaceID)
	}
	if params.Priority != nil {
		sets = append(sets, "priority = ?")
		args = append(args, string(*params.Priority))
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

// Complete marks a todo as completed and spawns the next occurrence if recurring.
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

	// If this is a recurring todo, create the next occurrence
	if todo.RecurrenceRule.Valid && todo.RecurrenceRule.String != "" {
		nextDate := NextRecurrenceDate(todo.PlannedDate, todo.RecurrenceRule.String)
		next, err := s.store.Todos().Create(ctx, repo_todos.CreateParams{
			ID:                 utils.GenerateULID(),
			UserID:             userID,
			Title:              todo.Title,
			Description:        todo.Description,
			Status:             models.TodoStatusPending,
			PlannedDate:        nextDate,
			WorkspaceID:        todo.WorkspaceID,
			Priority:           todo.Priority,
			RecurrenceRule:     todo.RecurrenceRule,
			RecurrenceParentID: sql.NullString{String: todo.ID, Valid: true},
		})
		if err == nil {
			// Copy tags from completed todo to the new occurrence
			tags, tagErr := s.store.Tags().FindByTodoID(ctx, todo.ID)
			if tagErr == nil && len(tags) > 0 {
				tagIDs := make([]string, len(tags))
				for i, t := range tags {
					tagIDs[i] = t.ID
				}
				_ = s.store.Tags().SetTodoTags(ctx, next.ID, tagIDs)
			}
		}
	}

	s.broker.Publish(TodoEvent{UserID: userID})
	return todo, nil
}

// NextRecurrenceDate calculates the next planned date based on the recurrence rule.
func NextRecurrenceDate(from time.Time, rule string) time.Time {
	switch rule {
	case "daily":
		return from.AddDate(0, 0, 1)
	case "weekly":
		return from.AddDate(0, 0, 7)
	case "monthly":
		return from.AddDate(0, 1, 0)
	default:
		return from.AddDate(0, 0, 1)
	}
}

// BulkDelete deletes all todos matching the given filters for the user. UserID from
// the parameter overrides any value in params to prevent cross-user deletion. The
// underlying repo enforces that at least one narrowing filter is set.
func (s *Service) BulkDelete(ctx context.Context, userID int64, params repo_todos.FindParams) (int64, error) {
	params.UserID = userID
	count, err := s.store.Todos().DeleteWhere(ctx, params)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		s.broker.Publish(TodoEvent{UserID: userID})
	}
	return count, nil
}

// Delete deletes a todo
func (s *Service) Delete(ctx context.Context, userID int64, id string) error {
	if id == "" {
		return storecmn.ErrEmptyID
	}
	// Clear recurrence_parent_id on any child todos before deleting the parent
	// to avoid FK constraint failures.
	_, _ = s.store.SQLite().ExecContext(ctx,
		"UPDATE todos SET recurrence_parent_id = NULL WHERE recurrence_parent_id = ?", id,
	)
	if err := s.store.Todos().Delete(ctx, id); err != nil {
		return err
	}
	s.broker.Publish(TodoEvent{UserID: userID})
	return nil
}
