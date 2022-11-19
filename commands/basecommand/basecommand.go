// Package basecommand contains shared types and utilities for command handlers.
package basecommand

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/permission"
)

// Command represents a command the bot handles.
type Command struct {
	// Name is the name of the command.
	Name string
	// Aliases are the aliases/alternate names for this command, if any.
	Aliases []string
	// Help is the help information for this command.
	Help string
	// Usage is the usage information for this command.
	// Should be in the format `$command <required-param> [optional-param]`
	Usage string
	// Permission is the permission level required to run the command.
	Permission permission.Level
	// ChannelCooldown is the cooldown between when a command can be used in a given channel.
	ChannelCooldown time.Duration
	// ChannelCooldown is the cooldown between when a command can be used by a given user.
	UserCooldown time.Duration
	// PrefixOnly is whether the command should only be triggered if used with the prefix.
	// i.e. `$title xqc` but not `title xqc`
	PrefixOnly bool
	// Pattern is the regexp pattern that should match for this command.
	Pattern *regexp.Regexp
	// Handler is the function to be run if this command matches.
	Handler func(msg *base.IncomingMessage) ([]*base.Message, error)
}

// PrefixPattern compiles a regex pattern matching the prefix of a string, ignoring whitespace.
func PrefixPattern(pattern string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(`^\s*%s\s*`, pattern))
}

// ParseTarget parses the first regex pattern match from the message's text.
// If no match is found, uses the user's username instead.
func ParseTarget(msg *base.IncomingMessage, pattern *regexp.Regexp) string {
	return ParseTargetWithDefault(msg, pattern, msg.Message.User)
}

// ParseTarget parses the first regex pattern match from the message's text.
// If no match is found, uses a default instead.
func ParseTargetWithDefault(msg *base.IncomingMessage, pattern *regexp.Regexp, defaultTarget string) string {
	matches := pattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) <= 1 {
		return defaultTarget
	}
	return strings.ToLower(matches[1])
}
