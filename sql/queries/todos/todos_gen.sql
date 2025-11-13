-- name: Create :one
INSERT INTO todos (id, user_id, title, description, status, planned_date, request_id)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	RETURNING *;

-- name: Delete :exec
DELETE FROM todos WHERE id=?;

-- name: GetByID :one
SELECT * FROM todos WHERE id=? LIMIT 1;

