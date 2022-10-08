// Package basecommand contains shared types and utilities for command handlers.
package basecommand

import (
	"fmt"
	"regexp"

	"airbot/message"
)

// Command represents a command the bot handles.
type Command struct {
	// Pattern is the regexp pattern that should match for this command.
	Pattern *regexp.Regexp
	// Handle is the function to be run if this command matches.
	Handle func(msg *message.IncomingMessage) ([]*message.Message, error)
}

func PrefixPattern(pattern string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(`^\s*%s\s*`, pattern))
}
