package repo_projects

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
	sb.From(TableNameProjects.String()).
		Where(sb.Equal(ColumnNameProjectsUserId.String(), params.UserID))

	if !params.IncludeArchived {
		sb.Where(sb.Equal(ColumnNameProjectsArchived.String(), false))
	}

	if params.Search != nil && *params.Search != "" {
		pattern := "%" + strings.ToLower(*params.Search) + "%"
		nameExpr := "unicode_lower(" + ColumnNameProjectsName.String() + ") LIKE " + sb.Var(pattern)
		sb.Where(nameExpr)
	}

	return sb
}

// Find projects by user ID with optional filters
func (s *CustomQueries) Find(ctx context.Context, params FindParams) (*storecmn.FindResponseWithCount[*models.Project], error) {
	sb := findBuilder(params, ProjectsColumnNames().Strings()...)

	sb.OrderByDesc("created_at")

	limit, offset, err := storecmn.Pagination(params.Page, params.PageSize)
	if err != nil {
		return nil, err
	}
	sb.Limit(int(limit)).Offset(int(offset))

	items := []*models.Project{}
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

	return &storecmn.FindResponseWithCount[*models.Project]{
		Count: totalCount,
		Items: items,
	}, nil
}

// FindByName finds a project by exact name for a user
func (s *CustomQueries) FindByName(ctx context.Context, userID int64, name string) (*models.Project, error) {
	var project models.Project
	err := sqlscan.Get(ctx, s.db,
		&project,
		"SELECT id, user_id, name, description, archived, created_at, updated_at FROM projects WHERE user_id = ? AND name = ? LIMIT 1",
		userID, name,
	)
	if err != nil {
		return nil, err
	}
	return &project, nil
}
