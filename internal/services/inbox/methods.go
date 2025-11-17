package inbox

import (
	"context"
	"fmt"

	"github.com/sxwebdev/donejournal/internal/models"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_inbox"
	"github.com/sxwebdev/donejournal/pkg/utils"
)

// Create a new request
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
