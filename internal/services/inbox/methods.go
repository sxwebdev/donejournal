package inbox

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sxwebdev/donejournal/internal/models"
	"github.com/sxwebdev/donejournal/internal/store/repos"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_inbox"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_todos"
	"github.com/sxwebdev/donejournal/internal/store/storecmn"
	"github.com/sxwebdev/donejournal/pkg/utils"
)

// Create a new inbox item
func (s *Service) Create(ctx context.Context, data, userID string) (*models.Inbox, error) {
	if data == "" {
		return nil, fmt.Errorf("empty message")
	}

	if len(data) > 200 {
		return nil, fmt.Errorf("message too long")
	}

	if userID == "" {
		return nil, fmt.Errorf("empty user ID")
	}

	req := repo_inbox.CreateParams{
		ID:     utils.GenerateULID(),
		Data:   data,
		UserID: userID,
	}

	return s.store.Inbox().Create(ctx, req)
}

// List returns paginated inbox items for a user
func (s *Service) List(ctx context.Context, userID string, page, pageSize *uint32) (*storecmn.FindResponseWithCount[*models.Inbox], error) {
	return s.store.Inbox().Find(ctx, repo_inbox.FindParams{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetByID returns a single inbox item
func (s *Service) GetByID(ctx context.Context, id string) (*models.Inbox, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}
	return s.store.Inbox().GetByID(ctx, id)
}

// Update updates an inbox item
func (s *Service) Update(ctx context.Context, id, data, additionalData string) (*models.Inbox, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}

	_, err := s.store.SQLite().ExecContext(ctx,
		"UPDATE inbox SET data = ?, additional_data = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		data, additionalData, id,
	)
	if err != nil {
		return nil, err
	}

	return s.store.Inbox().GetByID(ctx, id)
}

// Delete deletes an inbox item
func (s *Service) Delete(ctx context.Context, id string) error {
	if id == "" {
		return storecmn.ErrEmptyID
	}
	return s.store.Inbox().Delete(ctx, id)
}

// ConvertToTodo converts an inbox item to a todo and deletes the inbox item
func (s *Service) ConvertToTodo(ctx context.Context, inboxItemID string, userID int64, title, description string, plannedDate time.Time) (string, error) {
	if inboxItemID == "" {
		return "", storecmn.ErrEmptyID
	}

	item, err := s.store.Inbox().GetByID(ctx, inboxItemID)
	if err != nil {
		return "", fmt.Errorf("inbox item not found: %w", err)
	}

	if title == "" {
		title = item.Data
	}

	todoID := utils.GenerateULID()

	err = storecmn.WrapTx(ctx, s.store.SQLite(), func(tx *sql.Tx) error {
		_, err := s.store.Todos(repos.WithTx(tx)).Create(ctx, repo_todos.CreateParams{
			ID:          todoID,
			UserID:      userID,
			Title:       title,
			Description: description,
			Status:      models.TodoStatusPending,
			PlannedDate: plannedDate,
		})
		if err != nil {
			return fmt.Errorf("create todo: %w", err)
		}

		if err := s.store.Inbox(repos.WithTx(tx)).Delete(ctx, inboxItemID); err != nil {
			return fmt.Errorf("delete inbox item: %w", err)
		}

		return nil
	})

	return todoID, err
}
