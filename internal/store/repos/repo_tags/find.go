package repo_tags

import (
	"context"
	"strings"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/huandu/go-sqlbuilder"
	"github.com/sxwebdev/donejournal/internal/models"
	"github.com/sxwebdev/donejournal/internal/store/storecmn"
)

type FindParams struct {
	UserID   int64
	Search   *string
	Page     *uint32
	PageSize *uint32
}

func findBuilder(params FindParams, col ...string) *sqlbuilder.SelectBuilder {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(col...)
	sb.From(TableNameTags.String()).
		Where(sb.Equal(ColumnNameTagsUserId.String(), params.UserID))

	if params.Search != nil && *params.Search != "" {
		pattern := "%" + strings.ToLower(*params.Search) + "%"
		nameExpr := "unicode_lower(" + ColumnNameTagsName.String() + ") LIKE " + sb.Var(pattern)
		sb.Where(nameExpr)
	}

	return sb
}

// Find tags by user ID with optional filters
func (s *CustomQueries) Find(ctx context.Context, params FindParams) (*storecmn.FindResponseWithCount[*models.Tag], error) {
	sb := findBuilder(params, TagsColumnNames().Strings()...)

	sb.OrderByDesc("created_at")

	limit, offset, err := storecmn.Pagination(params.Page, params.PageSize)
	if err != nil {
		return nil, err
	}
	sb.Limit(int(limit)).Offset(int(offset))

	items := []*models.Tag{}
	sql, args := sb.Build()
	if err := sqlscan.Select(ctx, s.db, &items, sql, args...); err != nil {
		return nil, err
	}

	var totalCount uint32
	countSQL, countArgs := findBuilder(params, "count(*)").Build()
	if err := sqlscan.Get(ctx, s.db, &totalCount, countSQL, countArgs...); err != nil {
		return nil, err
	}

	return &storecmn.FindResponseWithCount[*models.Tag]{
		Count: totalCount,
		Items: items,
	}, nil
}

// FindByName finds a tag by exact name for a user
func (s *CustomQueries) FindByName(ctx context.Context, userID int64, name string) (*models.Tag, error) {
	var tag models.Tag
	err := sqlscan.Get(ctx, s.db,
		&tag,
		"SELECT id, user_id, name, color, created_at, updated_at FROM tags WHERE user_id = ? AND name = ? LIMIT 1",
		userID, name,
	)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}
