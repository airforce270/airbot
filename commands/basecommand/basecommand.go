// Package basecommand contains shared types and utilities for command handlers.
package basecommand

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/base/arg"
	"github.com/airforce270/airbot/permission"
)

// ErrBadUsage indicates the command was used incorrectly.
// i.e.: not enough args, etc.
var ErrBadUsage = errors.New("bad usage")

// Command represents a command the bot handles.
type Command struct {
	// Name is the name of the command.
	Name string
	// Aliases are the aliases/alternate names for this command, if any.
	Aliases []string
	// Desc is the description of this command.
	Desc string
	// Params contains the parameters to the command.
	// Currently, only the last param can be optional.
	// All params before the last should be required.
	Params []arg.Param
	// Permission is the permission level required to run the command.
	Permission permission.Level
	// ChannelCooldown is the cooldown between when a command can be used in a given channel.
	ChannelCooldown time.Duration
	// ChannelCooldown is the cooldown between when a command can be used by a given user.
	UserCooldown time.Duration
	// DisableReplies will disable the returned messages from being sent as replies.
	// This only has an effect on platforms that support replies.
	DisableReplies bool
	// Handler is the function to be run if this command matches.
	// args contains the arguments to the command as specified by Params.
	Handler func(ctx context.Context, msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error)
}

// Pattern compiles the regexp to match this command.
// The returned command assumes the text has no prefix.
func (c *Command) Compile() (*regexp.Regexp, error) {
	nameParts := []string{c.Name}
	nameParts = append(nameParts, c.Aliases...)
	var paramGroups strings.Builder
	for _, a := range c.Params {
		if a.Type == arg.Variadic {
			paramGroups.WriteString(`(?:\s+(.*))?`)
		} else {
			paramGroups.WriteString(`(?:\s+(\S*))?`)
		}
	}
	if paramGroups.Len() == 0 {
		paramGroups.WriteString(`(?:\s+(\S*))?`)
	}
	return regexp.Compile(fmt.Sprintf(`^\s*(?:%s)(?:\s+|(?:%s))?\s*$`, strings.Join(nameParts, "|"), paramGroups.String()))
}

// Usage returns usage information for the command.
func (c *Command) Usage(prefix string) string {
	var resp strings.Builder
	resp.WriteString(prefix + c.Name)
	for _, arg := range c.Params {
		if arg.Required {
			fmt.Fprintf(&resp, " <%s>", arg.UsageForDocString())
		} else {
			fmt.Fprintf(&resp, " [%s]", arg.UsageForDocString())
		}
	}
	return resp.String()
}

// Help returns help information for the command.
func (c *Command) Help() string {
	if c.Desc == "" {
		return "<no help information found>"
	}

	if c.ChannelCooldown == 0 && c.UserCooldown == 0 {
		return c.Desc
	}

	var parts strings.Builder
	parts.WriteString(c.Desc)

	if !strings.HasSuffix(c.Desc, " ") {
		parts.WriteString(" ")
	}

	var cooldowns []string
	if c.ChannelCooldown > 0 && c.UserCooldown > 0 {
		cooldowns = append(cooldowns, fmt.Sprintf("Channel-wide cooldown: %s, user-specific cooldown: %s", c.ChannelCooldown, c.UserCooldown))
	} else {
		if c.ChannelCooldown > 0 {
			cooldowns = append(cooldowns, fmt.Sprintf("Channel-wide cooldown: %s", c.ChannelCooldown))
		}
		if c.UserCooldown > 0 {
			cooldowns = append(cooldowns, fmt.Sprintf("User-specific cooldown: %s", c.UserCooldown))
		}
	}
	parts.WriteString(strings.Join(cooldowns, ", "))

	return parts.String()
}

// FirstArgOrUsername returns the first provided arg, or the message's sender if not present.
func FirstArgOrUsername(args []arg.Arg, msg *base.IncomingMessage) string {
	if len(args) == 0 {
		return msg.Message.User
	}
	if firstArg := args[0]; firstArg.Present {
		return firstArg.StringValue
	}
	return msg.Message.User
}

// FirstArgOrChannel returns the first provided arg, or the message's channel if not present.
func FirstArgOrChannel(args []arg.Arg, msg *base.IncomingMessage) string {
	if len(args) == 0 {
		return msg.Message.Channel
	}
	if firstArg := args[0]; firstArg.Present {
		return firstArg.StringValue
	}
	return msg.Message.Channel
}
