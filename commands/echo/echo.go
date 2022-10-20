// Package echo implements commands that do simple echoes.
package echo

import (
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/message"
	"github.com/airforce270/airbot/permission"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	{
		Name:       "commands",
		Help:       "Replies with a link to the commands.",
		Pattern:    basecommand.PrefixPattern("commands"),
		Handler:    commands,
		PrefixOnly: true,
		Permission: permission.Normal,
	},
	{
		Name:       "TriHard",
		Help:       "Replies with TriHard 7.",
		Pattern:    basecommand.PrefixPattern("TriHard"),
		Handler:    triHard,
		PrefixOnly: true,
		Permission: permission.Normal,
	},
}

func commands(msg *message.IncomingMessage) ([]*message.Message, error) {
	return []*message.Message{
		{
			Channel: msg.Message.Channel,
			Text:    "Commands available here: https://github.com/airforce270/airbot#commands",
		},
	}, nil
}

func triHard(msg *message.IncomingMessage) ([]*message.Message, error) {
	return []*message.Message{
		{
			Channel: msg.Message.Channel,
			Text:    "TriHard 7",
		},
	}, nil
}
