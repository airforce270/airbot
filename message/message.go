// Package message provides a platform-agnostic message type.
package message

import "time"

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

type IncomingMessage struct {
	Message Message
	Prefix  string
}
