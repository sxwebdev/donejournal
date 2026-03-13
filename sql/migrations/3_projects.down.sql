-- Remove project_id from todos (SQLite requires table recreation)
CREATE TABLE todos_backup AS SELECT id, user_id, title, description, status, planned_date, completed_at, created_at, updated_at FROM todos;
DROP TABLE todos;
CREATE TABLE todos (
  id TEXT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  title TEXT NOT NULL,
  description TEXT NOT NULL,
  status TEXT NOT NULL,
  planned_date DATETIME NOT NULL,
  completed_at DATETIME,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO todos SELECT * FROM todos_backup;
DROP TABLE todos_backup;
CREATE INDEX IF NOT EXISTS idx_todos_user_id ON todos(user_id);
CREATE INDEX IF NOT EXISTS idx_todos_status ON todos(status);

-- Remove project_id from notes (SQLite requires table recreation)
CREATE TABLE notes_backup AS SELECT id, user_id, title, body, created_at, updated_at FROM notes;
DROP TABLE notes;
CREATE TABLE notes (
  id TEXT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  title TEXT NOT NULL,
  body TEXT NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO notes SELECT * FROM notes_backup;
DROP TABLE notes_backup;
CREATE INDEX IF NOT EXISTS idx_notes_user_id ON notes(user_id);

-- Drop projects table
DROP TABLE IF EXISTS projects;
