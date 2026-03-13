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
	Statuses  []models.TodoStatusType
	UserID    int64
	DateFrom  *time.Time
	DateTo    *time.Time
	WorkspaceID *string
	OrderBy   string
	Page      *uint32
	PageSize  *uint32
}

func findBuilder(params FindParams, col ...string) *sqlbuilder.SelectBuilder {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(col...)
	sb.From(TableNameTodos.String()).
		Where(sb.Equal(ColumnNameTodosUserId.String(), params.UserID))

	if len(params.Statuses) > 0 {
		vals := make([]interface{}, len(params.Statuses))
		for i, s := range params.Statuses {
			vals[i] = s
		}
		sb.Where(sb.In(ColumnNameTodosStatus.String(), vals...))
	}

	// Use completed_at for date filtering only when filtering exclusively by completed status
	onlyCompleted := len(params.Statuses) == 1 && params.Statuses[0] == models.TodoStatusCompleted
	if onlyCompleted {
		if params.DateFrom != nil {
			sb.Where(sb.GreaterEqualThan(ColumnNameTodosCompletedAt.String(), *params.DateFrom))
		}
		if params.DateTo != nil {
			sb.Where(sb.LessEqualThan(ColumnNameTodosCompletedAt.String(), *params.DateTo))
		}
	} else {
		if params.DateFrom != nil {
			sb.Where(sb.GreaterEqualThan(ColumnNameTodosPlannedDate.String(), *params.DateFrom))
		}
		if params.DateTo != nil {
			sb.Where(sb.LessEqualThan(ColumnNameTodosPlannedDate.String(), *params.DateTo))
		}
	}

	if params.WorkspaceID != nil {
		sb.Where(sb.Equal(ColumnNameTodosWorkspaceId.String(), *params.WorkspaceID))
	}

	return sb
}

// Count returns the number of todos matching the given filters.
func (s *CustomQueries) Count(ctx context.Context, params FindParams) (uint32, error) {
	var count uint32
	sql, args := findBuilder(params, "count(*)").Build()
	if err := sqlscan.Get(ctx, s.db, &count, sql, args...); err != nil {
		return 0, err
	}
	return count, nil
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
