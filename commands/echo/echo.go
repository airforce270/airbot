// Package echo implements commands that do simple echoes.
package echo

import (
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/message"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	{
		Pattern:    basecommand.PrefixPattern("TriHard"),
		Handle:     triHard,
		PrefixOnly: true,
	},
}

func triHard(msg *message.IncomingMessage) ([]*message.Message, error) {
	return []*message.Message{
		{
			Channel: msg.Message.Channel,
			Text:    "TriHard 7",
		},
	}, nil
}
