// Package fun implements fun commands that hit APIs.
package fun

import (
	"crypto/rand"
	"fmt"
	"log"
	"math"
	"math/big"
	"strings"

	"github.com/airforce270/airbot/apiclients/bible"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/permission"

	"gonum.org/v1/gonum/stat/distuv"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	bibleVerseCommand,
	cockCommand,
	iqCommand,
}

var (
	bibleVerseCommand = basecommand.Command{
		Name:    "bibleverse",
		Aliases: []string{"bv"},
		Help:    "Looks up a bible verse.",
		Args: []basecommand.Argument{
			{Name: "book", Required: true},
			{Name: "chapter:verse", Required: true},
		},
		Permission: permission.Normal,
		Handler:    bibleVerse,
	}

	cockCommand = basecommand.Command{
		Name:       "cock",
		Help:       "Tells you the length :)",
		Args:       []basecommand.Argument{{Name: "user", Required: false}},
		Permission: permission.Normal,
		Handler:    cock,
	}

	iqCommand = basecommand.Command{
		Name:       "iq",
		Help:       "Tells you someone's IQ",
		Args:       []basecommand.Argument{{Name: "user", Required: false}},
		Permission: permission.Normal,
		Handler:    iq,
	}
)

func bibleVerse(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
	if len(args) < 2 {
		return nil, basecommand.ErrReturnUsage
	}

	verses, err := bible.FetchVerses(strings.Join(args, " "))
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

func cock(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
	target := basecommand.FirstArgOrUsername(args, msg)

	length, err := rand.Int(base.RandReader, big.NewInt(cockMaxLength+1))
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

func iq(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
	target := basecommand.FirstArgOrUsername(args, msg)

	userIqFloat := distuv.Normal{Mu: 100, Sigma: 15, Src: base.RandSource}.Rand()
	userIq := int64(math.Round(userIqFloat))

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("%s's IQ is %d", target, userIq),
		},
	}, nil
}
