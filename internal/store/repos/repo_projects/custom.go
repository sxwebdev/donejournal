package repo_projects

import (
	"context"
	"database/sql"

	"github.com/sxwebdev/donejournal/internal/models"
	"github.com/sxwebdev/donejournal/internal/store/storecmn"
)

type ProjectStats struct {
	models.Project
	TodoCount          int32 `db:"todo_count" json:"todo_count"`
	NoteCount          int32 `db:"note_count" json:"note_count"`
	CompletedTodoCount int32 `db:"completed_todo_count" json:"completed_todo_count"`
}

type ICustomQuerier interface {
	Querier
	Find(ctx context.Context, params FindParams) (*storecmn.FindResponseWithCount[*models.Project], error)
	GetStats(ctx context.Context, projectID string, userID int64) (*ProjectStats, error)
	FindByName(ctx context.Context, userID int64, name string) (*models.Project, error)
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
