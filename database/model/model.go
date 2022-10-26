// Package model defines database data models.
package model

import (
	"time"

	"gorm.io/gorm"
)

// AllModels contains one of each defined data model, for auto-migrations.
var AllModels = []any{
	BotBan{},
	JoinedChannel{},
	Message{},
	User{},
}

// BotBan represents a bot being banned from a channel.
type BotBan struct {
	gorm.Model

	// Platform contains the which platform this channel is on.
	Platform string
	// Channel is which channel should be joined.
	Channel string
	// JoinedAt is when the channel was joined.
	BannedAt time.Time
}

// JoinedChannel represents a channel the bot should join.
type JoinedChannel struct {
	gorm.Model

	// Platform contains the which platform this channel is on.
	Platform string
	// Channel is which channel should be joined.
	Channel string
	// Prefix is the prefix to be used in the channel.
	Prefix string
	// JoinedAt is when the channel was joined.
	JoinedAt time.Time
}

// Message represents a chat message.
type Message struct {
	gorm.Model

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
	gorm.Model

	// TwitchID is the user's ID on Twitch, if known
	TwitchID string
	// TwitchName is the user's username on Twitch, if known
	TwitchName string
}
