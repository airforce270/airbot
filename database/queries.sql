-- name: SelectUserPoints :one
SELECT CAST(COALESCE(SUM(delta)) AS INTEGER) AS points
FROM gamba_transactions
WHERE user_id = ?;

-- name: CreateTwitchUser :exec
INSERT INTO users (
  created_at, updated_at,
  twitch_id, twitch_name
) VALUES (
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP,
  ?, ?
);

-- name: SelectTwitchUser :one
SELECT * 
FROM users 
WHERE twitch_id = ? AND twitch_name = ?;
