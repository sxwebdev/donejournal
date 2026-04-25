package repo_todos

import (
	"context"
	"database/sql"

	"github.com/sxwebdev/donejournal/internal/models"
	"github.com/sxwebdev/donejournal/internal/store/storecmn"
)

type ICustomQuerier interface {
	Querier
	Count(ctx context.Context, params FindParams) (uint32, error)
	Find(ctx context.Context, params FindParams) (*storecmn.FindResponseWithCount[*models.Todo], error)
	DeleteWhere(ctx context.Context, params FindParams) (int64, error)
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
