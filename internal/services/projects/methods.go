package projects

import (
	"context"
	"database/sql"
	"errors"

	"github.com/sxwebdev/donejournal/internal/models"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_projects"
	"github.com/sxwebdev/donejournal/internal/store/storecmn"
	"github.com/sxwebdev/donejournal/pkg/utils"
)

// Create a new project
func (s *Service) Create(ctx context.Context, userID int64, name, description string) (*models.Project, error) {
	project, err := s.store.Projects().Create(ctx, repo_projects.CreateParams{
		ID:          utils.GenerateULID(),
		UserID:      userID,
		Name:        name,
		Description: description,
	})
	if err != nil {
		return nil, err
	}
	s.broker.Publish(ProjectEvent{UserID: userID})
	return project, nil
}

// Find projects by params
func (s *Service) Find(ctx context.Context, params repo_projects.FindParams) (*storecmn.FindResponseWithCount[*models.Project], error) {
	return s.store.Projects().Find(ctx, params)
}

// GetByID returns a single project
func (s *Service) GetByID(ctx context.Context, id string) (*models.Project, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}
	return s.store.Projects().GetByID(ctx, id)
}

// GetStats returns project with stats
func (s *Service) GetStats(ctx context.Context, projectID string, userID int64) (*repo_projects.ProjectStats, error) {
	if projectID == "" {
		return nil, storecmn.ErrEmptyID
	}
	return s.store.Projects().GetStats(ctx, projectID, userID)
}

// UpdateParams contains optional fields for partial update
type UpdateParams struct {
	Name        *string
	Description *string
}

// Update performs a partial update on a project
func (s *Service) Update(ctx context.Context, userID int64, id string, params UpdateParams) (*models.Project, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}

	sets := []string{}
	args := []any{}

	if params.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *params.Name)
	}
	if params.Description != nil {
		sets = append(sets, "description = ?")
		args = append(args, *params.Description)
	}

	if len(sets) == 0 {
		return s.store.Projects().GetByID(ctx, id)
	}

	query := "UPDATE projects SET "
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

	project, err := s.store.Projects().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.broker.Publish(ProjectEvent{UserID: userID})
	return project, nil
}

// Archive marks a project as archived
func (s *Service) Archive(ctx context.Context, userID int64, id string) (*models.Project, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}

	_, err := s.store.SQLite().ExecContext(ctx,
		"UPDATE projects SET archived = TRUE, updated_at = CURRENT_TIMESTAMP WHERE id = ?", id,
	)
	if err != nil {
		return nil, err
	}

	project, err := s.store.Projects().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.broker.Publish(ProjectEvent{UserID: userID})
	return project, nil
}

// Unarchive marks a project as not archived
func (s *Service) Unarchive(ctx context.Context, userID int64, id string) (*models.Project, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}

	_, err := s.store.SQLite().ExecContext(ctx,
		"UPDATE projects SET archived = FALSE, updated_at = CURRENT_TIMESTAMP WHERE id = ?", id,
	)
	if err != nil {
		return nil, err
	}

	project, err := s.store.Projects().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.broker.Publish(ProjectEvent{UserID: userID})
	return project, nil
}

// Delete deletes a project
func (s *Service) Delete(ctx context.Context, userID int64, id string) error {
	if id == "" {
		return storecmn.ErrEmptyID
	}
	if err := s.store.Projects().Delete(ctx, id); err != nil {
		return err
	}
	s.broker.Publish(ProjectEvent{UserID: userID})
	return nil
}

// FindOrCreateByName finds a project by name, creating it if it doesn't exist
func (s *Service) FindOrCreateByName(ctx context.Context, userID int64, name string) (*models.Project, error) {
	project, err := s.store.Projects().FindByName(ctx, userID, name)
	if err == nil {
		return project, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// Project doesn't exist, create it
	return s.Create(ctx, userID, name, "")
}
