-- bot_bans contains a record of the bot being banned from channels.
CREATE TABLE bot_bans (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME,
  -- platform contains the which platform this channel is on.
  platform TEXT,
  -- channel is which channel the bot was banned from.
  channel TEXT,
  -- banned_at is when the bot was banned.
  banned_at DATETIME
);
CREATE INDEX idx_bot_bans_deleted_at ON bot_bans (deleted_at);

-- cache_bool_items contains cache items of type bool.
CREATE TABLE cache_bool_items (
  created_at DATETIME,
  updated_at DATETIME,
  -- keyy is the item's key.
  -- Named keyy to avoid conflict within builtin identifier.
  keyy TEXT,
  -- value is the item's value.
  value NUMERIC,
  -- expires_at is when the item expires.
  -- If 0, the item never expires.
  expires_at DATETIME,
  PRIMARY KEY (key)
);

-- cache_string_items contains cache items of type string.
CREATE TABLE cache_string_items (
  created_at DATETIME,
  updated_at DATETIME,
  -- keyy is the item's key.
  -- Named keyy to avoid conflict within builtin identifier.
  keyy TEXT,
  -- value is the item's value.
  value TEXT,
  -- expires_at is when the item expires.
  -- If 0, the item never expires.
  expires_at DATETIME,
  PRIMARY KEY (key)
);

-- channel_command_cooldowns contains a record of a command cooldowns in channels.
CREATE TABLE channel_command_cooldowns (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME,
  -- channel is the channel the command has a cooldown in.
  channel TEXT,
  -- command is the name of the command with a cooldown.
  command TEXT,
  -- last_run is when the command was last run in the channel.
  last_run DATETIME
);
CREATE INDEX idx_channel_command_cooldowns_deleted_at ON channel_command_cooldowns (
  deleted_at
);

-- duels contains a record of gamba duels.
CREATE TABLE duels (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME,
  -- user_id is the ID of the user that initiated the duel.
  user_id INTEGER,
  -- target_id is the ID of the target of the duel.
  target_id INTEGER,
  -- amount is the amount duelled.
  amount INTEGER,
  -- pending is whether the duel is pending.
  pending NUMERIC,
  -- accepted is whether the target user has accepted the duel.
  accepted NUMERIC,
  -- won is whether the user won the duel.
  won NUMERIC,
  CONSTRAINT fk_duels_user FOREIGN KEY (user_id) REFERENCES users (
    id
  ),
  CONSTRAINT fk_duels_target FOREIGN KEY (target_id) REFERENCES users (
    id
  )
);
CREATE INDEX idx_duels_deleted_at ON duels (deleted_at);

-- gamba_transactions contains a record of gamba transactions.
CREATE TABLE gamba_transactions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME,
  -- user_id is the ID of the user that executed the transaction.
  user_id INTEGER,
  -- game is the gamba game the transaction was for.
  game TEXT,
  -- delta is the win/loss of the transaction.
  delta INTEGER,
  CONSTRAINT fk_gamba_transactions_user FOREIGN KEY (
    user_id
  ) REFERENCES users (id)
);
CREATE INDEX idx_gamba_transactions_deleted_at ON gamba_transactions (
  deleted_at
);

-- joined_channels contains a record of channels the bot should join.
CREATE TABLE joined_channels (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME,
  -- platform contains which platform this channel is on.
  platform TEXT,
  -- channel is which channel should be joined.
  channel TEXT,
  -- channel_id is the ID of the channel to be joined.
  channel_id TEXT,
  -- prefix is the command prefix to be used in the channel.
  prefix TEXT,
  -- joined_at is when the channel was joined.
  joined_at DATETIME
);
CREATE INDEX idx_joined_channels_deleted_at ON joined_channels (
  deleted_at
);

-- messages contains a record of chat messages.
CREATE TABLE messages (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME,
  -- text contains the text of the message.
  `text` TEXT,
  -- channel represents the channel the message was sent in
  -- (or should be sent in).
  channel TEXT,
  -- user_id is the ID of the user that sent the message.
  user_id INTEGER,
  -- time is when the message was sent.
  time DATETIME,
  CONSTRAINT fk_messages_user FOREIGN KEY (user_id) REFERENCES users (
    id
  )
);
CREATE INDEX idx_messages_deleted_at ON messages (deleted_at);

-- user_command_cooldowns contains a record of command cooldowns for users.
CREATE TABLE user_command_cooldowns (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME,

  -- user_id is the ID of the user with the cooldown.
  user_id INTEGER,
  -- command is the name of the command with a cooldown.
  command TEXT,
  -- last_run is when the command was last run in the channel.
  last_run DATETIME,
  CONSTRAINT fk_user_command_cooldowns_user FOREIGN KEY (
    user_id
  ) REFERENCES users (id)
);
CREATE INDEX idx_user_command_cooldowns_deleted_at ON user_command_cooldowns (
  deleted_at
);

CREATE TABLE users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME,
  -- twitch_id is the user's ID on Twitch, if known.
  twitch_id TEXT,
  -- twitch_name is the user's username on Twitch, if known.
  twitch_name TEXT
);
CREATE INDEX idx_users_deleted_at ON users (deleted_at);
