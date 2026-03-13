package repo_notes

import (
	"context"
	"strings"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/huandu/go-sqlbuilder"
	"github.com/sxwebdev/donejournal/internal/models"
	"github.com/sxwebdev/donejournal/internal/store/storecmn"
)

type FindParams struct {
	UserID    int64
	Search    *string
	ProjectID *string
	OrderBy   string
	Page      *uint32
	PageSize  *uint32
}

func findBuilder(params FindParams, col ...string) *sqlbuilder.SelectBuilder {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(col...)
	sb.From(TableNameNotes.String()).
		Where(sb.Equal(ColumnNameNotesUserId.String(), params.UserID))

	if params.Search != nil && *params.Search != "" {
		pattern := "%" + strings.ToLower(*params.Search) + "%"
		titleExpr := "unicode_lower(" + ColumnNameNotesTitle.String() + ") LIKE " + sb.Var(pattern)
		bodyExpr := "unicode_lower(" + ColumnNameNotesBody.String() + ") LIKE " + sb.Var(pattern)
		sb.Where(sb.Or(titleExpr, bodyExpr))
	}

	if params.ProjectID != nil {
		sb.Where(sb.Equal(ColumnNameNotesProjectId.String(), *params.ProjectID))
	}

	return sb
}

// Find notes by user ID with optional search
func (s *CustomQueries) Find(ctx context.Context, params FindParams) (*storecmn.FindResponseWithCount[*models.Note], error) {
	sb := findBuilder(params, NotesColumnNames().Strings()...)

	if params.OrderBy == "" {
		params.OrderBy = "updated_at"
	}

	sb.OrderByDesc(params.OrderBy)

	limit, offset, err := storecmn.Pagination(params.Page, params.PageSize)
	if err != nil {
		return nil, err
	}
	sb.Limit(int(limit)).Offset(int(offset))

	items := []*models.Note{}
	sql, args := sb.Build()
	if err := sqlscan.Select(ctx, s.db, &items, sql, args...); err != nil {
		return nil, err
	}

	// Get total count
	var totalCount uint32
	countSQL, countArgs := findBuilder(params, "count(*)").Build()
	if err := sqlscan.Get(ctx, s.db, &totalCount, countSQL, countArgs...); err != nil {
		return nil, err
	}

	return &storecmn.FindResponseWithCount[*models.Note]{
		Count: totalCount,
		Items: items,
	}, nil
}
