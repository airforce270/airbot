// Package basecommand contains shared types and utilities for command handlers.
package basecommand

import "airbot/message"

// Command represents a command the bot handles.
type Command struct {
	// Prefix is the prefix the message should have to be handled by this command.
	Prefix string
	// F is the function to be run if this command matches.
	F func(msg *message.Message) (*message.Message, error)
}
