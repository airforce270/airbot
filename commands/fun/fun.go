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
	"github.com/airforce270/airbot/base/arg"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/permission"

	"gonum.org/v1/gonum/stat/distuv"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	bibleVerseCommand,
	cockCommand,
	iqCommand,
	shipCommand,
}

var (
	bibleVerseCommand = basecommand.Command{
		Name:    "bibleverse",
		Aliases: []string{"bv"},
		Help:    "Looks up a bible verse.",
		Params: []arg.Param{
			{Name: "book", Type: arg.String, Required: true},
			{Name: "chapter:verse", Type: arg.String, Required: true},
		},
		Permission: permission.Normal,
		Handler:    bibleVerse,
	}

	cockCommand = basecommand.Command{
		Name:       "cock",
		Help:       "Tells you the length :)",
		Params:     []arg.Param{{Name: "user", Type: arg.Username, Required: false}},
		Permission: permission.Normal,
		Handler:    cock,
	}

	iqCommand = basecommand.Command{
		Name:       "iq",
		Help:       "Tells you someone's IQ",
		Params:     []arg.Param{{Name: "user", Type: arg.Username, Required: false}},
		Permission: permission.Normal,
		Handler:    iq,
	}

	shipCommand = basecommand.Command{
		Name: "ship",
		Help: "Tells you the compatibility of two people.",
		Params: []arg.Param{
			{Name: "first-person", Type: arg.Username, Required: true},
			{Name: "second-person", Type: arg.Username, Required: true},
		},
		Permission: permission.Normal,
		Handler:    ship,
	}
)

func bibleVerse(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	bookArg, chapterVerseArg := args[0], args[1]
	if !bookArg.IsPresent || !chapterVerseArg.IsPresent {
		return nil, basecommand.ErrBadUsage
	}
	book, chapterVerse := bookArg.Value.(string), chapterVerseArg.Value.(string)

	verses, err := bible.FetchVerses(book + " " + chapterVerse)
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

func cock(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
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

func iq(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
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

func ship(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	person1Arg, person2Arg := args[0], args[1]
	if !person1Arg.IsPresent || !person2Arg.IsPresent {
		return nil, basecommand.ErrBadUsage
	}
	person1, person2 := person1Arg.Value.(string), person2Arg.Value.(string)

	percentBigInt, err := rand.Int(base.RandReader, big.NewInt(101))
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
