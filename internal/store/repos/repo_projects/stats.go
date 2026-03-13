package repo_projects

import (
	"context"

	"github.com/georgysavva/scany/v2/sqlscan"
)

// GetStats returns project stats with todo and note counts
func (s *CustomQueries) GetStats(ctx context.Context, projectID string, userID int64) (*ProjectStats, error) {
	query := `
		SELECT
			p.id, p.user_id, p.name, p.description, p.archived, p.created_at, p.updated_at,
			(SELECT COUNT(*) FROM todos WHERE project_id = p.id) as todo_count,
			(SELECT COUNT(*) FROM notes WHERE project_id = p.id) as note_count,
			(SELECT COUNT(*) FROM todos WHERE project_id = p.id AND status = 'completed') as completed_todo_count
		FROM projects p
		WHERE p.id = ? AND p.user_id = ?
		LIMIT 1
	`

	var stats ProjectStats
	if err := sqlscan.Get(ctx, s.db, &stats, query, projectID, userID); err != nil {
		return nil, err
	}
	return &stats, nil
}
