-- name: Create :one
INSERT INTO inbox (id, data, additional_data, user_id)
	VALUES (?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM inbox WHERE id=?;

-- name: GetByID :one
SELECT * FROM inbox WHERE id=? LIMIT 1;

