-- name: Create :one
INSERT INTO projects (id, user_id, name, description, archived)
	VALUES (?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM projects WHERE id=?;

-- name: GetByID :one
SELECT * FROM projects WHERE id=? LIMIT 1;

