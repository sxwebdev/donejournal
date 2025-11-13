-- name: Create :one
INSERT INTO requests (id, data, status, error_message, user_id)
	VALUES (?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM requests WHERE id=?;

-- name: GetByID :one
SELECT * FROM requests WHERE id=? LIMIT 1;

