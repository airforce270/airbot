// Package base provides base structs used throughout the application.
package base

import (
	"errors"
	"io"
	"strings"
	"time"

	"github.com/airforce270/airbot/apiclients/bible"
	"github.com/airforce270/airbot/apiclients/ivr"
	"github.com/airforce270/airbot/apiclients/kick"
	"github.com/airforce270/airbot/apiclients/seventv"
	"github.com/airforce270/airbot/cache"
	"github.com/airforce270/airbot/database/models"
	"github.com/airforce270/airbot/permission"
	"gorm.io/gorm"

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
	Send(m Message) error
	// Reply sends a message in reply to another message.
	Reply(m Message, replyToID string) error

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
	// Resources contains resources available to an incoming message.
	Resources Resources
}

// MessageTextWithoutPrefix returns the message's text without the prefix.
func (m *IncomingMessage) MessageTextWithoutPrefix() string {
	return strings.Replace(m.Message.Text, m.Prefix, "", 1)
}

// OutgoingMessage represents an outgoing chat message.
type OutgoingMessage struct {
	// Message is the message.
	Message
	// ReplyToID is the ID of the message being replied to.
	// This field is only set if:
	//   1. The platform provides IDs for individual messages
	//   2. The platform supports replying to messages
	//   3. The message is a reply to another message
	ReplyToID string
}

// Resources contains references to app-level resources.
type Resources struct {
	// Platform is the current platform.
	Platform Platform
	// DB is a reference to the database.
	DB *gorm.DB
	// Cache is a reference to the cache.
	Cache cache.Cache
	// AllPlatforms contains all platforms currently registered with the bot.
	AllPlatforms map[string]Platform
	// NewConfigSource is a function that returns a source of data
	// for the latest config.
	NewConfigSource func() (io.ReadCloser, error)
	// Rand is a reference to random sources.
	Rand RandResources
	// Clients contains API clients.
	Clients APIClients
}

// PlatformByName returns the platform with a given name
// if it's currently registered and configured.
func (r Resources) PlatformByName(name string) (plat Platform, ok bool) {
	plat, ok = r.AllPlatforms[name]
	return plat, ok
}

// RandResources contains references to random number resources.
type RandResources struct {
	// Reader will be used as the reader for random values.
	Reader io.Reader
	// Source is a source of random numbers.
	// Optional - a default will be used if not provided.
	Source exprand.Source
}

// APIClients contains external API clients.
type APIClients struct {
	// Bible API client.
	Bible *bible.Client
	// IVR API client.
	IVR *ivr.Client
	// Kick API client.
	Kick *kick.Client
	// Pastebin FetchPaste URL override.
	// If set, this will override whatever the user enters.
	// Therefore, it should only be set in test.
	PastebinFetchPasteURLOverride string
	// 7TV API client.
	SevenTV *seventv.Client
}
