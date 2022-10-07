// Package echo implements commands that do simple echoes.
package echo

import (
	"airbot/commands/basecommand"
	"airbot/message"
)

// Commands contains this package's commands.
var Commands = []basecommand.Command{
	{Prefix: "trihard", F: triHard},
}

func triHard(msg *message.Message) (*message.Message, error) {
	return &message.Message{
		Channel: msg.Channel,
		Text:    "TriHard 7",
	}, nil
}
