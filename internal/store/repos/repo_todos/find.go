package repo_todos

import (
	"context"
	"time"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/huandu/go-sqlbuilder"
	"github.com/sxwebdev/donejournal/internal/models"
	"github.com/sxwebdev/donejournal/internal/store/storecmn"
)

type FindParams struct {
	IsCompleted *bool
	UserID      int64
	DateFrom    *time.Time
	DateTo      *time.Time
	OrderBy     string
	Page        *uint32
	PageSize    *uint32
}

func findBuilder(params FindParams, col ...string) *sqlbuilder.SelectBuilder {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(col...)
	sb.From(TableNameTodos.String()).
		Where(sb.Equal(ColumnNameTodosUserId.String(), params.UserID))

	if params.IsCompleted != nil {
		if *params.IsCompleted {
			sb.Where(sb.EQ(ColumnNameTodosStatus.String(), models.TodoStatusCompleted))

			// For completed tasks, filter by completed_at date
			if params.DateFrom != nil {
				sb.Where(sb.GreaterEqualThan(ColumnNameTodosCompletedAt.String(), *params.DateFrom))
			}
			if params.DateTo != nil {
				sb.Where(sb.LessEqualThan(ColumnNameTodosCompletedAt.String(), *params.DateTo))
			}
		} else {
			sb.Where(
				sb.In(
					ColumnNameTodosStatus.String(),
					models.TodoStatusPending,
					models.TodoStatusInProgress,
				),
			)

			// For pending/in-progress tasks, filter by planned_date
			if params.DateFrom != nil {
				sb.Where(sb.GreaterEqualThan(ColumnNameTodosPlannedDate.String(), *params.DateFrom))
			}
			if params.DateTo != nil {
				sb.Where(sb.LessEqualThan(ColumnNameTodosPlannedDate.String(), *params.DateTo))
			}
		}
	} else {
		// If IsCompleted is not set, use planned_date for filtering
		if params.DateFrom != nil {
			sb.Where(sb.GreaterEqualThan(ColumnNameTodosPlannedDate.String(), *params.DateFrom))
		}
		if params.DateTo != nil {
			sb.Where(sb.LessEqualThan(ColumnNameTodosPlannedDate.String(), *params.DateTo))
		}
	}

	return sb
}

// Find todos by user ID and status
func (s *CustomQueries) Find(ctx context.Context, params FindParams) (*storecmn.FindResponseWithCount[*models.Todo], error) {
	sb := findBuilder(params, TodosColumnNames().Strings()...)

	if params.OrderBy == "" {
		params.OrderBy = "created_at"
	}

	sb.OrderByDesc(params.OrderBy)

	limit, offset, err := storecmn.Pagination(params.Page, params.PageSize)
	if err != nil {
		return nil, err
	}
	sb.Limit(int(limit)).Offset(int(offset))

	items := []*models.Todo{}
	sql, args := sb.Build()
	if err := sqlscan.Select(ctx, s.db, &items, sql, args...); err != nil {
		return nil, err
	}

	// Get total count of services
	var totalCount uint32
	countSQL, countArgs := findBuilder(params, "count(*)").Build()
	if err := sqlscan.Get(ctx, s.db, &totalCount, countSQL, countArgs...); err != nil {
		return nil, err
	}

	return &storecmn.FindResponseWithCount[*models.Todo]{
		Count: totalCount,
		Items: items,
	}, nil
}
