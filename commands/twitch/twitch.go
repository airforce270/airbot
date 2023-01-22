// Package twitch implements Twitch commands.
package twitch

import (
	"fmt"
	"strings"

	"github.com/airforce270/airbot/apiclients/ivr"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/base/arg"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/permission"
	twitchplatform "github.com/airforce270/airbot/platforms/twitch"
	"github.com/airforce270/airbot/utils"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	banReasonCommand,
	currentGameCommand,
	foundersCommand,
	logsCommand,
	modsCommand,
	nameColorCommand,
	titleCommand,
	verifiedBotCommand,
	verifiedBotQuietCommand,
	vipsCommand,
}

const maxUsersPerMessage = 15

var (
	banReasonCommand = basecommand.Command{
		Name:       "banreason",
		Aliases:    []string{"br"},
		Help:       "Replies with the reason someone was banned on Twitch.",
		Params:     []arg.Param{{Name: "user", Type: arg.Username, Required: true}},
		Permission: permission.Normal,
		Handler:    banReason,
	}

	currentGameCommand = basecommand.Command{
		Name:       "currentgame",
		Help:       "Replies with the game that's currently being streamed on a channel.",
		Params:     []arg.Param{{Name: "channel", Type: arg.Username, Required: true}},
		Permission: permission.Normal,
		Handler:    currentGame,
	}

	foundersCommand = basecommand.Command{
		Name:       "founders",
		Help:       "Replies with a channel's founders. If no channel is provided, the current channel will be used.",
		Params:     []arg.Param{{Name: "channel", Type: arg.Username, Required: false}},
		Permission: permission.Normal,
		Handler:    founders,
	}

	logsCommand = basecommand.Command{
		Name: "logs",
		Help: "Replies with a link to a Twitch user's logs in a channel.",
		Params: []arg.Param{
			{Name: "channel", Type: arg.Username, Required: true},
			{Name: "user", Type: arg.Username, Required: true},
		},
		Permission: permission.Normal,
		Handler:    logs,
	}

	modsCommand = basecommand.Command{
		Name:       "mods",
		Help:       "Replies with a channel's mods. If no channel is provided, the current channel will be used.",
		Params:     []arg.Param{{Name: "channel", Type: arg.Username, Required: false}},
		Permission: permission.Normal,
		Handler:    mods,
	}

	nameColorCommand = basecommand.Command{
		Name:       "namecolor",
		Help:       "Replies with a user's name color.",
		Params:     []arg.Param{{Name: "user", Type: arg.Username, Required: false}},
		Permission: permission.Normal,
		Handler:    nameColor,
	}

	titleCommand = basecommand.Command{
		Name:       "title",
		Help:       "Replies with a channel's title. If no channel is provided, the current channel will be used.",
		Params:     []arg.Param{{Name: "channel", Type: arg.Username, Required: false}},
		Permission: permission.Normal,
		Handler:    title,
	}

	verifiedBotCommand = basecommand.Command{
		Name:       "verifiedbot",
		Aliases:    []string{"vb"},
		Help:       "Replies whether a user is a verified bot.",
		Params:     []arg.Param{{Name: "user", Type: arg.Username, Required: false}},
		Permission: permission.Normal,
		Handler:    verifiedBot,
	}

	verifiedBotQuietCommand = basecommand.Command{
		Name:       "verifiedbotquiet",
		Aliases:    []string{"verifiedbotq", "vbquiet", "vbq"},
		Help:       "Replies whether a user is a verified bot, but responds quietly.",
		Params:     []arg.Param{{Name: "user", Type: arg.Username, Required: false}},
		Permission: permission.Normal,
		Handler:    verifiedBotQuiet,
	}

	vipsCommand = basecommand.Command{
		Name:       "vips",
		Help:       "Replies with a channel's VIPs. If no channel is provided, the current channel will be used.",
		Params:     []arg.Param{{Name: "channel", Type: arg.Username, Required: false}},
		Permission: permission.Normal,
		Handler:    vips,
	}
)

func banReason(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	targetUserArg := args[0]
	if !targetUserArg.Present {
		return nil, basecommand.ErrBadUsage
	}
	targetUser := targetUserArg.StringValue

	users, err := ivr.FetchUsers(targetUser)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Couldn't find user %s", targetUser),
			},
		}, nil
	}
	if len(users) > 1 {
		return nil, fmt.Errorf("more than 1 user returned for %s: %v", targetUser, users)
	}
	user := users[0]

	var resp string
	if !user.IsBanned {
		resp = fmt.Sprintf("%s is not banned.", user.DisplayName)
	} else {
		resp = fmt.Sprintf("%s's ban reason: %s", user.DisplayName, user.BanReason)
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    resp,
		},
	}, nil
}

func currentGame(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	targetChannel := basecommand.FirstArgOrChannel(args, msg)

	tw := twitchplatform.Instance
	if tw == nil {
		return nil, fmt.Errorf("twitch platform connection not initialized")
	}

	channel, err := tw.Channel(targetChannel)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve channel info for %s: %w", targetChannel, err)
	}

	if channel.GameName == "" {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("%s is not currently playing anything", channel.BroadcasterName),
			},
		}, nil
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("%s is currently playing %s", channel.BroadcasterName, channel.GameName),
		},
	}, nil
}

