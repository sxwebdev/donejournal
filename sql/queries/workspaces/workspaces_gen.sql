-- name: Create :one
INSERT INTO workspaces (id, user_id, name, description, archived)
	VALUES (?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM workspaces WHERE id=?;

-- name: GetByID :one
SELECT * FROM workspaces WHERE id=? LIMIT 1;
