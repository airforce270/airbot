// Package bulk handles commands that perform bulk operations.
package bulk

import (
	"fmt"

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

	client := pastebin.NewClient(msg.Resources.Clients.PastebinFetchPasteURLOverride)
	paste, err := client.FetchPaste(pastebinURLArg.StringValue)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch paste %s: %w", pastebinURLArg.StringValue, err)
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
