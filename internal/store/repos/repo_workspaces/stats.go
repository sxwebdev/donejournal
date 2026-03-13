package repo_workspaces

import (
	"context"

	"github.com/georgysavva/scany/v2/sqlscan"
)

// GetStats returns workspace stats with todo and note counts
func (s *CustomQueries) GetStats(ctx context.Context, workspaceID string, userID int64) (*WorkspaceStats, error) {
	query := `
		SELECT
			w.id, w.user_id, w.name, w.description, w.archived, w.created_at, w.updated_at,
			(SELECT COUNT(*) FROM todos WHERE workspace_id = w.id) as todo_count,
			(SELECT COUNT(*) FROM notes WHERE workspace_id = w.id) as note_count,
			(SELECT COUNT(*) FROM todos WHERE workspace_id = w.id AND status = 'completed') as completed_todo_count
		FROM workspaces w
		WHERE w.id = ? AND w.user_id = ?
		LIMIT 1
	`

	var stats WorkspaceStats
	if err := sqlscan.Get(ctx, s.db, &stats, query, workspaceID, userID); err != nil {
		return nil, err
	}
	return &stats, nil
}
