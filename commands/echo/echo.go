// Package echo implements commands that do simple echoes.
package echo

import (
	"airbot/commands/basecommand"
	"airbot/message"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	{
		Pattern: basecommand.PrefixPattern("TriHard"),
		Handle:  triHard,
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
