// Package echo implements commands that do simple echoes.
package echo

import (
	"fmt"
	"strings"
	"time"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/base/arg"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/permission"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	{
		Name:       "commands",
		Help:       "Replies with a link to the commands.",
		Permission: permission.Normal,
		Handler: func(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
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
		Handler: func(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
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
		Handler: func(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
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
		Params: []arg.Param{
			{Name: "width", Type: arg.Int, Required: true},
			{Name: "text", Type: arg.String, Required: true},
		},
		Permission:      permission.Mod,
		ChannelCooldown: time.Duration(30) * time.Second,
		Handler:         pyramid,
	}

	spamCommand = basecommand.Command{
		Name: "spam",
		Help: fmt.Sprintf("Sends a message many times. Max amount %d.", maxSpamAmount),
		Params: []arg.Param{
			{Name: "count", Type: arg.Int, Required: true},
			{Name: "text", Type: arg.Variadic, Required: true},
		},
		Permission:      permission.Mod,
		ChannelCooldown: time.Duration(30) * time.Second,
		Handler:         spam,
	}

	tuckCommand = basecommand.Command{
		Name:       "tuck",
		Help:       "Tuck someone to bed.",
		Params:     []arg.Param{{Name: "user", Type: arg.Username, Required: true}},
		Permission: permission.Normal,
		Handler:    tuck,
	}
)

func spam(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	countArg, textArg := args[0], args[1]
	if !countArg.Present || !textArg.Present {
		return nil, basecommand.ErrBadUsage
	}

	count := countArg.IntValue

	if count > maxSpamAmount {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Max spam amount is %d", maxSpamAmount),
			},
		}, nil
	}

	text := textArg.StringValue

	msgsCount := count
	msgs := make([]*base.Message, msgsCount)

	out := base.Message{Channel: msg.Message.Channel, Text: text}
	for i := 0; i < count; i++ {
		msgs[i] = &out
	}

	return msgs, nil
}

func pyramid(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	widthArg, textArg := args[0], args[1]
	if !widthArg.Present || !textArg.Present {
		return nil, basecommand.ErrBadUsage
	}

	width := widthArg.IntValue

	if width > maxPyramidWidth {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Max pyramid width is %d", maxPyramidWidth),
			},
		}, nil
	}

	text := textArg.StringValue

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

func tuck(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	userArg := args[0]
	if !userArg.Present {
		return nil, basecommand.ErrBadUsage
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("Bedge %s tucks %s into bed.", msg.Message.User, userArg.StringValue),
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
