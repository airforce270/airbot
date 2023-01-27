// Package base provides base structs used throughout the application.
package base

import (
	"crypto/rand"
	"errors"
	"strings"
	"time"

	"github.com/airforce270/airbot/database/models"
	"github.com/airforce270/airbot/permission"

	exprand "golang.org/x/exp/rand"
)

var (
	ErrUserUnknown = errors.New("user has never been seen by the bot")
)

// Platform represents a connection to a given platform (i.e. Twitch, Discord)
type Platform interface {
	// Name returns the platform's name.
	Name() string
	// Username returns the bot's username within the platform.
	Username() string

	// Connect connects to the platform.
	Connect() error
	// Disconnect disconnects from the platform and should be called before exiting.
	Disconnect() error

	// Listen returns a channel that will provide incoming messages.
	Listen() <-chan IncomingMessage
	// Send sends a message.
	Send(m OutgoingMessage) error

	// Join joins a channel.
	Join(channel, prefix string) error
	// Leave leaves a channel.
	Leave(channel string) error
	// SetPrefix sets the prefix for a channel.
	SetPrefix(channel, prefix string) error

	// User returns the (database) user for the username of a user on the platform.
	// It will return ErrUserUnknown if the user has never been seen by the bot.
	User(username string) (models.User, error)
	// CurrentUsers returns the names of the current users in all channels the bot has joined.
	CurrentUsers() ([]string, error)

	// Timeout times out a user in a channel.
	Timeout(username, channel string, duration time.Duration) error
}

// Message represents a chat message.
type Message struct {
	// Text contains the text of the message.
	Text string
	// Channel represents the channel the message was sent in
	// (or should be sent in).
	Channel string
	// ID is the unique ID of the message, as provided by the platform.
	// This may not be set - some platforms do not provide an ID.
	ID string
	// UserID is the unique ID of the user that sent the message.
	UserID string
	// User is the username of the user that sent the message.
	User string
	// Time is when the message was sent.
	Time time.Time
}

// IncomingMessage represents an incoming chat message.
type IncomingMessage struct {
	// Message is the message.
	Message Message
	// Prefix is the prefix for the channel the message was sent in.
	Prefix string
	// PermissionLevel is the permission level of the user that sent the message.
	PermissionLevel permission.Level
	// Platform is the platform the message was sent on.
	Platform Platform
}

// MessageTextWithoutPrefix returns the message's text without the prefix.
func (m *IncomingMessage) MessageTextWithoutPrefix() string {
	return strings.Replace(m.Message.Text, m.Prefix, "", 1)
}

// OutgoingMessage represents an outgoing chat message.
type OutgoingMessage struct {
	// Message is the message.
	Message Message
	// ReplyToID is the ID of the message being replied to.
	// This field is only set if:
	//   1. The platform provides IDs for individual messages
	//   2. The platform supports replying to messages
	//   3. The message is a reply to another message
	ReplyToID string
}

var (
	RandReader                = rand.Reader
	RandSource exprand.Source = nil
)
