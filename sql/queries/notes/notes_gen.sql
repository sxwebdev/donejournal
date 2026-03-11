-- name: Create :one
INSERT INTO notes (id, user_id, title, body)
	VALUES (?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM notes WHERE id=?;

-- name: GetByID :one
SELECT * FROM notes WHERE id=? LIMIT 1;

