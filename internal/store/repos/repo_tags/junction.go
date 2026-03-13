package repo_tags

import (
	"context"
	"fmt"
	"strings"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/sxwebdev/donejournal/internal/models"
)

// FindByTodoID returns all tags associated with a todo.
func (s *CustomQueries) FindByTodoID(ctx context.Context, todoID string) ([]*models.Tag, error) {
	items := []*models.Tag{}
	err := sqlscan.Select(ctx, s.db, &items,
		"SELECT t.id, t.user_id, t.name, t.color, t.created_at, t.updated_at FROM tags t JOIN todo_tags tt ON t.id = tt.tag_id WHERE tt.todo_id = ?",
		todoID,
	)
	return items, err
}

// FindByNoteID returns all tags associated with a note.
func (s *CustomQueries) FindByNoteID(ctx context.Context, noteID string) ([]*models.Tag, error) {
	items := []*models.Tag{}
	err := sqlscan.Select(ctx, s.db, &items,
		"SELECT t.id, t.user_id, t.name, t.color, t.created_at, t.updated_at FROM tags t JOIN note_tags nt ON t.id = nt.tag_id WHERE nt.note_id = ?",
		noteID,
	)
	return items, err
}

// FindTagIDsByTodoIDs returns a map of todoID -> []tagID for batch loading.
func (s *CustomQueries) FindTagIDsByTodoIDs(ctx context.Context, todoIDs []string) (map[string][]string, error) {
	if len(todoIDs) == 0 {
		return map[string][]string{}, nil
	}

	placeholders := make([]string, len(todoIDs))
	args := make([]any, len(todoIDs))
	for i, id := range todoIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	type row struct {
		TodoID string `db:"todo_id"`
		TagID  string `db:"tag_id"`
	}

	rows := []row{}
	query := fmt.Sprintf("SELECT todo_id, tag_id FROM todo_tags WHERE todo_id IN (%s)", strings.Join(placeholders, ","))
	if err := sqlscan.Select(ctx, s.db, &rows, query, args...); err != nil {
		return nil, err
	}

	result := make(map[string][]string, len(todoIDs))
	for _, r := range rows {
		result[r.TodoID] = append(result[r.TodoID], r.TagID)
	}
	return result, nil
}

// FindTagIDsByNoteIDs returns a map of noteID -> []tagID for batch loading.
func (s *CustomQueries) FindTagIDsByNoteIDs(ctx context.Context, noteIDs []string) (map[string][]string, error) {
	if len(noteIDs) == 0 {
		return map[string][]string{}, nil
	}

	placeholders := make([]string, len(noteIDs))
	args := make([]any, len(noteIDs))
	for i, id := range noteIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	type row struct {
		NoteID string `db:"note_id"`
		TagID  string `db:"tag_id"`
	}

	rows := []row{}
	query := fmt.Sprintf("SELECT note_id, tag_id FROM note_tags WHERE note_id IN (%s)", strings.Join(placeholders, ","))
	if err := sqlscan.Select(ctx, s.db, &rows, query, args...); err != nil {
		return nil, err
	}

	result := make(map[string][]string, len(noteIDs))
	for _, r := range rows {
		result[r.NoteID] = append(result[r.NoteID], r.TagID)
	}
	return result, nil
}

// SetTodoTags replaces all tags for a todo.
func (s *CustomQueries) SetTodoTags(ctx context.Context, todoID string, tagIDs []string) error {
	if _, err := s.db.ExecContext(ctx, "DELETE FROM todo_tags WHERE todo_id = ?", todoID); err != nil {
		return err
	}

	for _, tagID := range tagIDs {
		if _, err := s.db.ExecContext(ctx, "INSERT OR IGNORE INTO todo_tags (todo_id, tag_id) VALUES (?, ?)", todoID, tagID); err != nil {
			return err
		}
	}
	return nil
}

// SetNoteTags replaces all tags for a note.
func (s *CustomQueries) SetNoteTags(ctx context.Context, noteID string, tagIDs []string) error {
	if _, err := s.db.ExecContext(ctx, "DELETE FROM note_tags WHERE note_id = ?", noteID); err != nil {
		return err
	}

	for _, tagID := range tagIDs {
		if _, err := s.db.ExecContext(ctx, "INSERT OR IGNORE INTO note_tags (note_id, tag_id) VALUES (?, ?)", noteID, tagID); err != nil {
			return err
		}
	}
	return nil
}
