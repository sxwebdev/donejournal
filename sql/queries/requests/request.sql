-- name: UpdateStatus :exec
UPDATE requests
  SET "status" = ?, error_message = ?, updated_at = CURRENT_TIMESTAMP
  WHERE id = ?;

-- name: GetPendingRequests :many
SELECT *
  FROM requests
  WHERE "status" = 'pending'
  ORDER BY created_at ASC LIMIT 3;
