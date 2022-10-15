// Package basecommand contains shared types and utilities for command handlers.
package basecommand

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/airforce270/airbot/message"
)

// Command represents a command the bot handles.
type Command struct {
	// Pattern is the regexp pattern that should match for this command.
	Pattern *regexp.Regexp
	// Handle is the function to be run if this command matches.
	Handle func(msg *message.IncomingMessage) ([]*message.Message, error)
	// PrefixOnly is whether the command should only be triggered if used with the prefix.
	// i.e. `$title xqc` but not `title xqc`
	PrefixOnly bool
	// AdminOnly is whether the command can only be run by admins.
	AdminOnly bool
}

// PrefixPattern compiles a regex pattern matching the prefix of a string, ignoring whitespace.
func PrefixPattern(pattern string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(`^\s*%s\s*`, pattern))
}

// ParseTarget parses the first regex pattern match from the message's text.
// If no match is found, uses the user's username instead.
func ParseTarget(msg *message.IncomingMessage, pattern *regexp.Regexp) string {
	matches := pattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) <= 1 {
		return strings.ToLower(msg.Message.User)
	}
	return strings.ToLower(matches[1])
}
