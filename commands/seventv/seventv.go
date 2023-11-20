// Package seventv implements 7TV commands.
package seventv

import (
	"errors"
	"fmt"
	"log"

	seventvclient "github.com/airforce270/airbot/apiclients/seventv"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/base/arg"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/permission"
	"github.com/airforce270/airbot/platforms/twitch"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	emoteCountCommand,
}

var (
	emoteCountCommand = basecommand.Command{
		Name:       "7tv emotecount",
		Desc:       "Counts emotes in a channel.",
		Params:     []arg.Param{{Name: "channel", Type: arg.Username, Required: false}},
		Permission: permission.Normal,
		Handler:    emoteCount,
	}
)

func emoteCount(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	target := basecommand.FirstArgOrUsername(args, msg)

	user, err := twitch.Instance().FetchUser(target)
	if err != nil {
		if errors.Is(err, twitch.ErrChannelNotFound) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    fmt.Sprintf("Channel %s not found", target),
				},
			}, nil
		}
		return nil, err
	}

	resp, err := seventvclient.FetchUserConnectionByTwitchUserId(user.ID)
	if err != nil {
		log.Printf("Failed to fetch 7TV user connection: %v", err)
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Looking up %s on 7TV failed", user.DisplayName),
			},
		}, nil
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("%s has %d emotes on 7TV", target, len(resp.EmoteSet.Emotes)),
		},
	}, nil
}
