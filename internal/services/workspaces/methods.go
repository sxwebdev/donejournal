package workspaces

import (
	"context"
	"database/sql"
	"errors"

	"github.com/sxwebdev/donejournal/internal/models"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_workspaces"
	"github.com/sxwebdev/donejournal/internal/store/storecmn"
	"github.com/sxwebdev/donejournal/pkg/utils"
)

// Create a new workspace
func (s *Service) Create(ctx context.Context, userID int64, name, description string) (*models.Workspace, error) {
	workspace, err := s.store.Workspaces().Create(ctx, repo_workspaces.CreateParams{
		ID:          utils.GenerateULID(),
		UserID:      userID,
		Name:        name,
		Description: description,
	})
	if err != nil {
		return nil, err
	}
	s.broker.Publish(WorkspaceEvent{UserID: userID})
	return workspace, nil
}

// Find workspaces by params
func (s *Service) Find(ctx context.Context, params repo_workspaces.FindParams) (*storecmn.FindResponseWithCount[*models.Workspace], error) {
	return s.store.Workspaces().Find(ctx, params)
}

// GetByID returns a single workspace
func (s *Service) GetByID(ctx context.Context, id string) (*models.Workspace, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}
	return s.store.Workspaces().GetByID(ctx, id)
}

// GetStats returns workspace with stats
func (s *Service) GetStats(ctx context.Context, workspaceID string, userID int64) (*repo_workspaces.WorkspaceStats, error) {
	if workspaceID == "" {
		return nil, storecmn.ErrEmptyID
	}
	return s.store.Workspaces().GetStats(ctx, workspaceID, userID)
}

// UpdateParams contains optional fields for partial update
type UpdateParams struct {
	Name        *string
	Description *string
}

// Update performs a partial update on a workspace
func (s *Service) Update(ctx context.Context, userID int64, id string, params UpdateParams) (*models.Workspace, error) {
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
		return s.store.Workspaces().GetByID(ctx, id)
	}

	query := "UPDATE workspaces SET "
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

	workspace, err := s.store.Workspaces().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.broker.Publish(WorkspaceEvent{UserID: userID})
	return workspace, nil
}

// Archive marks a workspace as archived
func (s *Service) Archive(ctx context.Context, userID int64, id string) (*models.Workspace, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}

	_, err := s.store.SQLite().ExecContext(ctx,
		"UPDATE workspaces SET archived = TRUE, updated_at = CURRENT_TIMESTAMP WHERE id = ?", id,
	)
	if err != nil {
		return nil, err
	}

	workspace, err := s.store.Workspaces().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.broker.Publish(WorkspaceEvent{UserID: userID})
	return workspace, nil
}

// Unarchive marks a workspace as not archived
func (s *Service) Unarchive(ctx context.Context, userID int64, id string) (*models.Workspace, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}

	_, err := s.store.SQLite().ExecContext(ctx,
		"UPDATE workspaces SET archived = FALSE, updated_at = CURRENT_TIMESTAMP WHERE id = ?", id,
	)
	if err != nil {
		return nil, err
	}

	workspace, err := s.store.Workspaces().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.broker.Publish(WorkspaceEvent{UserID: userID})
	return workspace, nil
}

// Delete deletes a workspace
func (s *Service) Delete(ctx context.Context, userID int64, id string) error {
	if id == "" {
		return storecmn.ErrEmptyID
	}
	if err := s.store.Workspaces().Delete(ctx, id); err != nil {
		return err
	}
	s.broker.Publish(WorkspaceEvent{UserID: userID})
	return nil
}

// FindOrCreateByName finds a workspace by name, creating it if it doesn't exist
func (s *Service) FindOrCreateByName(ctx context.Context, userID int64, name string) (*models.Workspace, error) {
	workspace, err := s.store.Workspaces().FindByName(ctx, userID, name)
	if err == nil {
		return workspace, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// Workspace doesn't exist, create it
	return s.Create(ctx, userID, name, "")
}
