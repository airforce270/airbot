// Package botinfo implements commands that return info about the bot.
package botinfo

import (
	"fmt"
	"regexp"

	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/message"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	{
		Name:       "prefix",
		Help:       "Replies with the prefix in this channel.",
		Pattern:    regexp.MustCompile(`\s*(^|(wh?at( i|')?s (the |air|af2)(bot('?s)?)? ?))prefix\??\s*`),
		Handler:    prefix,
		PrefixOnly: false,
	},
}

func prefix(msg *message.IncomingMessage) ([]*message.Message, error) {
	return []*message.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("This channel's prefix is %s", msg.Prefix),
		},
	}, nil
}
