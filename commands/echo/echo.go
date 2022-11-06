// Package echo implements commands that do simple echoes.
package echo

import (
	"fmt"
	"log"
	"regexp"
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
		Usage:      "$commands",
		Permission: permission.Normal,
		PrefixOnly: true,
		Pattern:    basecommand.PrefixPattern("commands"),
		Handler: func(msg *base.IncomingMessage) ([]*base.Message, error) {
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
		Usage:      "$gn",
		Permission: permission.Normal,
		PrefixOnly: true,
		Pattern:    basecommand.PrefixPattern("gn"),
		Handler: func(msg *base.IncomingMessage) ([]*base.Message, error) {
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
		Usage:      "$TriHard",
		Permission: permission.Normal,
		PrefixOnly: true,
		Pattern:    basecommand.PrefixPattern("TriHard"),
		Handler: func(msg *base.IncomingMessage) ([]*base.Message, error) {
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
	pyramidCommandPattern = basecommand.PrefixPattern("pyramid")
	pyramidCommand        = basecommand.Command{
		Name:            "pyramid",
		Help:            fmt.Sprintf("Makes a pyramid in chat. Max width %d.", maxPyramidWidth),
		Usage:           "$pyramid <width> <text>",
		Permission:      permission.Mod,
		ChannelCooldown: time.Duration(30) * time.Second,
		PrefixOnly:      true,
		Pattern:         pyramidCommandPattern,
		Handler:         pyramid,
	}
	pyramidPattern = regexp.MustCompile(pyramidCommandPattern.String() + `(\d+)\s+(.*)`)

	spamCommandPattern = basecommand.PrefixPattern("spam")
	spamCommand        = basecommand.Command{
		Name:            "spam",
		Help:            fmt.Sprintf("Sends a message many times. Max amount %d.", maxSpamAmount),
		Usage:           "$spam <count> <text>",
		Permission:      permission.Mod,
		ChannelCooldown: time.Duration(30) * time.Second,
		PrefixOnly:      true,
		Pattern:         spamCommandPattern,
		Handler:         spam,
	}
	spamPattern = regexp.MustCompile(spamCommandPattern.String() + `(\d+)\s+(.*)`)

	tuckCommandPattern = basecommand.PrefixPattern("tuck")
	tuckCommand        = basecommand.Command{
		Name:       "tuck",
		Help:       "Tuck someone to bed.",
		Usage:      "$tuck <user>",
		Permission: permission.Normal,
		PrefixOnly: true,
		Pattern:    tuckCommandPattern,
		Handler:    tuck,
	}
	tuckPattern = regexp.MustCompile(tuckCommandPattern.String() + `(\w+).*`)
)

func spam(msg *base.IncomingMessage) ([]*base.Message, error) {
	matches := spamPattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) != 3 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "Usage: $spam <count> <text>",
			},
		}, nil
	}

	count64, err := strconv.ParseInt(matches[1], 10, 32)
	if err != nil {
		log.Printf("Failed to parse %q as int", matches[1])
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "Usage: $spam <count> <text>",
			},
		}, nil
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

	text := matches[2]

	msgsCount := count
	msgs := make([]*base.Message, msgsCount)

	out := base.Message{Channel: msg.Message.Channel, Text: text}
	for i := 0; i < count; i++ {
		msgs[i] = &out
	}

	return msgs, nil
}

func pyramid(msg *base.IncomingMessage) ([]*base.Message, error) {
	matches := pyramidPattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) != 3 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "Usage: $pyramid <width> <text>",
			},
		}, nil
	}

	width64, err := strconv.ParseInt(matches[1], 10, 32)
	if err != nil {
		log.Printf("Failed to parse %q as int", matches[1])
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "Usage: $pyramid <width> <text>",
			},
		}, nil
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

	text := matches[2]

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

func tuck(msg *base.IncomingMessage) ([]*base.Message, error) {
	target := basecommand.ParseTarget(msg, tuckPattern)
	if strings.EqualFold(target, msg.Message.User) {
		return nil, nil
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("Bedge %s tucks %s into bed.", msg.Message.User, target),
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
