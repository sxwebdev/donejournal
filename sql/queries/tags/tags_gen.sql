-- name: Create :one
INSERT INTO tags (id, user_id, name, color)
	VALUES (?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM tags WHERE id=?;

-- name: GetByID :one
SELECT * FROM tags WHERE id=? LIMIT 1;

