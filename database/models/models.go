// Package models defines database data models.
package models

import (
	"time"

	"gorm.io/gorm"
)

// AllModels contains one of each defined data model, for auto-migrations.
var AllModels = []any{
	BotBan{},
	ChannelCommandCooldown{},
	Duel{},
	GambaTransaction{},
	JoinedChannel{},
	Message{},
	User{},
	UserCommandCooldown{},
}

// BotBan represents a bot being banned from a channel.
type BotBan struct {
	ID        uint           `gorm:"primarykey"`
	CreatedAt time.Time      `gorm:"type:datetime"`
	UpdatedAt time.Time      `gorm:"type:datetime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Platform contains the which platform this channel is on.
	Platform string
	// Channel is which channel should be joined.
	Channel string
	// JoinedAt is when the channel was joined.
	BannedAt time.Time
}

// ChannelCommandCooldown contains a record of a command cooldown in a channel.
type ChannelCommandCooldown struct {
	ID        uint           `gorm:"primarykey"`
	CreatedAt time.Time      `gorm:"type:datetime"`
	UpdatedAt time.Time      `gorm:"type:datetime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Channel is the channel the command has a cooldown in.
	Channel string
	// Command is the name of the command with a cooldown.
	Command string
	// LastRun is when the command was last run in the channel.
	LastRun time.Time
}

// Duel represents a gamba duel.
type Duel struct {
	ID        uint           `gorm:"primarykey"`
	CreatedAt time.Time      `gorm:"type:datetime"`
	UpdatedAt time.Time      `gorm:"type:datetime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// UserID is the ID of the user that initiated the duel.
	UserID uint
	// User is the user that initiated the duel.
	User User
	// TargetID is the ID of the target of the duel.
	TargetID uint
	// Target is the target of the duel.
	Target User
	// Amount is the amount duelled.
	Amount int64
	// Pending is whether the duel is pending.
	Pending bool
	// Accepted is whether the target user has accepted the duel.
	Accepted bool
	// Won is whether the user won the duel.
	Won bool
}

// GambaTransaction represents a single gamba transaction.
type GambaTransaction struct {
	ID        uint           `gorm:"primarykey"`
	CreatedAt time.Time      `gorm:"type:datetime"`
	UpdatedAt time.Time      `gorm:"type:datetime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// UserID is the ID of the user that executed the transaction.
	UserID uint
	// User is the user that executed the transaction.
	User User
	// Game is the gamba game the transaction was for.
	Game string
	// Delta is the win/loss of the transaction.
	Delta int64
}

// JoinedChannel represents a channel the bot should join.
type JoinedChannel struct {
	ID        uint           `gorm:"primarykey"`
	CreatedAt time.Time      `gorm:"type:datetime"`
	UpdatedAt time.Time      `gorm:"type:datetime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Platform contains the which platform this channel is on.
	Platform string
	// Channel is which channel should be joined.
	Channel string
	// ChannelID is the ID of the channel to be joined.
	ChannelID string
	// Prefix is the prefix to be used in the channel.
	Prefix string
	// JoinedAt is when the channel was joined.
	JoinedAt time.Time
}

// Message represents a chat message.
type Message struct {
	ID        uint           `gorm:"primarykey"`
	CreatedAt time.Time      `gorm:"type:datetime"`
	UpdatedAt time.Time      `gorm:"type:datetime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Text contains the text of the message.
	Text string
	// Channel represents the channel the message was sent in
	// (or should be sent in).
	Channel string
	// UserID is the ID of the user that sent the message.
	UserID uint
	// User is the username of the user that sent the message.
	User User
	// Time is when the message was sent.
	Time time.Time
}

// User represents a user.
type User struct {
	ID        uint           `gorm:"primarykey"`
	CreatedAt time.Time      `gorm:"type:datetime"`
	UpdatedAt time.Time      `gorm:"type:datetime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// TwitchID is the user's ID on Twitch, if known
	TwitchID string
	// TwitchName is the user's username on Twitch, if known
	TwitchName string
}

// UserCommandCooldown contains a record of a command cooldown for a user.
type UserCommandCooldown struct {
	ID        uint           `gorm:"primarykey"`
	CreatedAt time.Time      `gorm:"type:datetime"`
	UpdatedAt time.Time      `gorm:"type:datetime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// UserID is the ID of the user with the cooldown.
	UserID uint
	// User is the user with the cooldown.
	User User
	// Command is the name of the command with a cooldown.
	Command string
	// LastRun is when the command was last run in the channel.
	LastRun time.Time
}
