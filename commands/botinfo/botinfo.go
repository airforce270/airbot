// Package botinfo implements commands that return info about the bot.
package botinfo

import (
	"fmt"

	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/message"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	{
		Pattern:    basecommand.PrefixPattern("prefix"),
		Handle:     prefix,
		PrefixOnly: true,
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
