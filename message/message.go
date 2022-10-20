// Package message provides a platform-agnostic message type.
package message

import (
	"strings"
	"time"

	"github.com/airforce270/airbot/permission"
)

// Message represents a chat message.
type Message struct {
	// Text contains the text of the message.
	Text string
	// Channel represents the channel the message was sent in
	// (or should be sent in).
	Channel string
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
}

// MessageTextWithoutPrefix returns the message's text without the prefix.
func (m *IncomingMessage) MessageTextWithoutPrefix() string {
	return strings.Replace(m.Message.Text, m.Prefix, "", 1)
}
