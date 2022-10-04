// Package echo implements commands that do simple echoes.
package echo

import (
	"regexp"

	"airbot/commands/basecommand"
	"airbot/message"
)

// Commands contains this package's commands.
var Commands = []basecommand.Command{
	{Prefix: "TriHard", F: triHard},
}

var triHardPattern = regexp.MustCompile(`TriHard\s+any\s+homies`)

func triHard(msg *message.Message) (*message.Message, error) {
	if triHardPattern.MatchString(msg.Text) {
		return &message.Message{
			Channel: msg.Channel,
			Text:    "TriHard 7",
		}, nil
	}
	return nil, nil
}
