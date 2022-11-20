// Package echo implements commands that do simple echoes.
package echo

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/permission"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	{
		Name:       "commands",
		Help:       "Replies with a link to the commands.",
		Permission: permission.Normal,
		Handler: func(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    "Commands available here: https://github.com/airforce270/airbot/blob/main/docs/commands.md",
				},
			}, nil
		},
	},
	{
		Name:       "gn",
		Help:       "Says good night.",
		Permission: permission.Normal,
		Handler: func(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    fmt.Sprintf("FeelsOkayMan <3 gn %s", msg.Message.User),
				},
			}, nil
		},
	},
	pyramidCommand,
	spamCommand,
	{
		Name:       "TriHard",
		Help:       "Replies with TriHard 7.",
		Permission: permission.Normal,
		Handler: func(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    "TriHard 7",
				},
			}, nil
		},
	},
	tuckCommand,
}

const (
	safeMessageCount = 50
	maxPyramidWidth  = safeMessageCount / 2
	maxSpamAmount    = safeMessageCount
)

var (
	pyramidCommand = basecommand.Command{
		Name: "pyramid",
		Help: fmt.Sprintf("Makes a pyramid in chat. Max width %d.", maxPyramidWidth),
		Args: []basecommand.Argument{
			{Name: "width", Required: true},
			{Name: "text", Required: true},
		},
		Permission:      permission.Mod,
		ChannelCooldown: time.Duration(30) * time.Second,
		Handler:         pyramid,
	}

	spamCommand = basecommand.Command{
		Name: "spam",
		Help: fmt.Sprintf("Sends a message many times. Max amount %d.", maxSpamAmount),
		Args: []basecommand.Argument{
			{Name: "count", Required: true},
			{Name: "text", Required: true},
		},
		Permission:      permission.Mod,
		ChannelCooldown: time.Duration(30) * time.Second,
		Handler:         spam,
	}

	tuckCommand = basecommand.Command{
		Name:       "tuck",
		Help:       "Tuck someone to bed.",
		Args:       []basecommand.Argument{{Name: "user", Required: true}},
		Permission: permission.Normal,
		Handler:    tuck,
	}
)

func spam(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
	if len(args) < 2 {
		return nil, basecommand.ErrReturnUsage
	}

	count64, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		log.Printf("Failed to parse %q as int", args[0])
		return nil, basecommand.ErrReturnUsage
	}
	count := int(count64)

	if count > maxSpamAmount {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Max spam amount is %d", maxSpamAmount),
			},
		}, nil
	}

	text := strings.Join(args[1:], " ")

	msgsCount := count
	msgs := make([]*base.Message, msgsCount)

	out := base.Message{Channel: msg.Message.Channel, Text: text}
	for i := 0; i < count; i++ {
		msgs[i] = &out
	}

	return msgs, nil
}

func pyramid(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
	if len(args) < 2 {
		return nil, basecommand.ErrReturnUsage
	}

	width64, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		log.Printf("Failed to parse %q as int", args[0])
		return nil, basecommand.ErrReturnUsage
	}
	width := int(width64)

	if width > maxPyramidWidth {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Max pyramid width is %d", maxPyramidWidth),
			},
		}, nil
	}

	text := strings.Join(args[1:], " ")

	msgsCount := width + width - 1
	msgs := make([]*base.Message, msgsCount)

	for i := 0; i < width; i++ {
		msgs[i] = &base.Message{Channel: msg.Message.Channel, Text: repeatJoin(text, i+1, " ")}
	}

	offset := 1
	for i := width; i < msgsCount; i++ {
		msgs[i] = &base.Message{Channel: msg.Message.Channel, Text: repeatJoin(text, i-offset, " ")}
		offset += 2
	}

	return msgs, nil
}

func tuck(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
	if len(args) == 0 {
		return nil, basecommand.ErrReturnUsage
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("Bedge %s tucks %s into bed.", msg.Message.User, args[0]),
		},
	}, nil
}

// repeatJoin repeats a string and joins it on a delimiter.
func repeatJoin(s string, n int, delimiter string) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString(s)
		if i == n-1 {
			continue
		}
		b.WriteString(delimiter)
	}
	return b.String()
}
