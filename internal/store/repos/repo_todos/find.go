package repo_todos

import (
	"context"
	"errors"
	"time"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/huandu/go-sqlbuilder"
	"github.com/sxwebdev/donejournal/internal/models"
	"github.com/sxwebdev/donejournal/internal/store/storecmn"
)

// ErrBulkDeleteNoFilters is returned by DeleteWhere when no narrowing filter is set.
// UserID alone is not enough — that would wipe the entire user's todos.
var ErrBulkDeleteNoFilters = errors.New("bulk delete requires at least one filter (status / date / workspace / tags)")

type FindParams struct {
	Statuses          []models.TodoStatusType
	UserID            int64
	DateFrom          *time.Time
	DateTo            *time.Time
	WorkspaceID       *string
	TagIDs            []string
	OrderBy           string
	Page              *uint32
	PageSize          *uint32
	HasRecurrenceRule bool
	NoPagination      bool
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

	if params.HasRecurrenceRule {
		sb.Where("recurrence_rule IS NOT NULL AND recurrence_rule != ''")
	}

	if len(params.TagIDs) > 0 {
		tagVals := make([]interface{}, len(params.TagIDs))
		for i, id := range params.TagIDs {
			tagVals[i] = id
		}
		sub := sqlbuilder.NewSelectBuilder()
		sub.Select("todo_id").From("todo_tags").Where(sub.In("tag_id", tagVals...))
		sb.Where(sb.In("id", sub))
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

	if !params.NoPagination {
		limit, offset, err := storecmn.Pagination(params.Page, params.PageSize)
		if err != nil {
			return nil, err
		}
		sb.Limit(int(limit)).Offset(int(offset))
	}

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

// DeleteWhere bulk-deletes todos matching the given filters. UserID is always required;
// at least one additional filter (Statuses / DateFrom / DateTo / WorkspaceID / TagIDs) must
// be set, otherwise it returns ErrBulkDeleteNoFilters. Children pointing at deleted rows via
// recurrence_parent_id are nullified first, mirroring Service.Delete's FK fix-up.
// Returns the number of rows deleted.
func (s *CustomQueries) DeleteWhere(ctx context.Context, params FindParams) (int64, error) {
	if params.UserID == 0 {
		return 0, errors.New("UserID required for bulk delete")
	}
	if len(params.Statuses) == 0 && params.DateFrom == nil && params.DateTo == nil &&
		params.WorkspaceID == nil && len(params.TagIDs) == 0 {
		return 0, ErrBulkDeleteNoFilters
	}

	// Step 1: nullify recurrence_parent_id on any children referencing rows we are about to delete.
	// findBuilder is rebuilt for each statement because its bound args belong to that builder instance.
	nullify := sqlbuilder.NewUpdateBuilder()
	nullify.Update(TableNameTodos.String())
	nullify.Set(nullify.Assign(ColumnNameTodosRecurrenceParentId.String(), nil))
	nullify.Where(nullify.In(ColumnNameTodosRecurrenceParentId.String(), findBuilder(params, ColumnNameTodosId.String())))
	nullifySQL, nullifyArgs := nullify.Build()
	if _, err := s.db.ExecContext(ctx, nullifySQL, nullifyArgs...); err != nil {
		return 0, err
	}

	// Step 2: delete the matching rows.
	del := sqlbuilder.NewDeleteBuilder()
	del.DeleteFrom(TableNameTodos.String())
	del.Where(del.In(ColumnNameTodosId.String(), findBuilder(params, ColumnNameTodosId.String())))
	delSQL, delArgs := del.Build()
	res, err := s.db.ExecContext(ctx, delSQL, delArgs...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
