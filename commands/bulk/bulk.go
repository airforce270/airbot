// Package bulk handles commands that perform bulk operations.
package bulk

import (
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
	filesayCommand = basecommand.Command{
		Name:       "filesay",
		Help:       "Runs all commands in a given pastebin file.",
		Args:       []basecommand.Argument{{Name: "pastebin raw URL", Required: true}},
		Permission: permission.Mod,
		Handler:    filesay,
	}
)

func filesay(msg *base.IncomingMessage, args []string) ([]*base.Message, error) {
	if len(args) == 0 {
		return nil, basecommand.ErrBadUsage
	}
	pastebinURL := args[0]

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
