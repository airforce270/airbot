// Package bulk handles commands that perform bulk operations.
package bulk

import (
	"github.com/airforce270/airbot/apiclients/pastebin"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/base/arg"
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
		Desc:       "Runs all commands in a given pastebin file.",
		Params:     []arg.Param{{Name: "pastebin raw URL", Type: arg.String, Required: true}},
		Permission: permission.Mod,
		Handler:    filesay,
	}
)

func filesay(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	pastebinURLArg := args[0]
	if !pastebinURLArg.Present {
		return nil, basecommand.ErrBadUsage
	}

	paste, err := pastebin.FetchPaste(pastebinURLArg.StringValue)
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
