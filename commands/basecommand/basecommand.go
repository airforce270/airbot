// Package basecommand contains shared types and utilities for command handlers.
package basecommand

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/permission"
)

// ErrReturnUsage should be returned from a command handler when usage information should be returned.
var ErrReturnUsage = errors.New("should return usage information")

// Command represents a command the bot handles.
type Command struct {
	// Name is the name of the command.
	Name string
	// Aliases are the aliases/alternate names for this command, if any.
	Aliases []string
	// Help is the help information for this command.
	Help string
	// Args contains the arguments to the command.
	// Currently, only the last argument can be optional.
	// All args before the last should be required.
	Args []Argument
	// Permission is the permission level required to run the command.
	Permission permission.Level
	// ChannelCooldown is the cooldown between when a command can be used in a given channel.
	ChannelCooldown time.Duration
	// ChannelCooldown is the cooldown between when a command can be used by a given user.
	UserCooldown time.Duration
	// Handler is the function to be run if this command matches.
	// args contains the arguments to the command as specified by Args.
	Handler func(msg *base.IncomingMessage, args []string) ([]*base.Message, error)
}

// Pattern compiles the regexp to match this command.
// The returned command assumes the text has no prefix.
func (c *Command) Compile() (*regexp.Regexp, error) {
	nameParts := []string{c.Name}
	nameParts = append(nameParts, c.Aliases...)
	var argGroups strings.Builder
	if len(c.Args) > 0 {
		argGroups.WriteString(`\s+(\S*)`)
	}
	if len(c.Args) > 1 {
		argGroups.WriteString(strings.Repeat(`\s*(\S*)`, len(c.Args)-1))
	}
	return regexp.Compile(fmt.Sprintf(`^\s*(?:%s)(?:\s+|(?:%s))?\s*$`, strings.Join(nameParts, "|"), argGroups.String()))
}

// Usage returns usage information for the command.
func (c *Command) Usage(prefix string) string {
	parts := []string{prefix + c.Name}
	for _, arg := range c.Args {
		if arg.ExcludeFromUsage {
			continue
		}
		if arg.Required {
			parts = append(parts, fmt.Sprintf("<%s>", arg.UsageForDocString()))
		} else {
			parts = append(parts, fmt.Sprintf("[%s]", arg.UsageForDocString()))
		}
	}
	return strings.Join(parts, " ")
}

// Argument represents an argument to a command.
type Argument struct {
	// Name is the name of the argument.
	Name string
	// Required is whether the argument is required.
	Required bool
	// Usage is an optional human-readable string describing the arg.
	// This is only used for the usage string, i.e. if Name:"myarg" Usage:"something",
	// the usage string will say $command <something> rather than $command <myarg>
	Usage string
	// Whether this arg should be excluded from usage information.
	ExcludeFromUsage bool
}

// UsageForDocString returns the usage information for putting in a usage docstring,
// i.e. for a command named "mycommand" with an arg named "myarg" with arg usage of "myargusage"
// $command <myargusage>
func (a Argument) UsageForDocString() string {
	if a.Usage != "" {
		return a.Usage
	}
	return a.Name
}

// FirstArgOrUsername returns the first provided arg, or the message's sender if not present.
func FirstArgOrUsername(args []string, msg *base.IncomingMessage) string {
	if len(args) < 1 {
		return msg.Message.User
	}
	return args[0]
}

// FirstArgOrChannel returns the first provided arg, or the message's channel if not present.
func FirstArgOrChannel(args []string, msg *base.IncomingMessage) string {
	if len(args) < 1 {
		return msg.Message.Channel
	}
	return args[0]
}
