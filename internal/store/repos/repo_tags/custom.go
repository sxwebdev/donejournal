package repo_tags

import (
	"context"
	"database/sql"

	"github.com/sxwebdev/donejournal/internal/models"
	"github.com/sxwebdev/donejournal/internal/store/storecmn"
)

type ICustomQuerier interface {
	Querier
	Find(ctx context.Context, params FindParams) (*storecmn.FindResponseWithCount[*models.Tag], error)
	FindByName(ctx context.Context, userID int64, name string) (*models.Tag, error)
	FindByTodoID(ctx context.Context, todoID string) ([]*models.Tag, error)
	FindByNoteID(ctx context.Context, noteID string) ([]*models.Tag, error)
	FindTagIDsByTodoIDs(ctx context.Context, todoIDs []string) (map[string][]string, error)
	FindTagIDsByNoteIDs(ctx context.Context, noteIDs []string) (map[string][]string, error)
	SetTodoTags(ctx context.Context, todoID string, tagIDs []string) error
	SetNoteTags(ctx context.Context, noteID string, tagIDs []string) error
}

type CustomQueries struct {
	*Queries
	db DBTX
}

func NewCustom(db DBTX) *CustomQueries {
	return &CustomQueries{
		Queries: New(db),
		db:      db,
	}
}

func (s *CustomQueries) WithTx(tx *sql.Tx) *CustomQueries {
	return &CustomQueries{
		Queries: New(tx),
		db:      tx,
	}
}

var _ ICustomQuerier = (*CustomQueries)(nil)
