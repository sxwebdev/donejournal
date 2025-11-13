package requests

import (
	"context"
	"fmt"

	"github.com/sxwebdev/donejournal/internal/models"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_requests"
	"github.com/sxwebdev/donejournal/internal/store/storecmn"
	"github.com/sxwebdev/donejournal/pkg/utils"
)

// Create a new request
func (s *Service) Create(ctx context.Context, data, userID string) (*models.Request, error) {
	if data == "" {
		return nil, fmt.Errorf("empty message")
	}

	if len(data) > 200 {
		return nil, fmt.Errorf("message too long")
	}

	if userID == "" {
		return nil, fmt.Errorf("empty user ID")
	}

	req := repo_requests.CreateParams{
		ID:     utils.GenerateULID(),
		Data:   data,
		Status: models.RequestStatusPending,
		UserID: userID,
	}

	return s.store.Requests().Create(ctx, req)
}

// GetPendingRequests retrieves pending requests
func (s *Service) GetPendingRequests(ctx context.Context) ([]*models.Request, error) {
	return s.store.Requests().GetPendingRequests(ctx)
}

// UpdateStatus updates the status of a request
func (s *Service) UpdateStatus(ctx context.Context, id string, status models.RequestStatusType, errorMessage *string) error {
	if id == "" {
		return storecmn.ErrEmptyID
	}

	if err := status.Validate(); err != nil {
		return err
	}

	if errorMessage != nil && *errorMessage == "" {
		errorMessage = nil
	}

	return s.store.Requests().UpdateStatus(ctx, repo_requests.UpdateStatusParams{
		ID:           id,
		Status:       status,
		ErrorMessage: errorMessage,
	})
}
