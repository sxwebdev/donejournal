package tags

import (
	"context"
	"database/sql"
	"errors"

	"github.com/sxwebdev/donejournal/internal/models"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_tags"
	"github.com/sxwebdev/donejournal/internal/store/storecmn"
	"github.com/sxwebdev/donejournal/pkg/utils"
)

// Create a new tag
func (s *Service) Create(ctx context.Context, userID int64, name, color string) (*models.Tag, error) {
	if color == "" {
		color = "#6366f1"
	}
	tag, err := s.store.Tags().Create(ctx, repo_tags.CreateParams{
		ID:     utils.GenerateULID(),
		UserID: userID,
		Name:   name,
		Color:  color,
	})
	if err != nil {
		return nil, err
	}
	s.broker.Publish(TagEvent{UserID: userID})
	return tag, nil
}

// Find tags by params
func (s *Service) Find(ctx context.Context, params repo_tags.FindParams) (*storecmn.FindResponseWithCount[*models.Tag], error) {
	return s.store.Tags().Find(ctx, params)
}

// GetByID returns a single tag
func (s *Service) GetByID(ctx context.Context, id string) (*models.Tag, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}
	return s.store.Tags().GetByID(ctx, id)
}

// UpdateParams contains optional fields for partial update
type UpdateParams struct {
	Name  *string
	Color *string
}

// Update performs a partial update on a tag
func (s *Service) Update(ctx context.Context, userID int64, id string, params UpdateParams) (*models.Tag, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}

	sets := []string{}
	args := []any{}

	if params.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *params.Name)
	}
	if params.Color != nil {
		sets = append(sets, "color = ?")
		args = append(args, *params.Color)
	}

	if len(sets) == 0 {
		return s.store.Tags().GetByID(ctx, id)
	}

	query := "UPDATE tags SET "
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

	tag, err := s.store.Tags().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.broker.Publish(TagEvent{UserID: userID})
	return tag, nil
}

// Delete deletes a tag
func (s *Service) Delete(ctx context.Context, userID int64, id string) error {
	if id == "" {
		return storecmn.ErrEmptyID
	}
	if err := s.store.Tags().Delete(ctx, id); err != nil {
		return err
	}
	s.broker.Publish(TagEvent{UserID: userID})
	return nil
}

// FindOrCreateByName finds a tag by name, creating it if it doesn't exist
func (s *Service) FindOrCreateByName(ctx context.Context, userID int64, name string) (*models.Tag, error) {
	tag, err := s.store.Tags().FindByName(ctx, userID, name)
	if err == nil {
		return tag, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return s.Create(ctx, userID, name, "")
}

// FindByTodoID returns all tags for a todo
func (s *Service) FindByTodoID(ctx context.Context, todoID string) ([]*models.Tag, error) {
	return s.store.Tags().FindByTodoID(ctx, todoID)
}

// FindByNoteID returns all tags for a note
func (s *Service) FindByNoteID(ctx context.Context, noteID string) ([]*models.Tag, error) {
	return s.store.Tags().FindByNoteID(ctx, noteID)
}

// FindTagIDsByTodoIDs batch loads tag IDs for multiple todos
func (s *Service) FindTagIDsByTodoIDs(ctx context.Context, todoIDs []string) (map[string][]string, error) {
	return s.store.Tags().FindTagIDsByTodoIDs(ctx, todoIDs)
}

// FindTagIDsByNoteIDs batch loads tag IDs for multiple notes
func (s *Service) FindTagIDsByNoteIDs(ctx context.Context, noteIDs []string) (map[string][]string, error) {
	return s.store.Tags().FindTagIDsByNoteIDs(ctx, noteIDs)
}

// SetTodoTags replaces all tags for a todo
func (s *Service) SetTodoTags(ctx context.Context, userID int64, todoID string, tagIDs []string) error {
	if err := s.store.Tags().SetTodoTags(ctx, todoID, tagIDs); err != nil {
		return err
	}
	s.broker.Publish(TagEvent{UserID: userID})
	return nil
}

// SetNoteTags replaces all tags for a note
func (s *Service) SetNoteTags(ctx context.Context, userID int64, noteID string, tagIDs []string) error {
	if err := s.store.Tags().SetNoteTags(ctx, noteID, tagIDs); err != nil {
		return err
	}
	s.broker.Publish(TagEvent{UserID: userID})
	return nil
}
