// Package fun implements fun commands that hit APIs.
package fun

import (
	"context"
	"crypto/rand"
	_ "embed"
	"fmt"
	"log"
	"math"
	"math/big"
	"strings"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/base/arg"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/permission"

	"gonum.org/v1/gonum/stat/distuv"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	bibleVerseCommand,
	cockCommand,
	fortuneCommand,
	iqCommand,
	shipCommand,
}

var (
	bibleVerseCommand = basecommand.Command{
		Name:    "bibleverse",
		Aliases: []string{"bv"},
		Desc:    "Looks up a bible verse.",
		Params: []arg.Param{
			{Name: "book", Type: arg.String, Required: true},
			{Name: "chapter:verse", Type: arg.String, Required: true},
		},
		Permission: permission.Normal,
		Handler:    bibleVerse,
	}

	cockCommand = basecommand.Command{
		Name:       "cock",
		Aliases:    []string{"kok"},
		Desc:       "Tells you the length :)",
		Params:     []arg.Param{{Name: "user", Type: arg.Username, Required: false}},
		Permission: permission.Normal,
		Handler:    cock,
	}

	//go:embed data/fortunes.txt
	fortunesText string

	fortunes    = truncate(readFileWithoutNewlines(fortunesText), 480 /* maxLength */)
	fortunesLen = big.NewInt(int64(len(fortunes)))

	fortuneCommand = basecommand.Command{
		Name:       "fortune",
		Desc:       "Replies with a fortune. Fortunes from https://github.com/bmc/fortunes",
		Permission: permission.Normal,
		Handler: func(ctx context.Context, msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
			i, err := rand.Int(msg.Resources.Rand.Reader, fortunesLen)
			if err != nil {
				return nil, fmt.Errorf("failed to generate random number: %w", err)
			}
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    fortunes[int(i.Uint64())],
				},
			}, nil
		},
	}

	iqCommand = basecommand.Command{
		Name:       "iq",
		Desc:       "Tells you someone's IQ",
		Params:     []arg.Param{{Name: "user", Type: arg.Username, Required: false}},
		Permission: permission.Normal,
		Handler:    iq,
	}

	shipCommand = basecommand.Command{
		Name: "ship",
		Desc: "Tells you the compatibility of two people.",
		Params: []arg.Param{
			{Name: "first-person", Type: arg.Username, Required: true},
			{Name: "second-person", Type: arg.Username, Required: true},
		},
		Permission: permission.Normal,
		Handler:    ship,
	}
)

func bibleVerse(ctx context.Context, msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	bookArg, chapterVerseArg := args[0], args[1]
	if !bookArg.Present || !chapterVerseArg.Present {
		return nil, basecommand.ErrBadUsage
	}
	book, chapterVerse := bookArg.StringValue, chapterVerseArg.StringValue

	verses, err := msg.Resources.Clients.Bible.FetchVerses(book + " " + chapterVerse)
	if err != nil {
		log.Printf("Failed to look up Bible verses: %v", err)
		return nil, nil
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("[%s]: %s", verses.Reference, verses.Text),
		},
	}, nil
}

const cockMaxLength = 14

func cock(ctx context.Context, msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	target := basecommand.FirstArgOrUsername(args, msg)

	length, err := rand.Int(msg.Resources.Rand.Reader, big.NewInt(cockMaxLength+1))
	if err != nil {
		return nil, fmt.Errorf("failed to generate random number: %w", err)
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("%s's cock is %d inches long", target, length),
		},
	}, nil
}

func iq(ctx context.Context, msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	target := basecommand.FirstArgOrUsername(args, msg)

	userIqFloat := distuv.Normal{Mu: 100, Sigma: 15, Src: msg.Resources.Rand.Source}.Rand()
	userIq := int64(math.Round(userIqFloat))

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("%s's IQ is %d", target, userIq),
		},
	}, nil
}

func ship(ctx context.Context, msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	person1Arg, person2Arg := args[0], args[1]
	if !person1Arg.Present || !person2Arg.Present {
		return nil, basecommand.ErrBadUsage
	}
	person1, person2 := person1Arg.StringValue, person2Arg.StringValue

	percentBigInt, err := rand.Int(msg.Resources.Rand.Reader, big.NewInt(101))
	if err != nil {
		return nil, fmt.Errorf("failed to generate random number: %w", err)
	}
	percent := percentBigInt.Int64()

	var suffix string
	if percent >= 90 {
		suffix = "invite me to the wedding please 😍"
	} else if percent >= 80 {
		suffix = "oh 😳"
	} else if percent >= 60 {
		suffix = "worth a shot ;)"
	} else if percent >= 40 {
		suffix = "it's a toss-up :/"
	} else if percent >= 20 {
		suffix = "not sure about this one... :("
	} else {
		suffix = "don't even think about it DansGame"
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text: strings.Join([]string{
				fmt.Sprintf("%s and %s have a %d", person1, person2, percent),
				"% ",
				fmt.Sprintf("compatibility, %s", suffix),
			}, ""),
		},
	}, nil
}

func truncate(strings []string, maxLength int) []string {
	truncated := make([]string, 0, maxLength)

	for _, s := range strings {
		if len(s) > maxLength {
			truncated = append(truncated, s[:maxLength])
		} else {
			truncated = append(truncated, s)
		}
	}

	return truncated
}

func readFileWithoutNewlines(text string) []string {
	linesWithNewlines := strings.Split(text, "\n")
	lines := make([]string, 0, len(linesWithNewlines))
	for _, line := range linesWithNewlines {
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}
