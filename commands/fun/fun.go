// Package fun implements fun commands that hit APIs.
package fun

import (
	"fmt"
	"log"
	"regexp"

	"github.com/airforce270/airbot/apiclients/bible"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/permission"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	bibleVerseCommand,
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

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("[%s %d:%d]: %s", verse.BookName, verse.Chapter, verse.Verse, verse.Text),
		},
	}, nil
}
