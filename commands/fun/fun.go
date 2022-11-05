// Package fun implements fun commands that hit APIs.
package fun

import (
	"crypto/rand"
	"fmt"
	"log"
	"math"
	"math/big"
	"regexp"
	"strings"

	"github.com/airforce270/airbot/apiclients/bible"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/permission"

	exprand "golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	bibleVerseCommand,
	cockCommand,
	iqCommand,
}

var (
	bibleVerseCommandPattern = basecommand.PrefixPattern(`(?:bibleverse|bv)`)
	bibleVerseCommand        = basecommand.Command{
		Name:           "bibleverse",
		AlternateNames: []string{"bv"},
		Help:           "Looks up a bible verse.",
		Usage:          "$bibleverse <book> <chapter:verse>",
		Permission:     permission.Normal,
		PrefixOnly:     true,
		Pattern:        bibleVerseCommandPattern,
		Handler:        bibleVerse,
	}
	bibleVersePattern = regexp.MustCompile(bibleVerseCommandPattern.String() + `(.*)`)

	cockCommandPattern = basecommand.PrefixPattern("cock")
	cockCommand        = basecommand.Command{
		Name:       "cock",
		Help:       "Tells you the length :)",
		Usage:      "$cock [user]",
		Permission: permission.Normal,
		PrefixOnly: true,
		Pattern:    cockCommandPattern,
		Handler:    cock,
	}
	cockPattern = regexp.MustCompile(cockCommandPattern.String() + `(\w+)`)

	iqCommandPattern = basecommand.PrefixPattern("iq")
	iqCommand        = basecommand.Command{
		Name:       "iq",
		Help:       "Tells you someone's IQ",
		Usage:      "$iq [user]",
		Permission: permission.Normal,
		PrefixOnly: true,
		Pattern:    iqCommandPattern,
		Handler:    iq,
	}
	iqPattern = regexp.MustCompile(iqCommandPattern.String() + `(\w+)`)
)

func bibleVerse(msg *base.IncomingMessage) ([]*base.Message, error) {
	matches := bibleVersePattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) < 2 {
		return nil, nil
	}

	verseQuery := matches[1]
	if verseQuery == "" {
		return nil, nil
	}

	verse, err := bible.FetchVerse(verseQuery)
	if err != nil {
		log.Printf("Failed to look up bible verse: %v", err)
		return nil, nil
	}

	verseText := strings.ReplaceAll(verse.Text, "\n", " ")

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("[%s %d:%d]: %s", verse.BookName, verse.Chapter, verse.Verse, verseText),
		},
	}, nil
}

const cockMaxLength = 14

var RandReader = rand.Reader

func cock(msg *base.IncomingMessage) ([]*base.Message, error) {
	target := basecommand.ParseTarget(msg, cockPattern)

	length, err := rand.Int(RandReader, big.NewInt(cockMaxLength+1))
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

var RandSource exprand.Source = nil

func iq(msg *base.IncomingMessage) ([]*base.Message, error) {
	target := basecommand.ParseTarget(msg, iqPattern)

	userIqFloat := distuv.Normal{Mu: 100, Sigma: 15, Src: RandSource}.Rand()
	userIq := int64(math.Round(userIqFloat))

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("%s's IQ is %d", target, userIq),
		},
	}, nil
}
