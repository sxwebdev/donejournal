-- name: UpdateStatus :exec
UPDATE todos
  SET status = ?, updated_at = CURRENT_TIMESTAMP
  WHERE id = ?
  RETURNING *;
