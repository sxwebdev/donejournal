package notes

import (
	"context"

	"github.com/sxwebdev/donejournal/internal/models"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_notes"
	"github.com/sxwebdev/donejournal/internal/store/storecmn"
	"github.com/sxwebdev/donejournal/pkg/utils"
)

// Create a new note
func (s *Service) Create(ctx context.Context, userID int64, title, body string, projectID *string) (*models.Note, error) {
	note, err := s.store.Notes().Create(ctx, repo_notes.CreateParams{
		ID:        utils.GenerateULID(),
		UserID:    userID,
		Title:     title,
		Body:      body,
		ProjectID: projectID,
	})
	if err != nil {
		return nil, err
	}
	s.broker.Publish(NoteEvent{UserID: userID})
	return note, nil
}

// Find notes by params
func (s *Service) Find(ctx context.Context, params repo_notes.FindParams) (*storecmn.FindResponseWithCount[*models.Note], error) {
	return s.store.Notes().Find(ctx, params)
}

// GetByID returns a single note
func (s *Service) GetByID(ctx context.Context, id string) (*models.Note, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}
	return s.store.Notes().GetByID(ctx, id)
}

// UpdateParams contains optional fields for partial update
type UpdateParams struct {
	Title     *string
	Body      *string
	ProjectID *string
}

// Update performs a partial update on a note
func (s *Service) Update(ctx context.Context, userID int64, id string, params UpdateParams) (*models.Note, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}

	sets := []string{}
	args := []any{}

	if params.Title != nil {
		sets = append(sets, "title = ?")
		args = append(args, *params.Title)
	}
	if params.Body != nil {
		sets = append(sets, "body = ?")
		args = append(args, *params.Body)
	}
	if params.ProjectID != nil {
		sets = append(sets, "project_id = ?")
		args = append(args, *params.ProjectID)
	}

	if len(sets) == 0 {
		return s.store.Notes().GetByID(ctx, id)
	}

	query := "UPDATE notes SET "
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

	note, err := s.store.Notes().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.broker.Publish(NoteEvent{UserID: userID})
	return note, nil
}

// Delete deletes a note
func (s *Service) Delete(ctx context.Context, userID int64, id string) error {
	if id == "" {
		return storecmn.ErrEmptyID
	}
	if err := s.store.Notes().Delete(ctx, id); err != nil {
		return err
	}
	s.broker.Publish(NoteEvent{UserID: userID})
	return nil
}
