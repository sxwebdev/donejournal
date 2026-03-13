package repo_workspaces

import (
	"context"
	"strings"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/huandu/go-sqlbuilder"
	"github.com/sxwebdev/donejournal/internal/models"
	"github.com/sxwebdev/donejournal/internal/store/storecmn"
)

type FindParams struct {
	UserID          int64
	Search          *string
	IncludeArchived bool
	Page            *uint32
	PageSize        *uint32
}

func findBuilder(params FindParams, col ...string) *sqlbuilder.SelectBuilder {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(col...)
	sb.From(TableNameWorkspaces.String()).
		Where(sb.Equal(ColumnNameWorkspacesUserId.String(), params.UserID))

	if !params.IncludeArchived {
		sb.Where(sb.Equal(ColumnNameWorkspacesArchived.String(), false))
	}

	if params.Search != nil && *params.Search != "" {
		pattern := "%" + strings.ToLower(*params.Search) + "%"
		nameExpr := "unicode_lower(" + ColumnNameWorkspacesName.String() + ") LIKE " + sb.Var(pattern)
		sb.Where(nameExpr)
	}

	return sb
}

// Find workspaces by user ID with optional filters
func (s *CustomQueries) Find(ctx context.Context, params FindParams) (*storecmn.FindResponseWithCount[*models.Workspace], error) {
	sb := findBuilder(params, WorkspacesColumnNames().Strings()...)

	sb.OrderByDesc("created_at")

	limit, offset, err := storecmn.Pagination(params.Page, params.PageSize)
	if err != nil {
		return nil, err
	}
	sb.Limit(int(limit)).Offset(int(offset))

	items := []*models.Workspace{}
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

	return &storecmn.FindResponseWithCount[*models.Workspace]{
		Count: totalCount,
		Items: items,
	}, nil
}

// FindByName finds a workspace by exact name for a user
func (s *CustomQueries) FindByName(ctx context.Context, userID int64, name string) (*models.Workspace, error) {
	var workspace models.Workspace
	err := sqlscan.Get(ctx, s.db,
		&workspace,
		"SELECT id, user_id, name, description, archived, created_at, updated_at FROM workspaces WHERE user_id = ? AND name = ? LIMIT 1",
		userID, name,
	)
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}