func founders(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	targetChannel := basecommand.FirstArgOrChannel(args, msg)

	founders, err := ivr.FetchFounders(targetChannel)
	if err != nil {
		if strings.Contains(err.Error(), "Specified user has no founders.") {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    fmt.Sprintf("%s has no founders", targetChannel),
				},
			}, nil
		}

		return nil, err
	}

	if len(founders.Founders) == 0 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("%s has no founders", targetChannel),
			},
		}, nil
	}

	foundersGroups := utils.Chunk(founders.Founders, maxUsersPerMessage)

	var messages []*base.Message

	for i, foundersGroup := range foundersGroups {
		var text string
		if i == 0 {
			text = fmt.Sprintf("%s's founders are: %s", targetChannel, strings.Join(namesFromFounders(foundersGroup), ", "))
		} else {
			text = strings.Join(namesFromFounders(foundersGroup), ", ")
		}
		if len(foundersGroups) > 1 && len(foundersGroups)-1 != i {
			text += ","
		}
		messages = append(messages, &base.Message{Channel: msg.Message.Channel, Text: text})
	}

	return messages, nil
}

func logs(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	targetChannelArg, targetUserArg := args[0], args[1]
	if !targetChannelArg.Present || !targetUserArg.Present {
		return nil, basecommand.ErrBadUsage
	}
	targetChannel := strings.ToLower(targetChannelArg.StringValue)
	targetUser := strings.ToLower(targetUserArg.StringValue)

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("%s's logs in %s's chat: https://logs.ivr.fi/?channel=%s&username=%s", targetUser, targetChannel, targetChannel, targetUser),
		},
	}, nil
}

func mods(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	targetChannel := basecommand.FirstArgOrChannel(args, msg)

	modsAndVIPs, err := ivr.FetchModsAndVIPs(targetChannel)
	if err != nil {
		return nil, err
	}

	if len(modsAndVIPs.Mods) == 0 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("%s has no mods", targetChannel),
			},
		}, nil
	}

	modGroups := utils.Chunk(modsAndVIPs.Mods, maxUsersPerMessage)

	var messages []*base.Message

	for i, modGroup := range modGroups {
		var text string
		if i == 0 {
			text = fmt.Sprintf("%s's mods are: %s", targetChannel, strings.Join(namesFromModsOrVIPs(modGroup), ", "))
		} else {
			text = strings.Join(namesFromModsOrVIPs(modGroup), ", ")
		}
		if len(modGroups) > 1 && len(modGroups)-1 != i {
			text += ","
		}
		messages = append(messages, &base.Message{Channel: msg.Message.Channel, Text: text})
	}

	return messages, nil
}

func nameColor(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	targetUser := basecommand.FirstArgOrUsername(args, msg)

	users, err := ivr.FetchUsers(targetUser)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Couldn't find user %s", targetUser),
			},
		}, nil
	}
	if len(users) > 1 {
		return nil, fmt.Errorf("more than 1 user returned for %s: %v", targetUser, users)
	}
	user := users[0]

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("%s's name color is %s", user.DisplayName, user.ChatColor),
		},
	}, nil
}

func title(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	targetChannel := basecommand.FirstArgOrChannel(args, msg)

	tw := twitchplatform.Instance
	if tw == nil {
		return nil, fmt.Errorf("twitch platform connection not initialized")
	}

	channel, err := tw.Channel(targetChannel)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve channel info for %s: %w", targetChannel, err)
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("%s's title: %s", channel.BroadcasterName, channel.Title),
		},
	}, nil
}

func verifiedBot(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	targetUser := basecommand.FirstArgOrUsername(args, msg)

	users, err := ivr.FetchUsers(targetUser)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Couldn't find user %s", targetUser),
			},
		}, nil
	}
	if len(users) > 1 {
		return nil, fmt.Errorf("more than 1 user returned for %s: %v", targetUser, users)
	}
	user := users[0]

	var resp string
	if user.IsVerifiedBot {
		resp = fmt.Sprintf("%s is a verified bot. ✅", user.DisplayName)
	} else {
		resp = fmt.Sprintf("%s is not a verified bot. ❌", user.DisplayName)
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    resp,
		},
	}, nil
}

func verifiedBotQuiet(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	targetUser := basecommand.FirstArgOrUsername(args, msg)

	users, err := ivr.FetchUsers(targetUser)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Couldn't find user %s", targetUser),
			},
		}, nil
	}
	if len(users) > 1 {
		return nil, fmt.Errorf("more than 1 user returned for %s: %v", targetUser, users)
	}
	user := users[0]

	var resp string
	if user.IsVerifiedBot {
		resp = "✅"
	} else {
		resp = "❌"
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    resp,
		},
	}, nil
}

func vips(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	targetChannel := basecommand.FirstArgOrChannel(args, msg)

	modsAndVIPs, err := ivr.FetchModsAndVIPs(targetChannel)
	if err != nil {
		return nil, err
	}

	if len(modsAndVIPs.VIPs) == 0 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("%s has no VIPs", targetChannel),
			},
		}, nil
	}

	vipGroups := utils.Chunk(modsAndVIPs.VIPs, maxUsersPerMessage)

	var messages []*base.Message
	for i, vipGroup := range vipGroups {
		var text string
		if i == 0 {
			text = fmt.Sprintf("%s's VIPs are: %s", targetChannel, strings.Join(namesFromModsOrVIPs(vipGroup), ", "))
		} else {
			text = strings.Join(namesFromModsOrVIPs(vipGroup), ", ")
		}
		if len(vipGroups) > 1 && len(vipGroups)-1 != i {
			text += ","
		}
		messages = append(messages, &base.Message{Channel: msg.Message.Channel, Text: text})
	}

	return messages, nil
}

func namesFromFounders(users []*ivr.Founder) []string {
	var names []string
	for _, user := range users {
		names = append(names, user.DisplayName)
	}
	return names
}

func namesFromModsOrVIPs(users []*ivr.ModOrVIPUser) []string {
	var names []string
	for _, user := range users {
		names = append(names, user.DisplayName)
	}
	return names
}
