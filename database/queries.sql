-- todo: update selects to exclute deleted_at
-- todo: update updates to set updated_at

-- bot_bans

-- name: CreateBotBan :one 
INSERT INTO bot_bans (
  created_at, updated_at,
  platform, channel, banned_at
) VALUES (
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP,
  ?, ?, ?
)
RETURNING *;

-- name: CountBotBansInChannel :one
SELECT COUNT(*)
FROM bot_bans
WHERE platform = ?
  AND channel = ?
  AND banned_at > ?;


-- cache_bool_item

-- name: SelectCacheBoolItem :one
SELECT *
FROM cache_bool_items
WHERE keyy = ?;

-- name: UpsertCacheBoolItem :exec
INSERT INTO cache_bool_items
(keyy, value, expires_at)
VALUES (?, ?, ?)
ON CONFLICT (keyy)
DO UPDATE SET value = ?, expires_at = ?;


-- cache_string_items

-- name: SelectCacheStringItem :one
SELECT *
FROM cache_string_items
WHERE keyy = ?;

-- name: UpsertCacheStringItem :exec
INSERT INTO cache_string_items
(keyy, value, expires_at)
VALUES (?, ?, ?)
ON CONFLICT (keyy)
DO UPDATE SET value = ?, expires_at = ?;


-- channel_command_cooldowns

-- name: SelectChannelCommandCooldown :one
SELECT *
FROM channel_command_cooldowns
WHERE channel = ?
  AND command = ?;

-- name: CreateChannelCommandCooldown :one
INSERT INTO channel_command_cooldowns (
  created_at, updated_at,
  channel, command, last_run
) VALUES (
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP,
  ?, ?, ?
)
RETURNING *;

-- name: SetChannelCommandCooldownLastRun :exec
UPDATE channel_command_cooldowns
SET last_run = ?
WHERE channel = ?
  AND command = ?;


-- duels

-- name: CreateDuel :one
INSERT INTO duels (
  created_at, updated_at,
  user_id, target_id,
  amount,
  pending, accepted,
  won
) VALUES (
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP,
  ?, ?,
  ?,
  ?, ?,
  ?
)
RETURNING *;

-- name: SelectUserOutboundDuels :many
SELECT *
FROM duels
WHERE user_id = ? AND created_at >= ?;

-- name: SelectUserInboundDuels :many
SELECT *
FROM duels
WHERE target_id = ? AND created_at >= ?;

-- name: FinalizeDuel :exec
UPDATE duels
SET pending = ?, accepted = ?, won = ?
WHERE user_id = ?
  AND target_id = ?;

-- name: DeleteAllDuelsForTest :exec
DELETE FROM duels;


-- gamba_transactions

-- name: CreateGambaTransaction :one
INSERT INTO gamba_transactions (
  created_at, updated_at,
  user_id, game, delta
) VALUES (
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP,
  ?, ?, ?
)
RETURNING *;

-- name: SelectUserPoints :one
SELECT CAST(COALESCE(SUM(delta)) AS INTEGER)
FROM gamba_transactions
WHERE user_id = ?;

-- name: SelectAllGambaTransactions :many
SELECT *
FROM gamba_transactions;

-- name: DeleteAllGambaTransactionsForTest :exec
DELETE FROM gamba_transactions;


-- joined_channels

-- name: CreateJoinedChannel :one
INSERT INTO joined_channels (
  created_at, updated_at,
  platform,
  channel, channel_id,
  prefix, joined_at
) VALUES (
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP,
  ?,
  ?, ?,
  ?, ?
)
RETURNING *;

-- name: SelectJoinedChannel :one
SELECT *
FROM joined_channels
WHERE platform = ? AND channel = ?;

-- name: SelectJoinedChannels :many
SELECT *
FROM joined_channels
WHERE platform = ?;

-- name: CountJoinedChannels :one
SELECT COUNT(*)
FROM joined_channels;

-- name: UpdateJoinedChannelName :exec 
UPDATE joined_channels
SET channel = ?
WHERE channel_id = ?;

-- name: SetJoinedChannelPrefix :execrows
UPDATE joined_channels
SET prefix = ?
WHERE platform = ? AND channel = ?;

-- name: LeaveChannel :execrows
UPDATE joined_channels
SET deleted_at = CURRENT_TIMESTAMP
WHERE platform = ? AND channel = ?;


-- messages

-- name: CreateMessage :one
INSERT INTO messages (
  created_at, updated_at,
  user_id, channel, `text`, time
) VALUES (
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP,
  ?, ?, ?, ?
)
RETURNING *;

-- name: CountMessagesCreatedSince :one
SELECT COUNT(*)
FROM messages
WHERE created_at > ?;


-- user_command_cooldowns

-- name: SelectUserCommandCooldown :one
SELECT *
FROM user_command_cooldowns
WHERE user_id = ?
  AND command = ?;

-- name: CreateUserCommandCooldown :one
INSERT INTO user_command_cooldowns (
  created_at, updated_at,
  user_id, command, last_run
) VALUES (
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP,
  ?, ?, ?
)
RETURNING *;

-- name: SetUserCommandCooldownLastRun :exec
UPDATE user_command_cooldowns
SET last_run = ?
WHERE user_id = ?
  AND command = ?;


-- users

-- name: CreateTwitchUser :one
INSERT INTO users (
  created_at, updated_at,
  twitch_id, twitch_name
) VALUES (
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP,
  ?, ?
)
RETURNING *;

-- name: SelectUserByID :one
SELECT *
FROM users
WHERE id = ?;

-- name: SelectTwitchUser :one
SELECT *
FROM users
WHERE twitch_id = ? AND twitch_name = ?;

-- name: SelectTwitchUserByTwitchName :one
SELECT *
FROM users
WHERE twitch_name = ?;

-- name: SelectActiveUsers :many
SELECT *
FROM users
WHERE id IN (
    SELECT DISTINCT m.user_id
    FROM messages AS m
    WHERE m.time > ?
  );

-- name: UpdateTwitchUserName :one
UPDATE users
SET twitch_name = ?
WHERE twitch_id = ?
RETURNING *;

-- name: DeleteAllUsersForTest :exec
DELETE FROM users;
