// Package bulk handles commands that perform bulk operations.
package bulk

import (
	"fmt"
	"regexp"

	"github.com/airforce270/airbot/apiclients/pastebin"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/permission"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	filesayCommand,
}

var (
	filesayCommandPattern = basecommand.PrefixPattern("filesay")
	filesayCommand        = basecommand.Command{
		Name:       "filesay",
		Help:       "Runs all commands in a given pastebin file.",
		Usage:      "$filesay <pastebin raw url>",
		Pattern:    filesayCommandPattern,
		Handler:    filesay,
		PrefixOnly: true,
		Permission: permission.Mod,
	}
	filesayPattern = regexp.MustCompile(filesayCommandPattern.String() + `(.+)`)
)

func filesay(msg *base.IncomingMessage) ([]*base.Message, error) {
	matches := filesayPattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) != 2 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("usage: %sfilesay <pastebin raw url>", msg.Prefix),
			},
		}, nil
	}
	pastebinURL := matches[1]

	paste, err := pastebin.FetchPaste(pastebinURL)
	if err != nil {
		return nil, err
	}

	msgs := make([]*base.Message, len(paste.Values()))
	for i, line := range paste.Values() {
		msgs[i] = &base.Message{
			Channel: msg.Message.Channel,
			Text:    line,
		}
	}
	return msgs, nil
}
