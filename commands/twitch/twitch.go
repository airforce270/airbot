// Package twitch implements Twitch commands.
package twitch

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/airforce270/airbot/apiclients/ivr"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/message"
	"github.com/airforce270/airbot/permission"
	twitchplatform "github.com/airforce270/airbot/platforms/twitch"
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
	vipsCommand,
}

const maxUsersPerMessage = 15

var (
	banReasonCommandPattern = basecommand.PrefixPattern("(?:banreason|br)")
	banReasonCommand        = basecommand.Command{
		Name:       "banreason",
		Help:       "Replies with the reason someone was banned on Twitch.",
		Pattern:    banReasonCommandPattern,
		Handler:    banReason,
		PrefixOnly: true,
		Permission: permission.Normal,
	}
	banReasonPattern = regexp.MustCompile(banReasonCommandPattern.String() + `(\w+).*`)

	currentGameCommandPattern = basecommand.PrefixPattern("currentgame")
	currentGameCommand        = basecommand.Command{
		Name:       "currentgame",
		Help:       "Replies with the game that's currently being streamed on a channel.",
		Pattern:    currentGameCommandPattern,
		Handler:    currentGame,
		PrefixOnly: true,
		Permission: permission.Normal,
	}
	currentGamePattern = regexp.MustCompile(currentGameCommandPattern.String() + `(\w+).*`)

	foundersCommandPattern = basecommand.PrefixPattern("founders")
	foundersCommand        = basecommand.Command{
		Name:       "founders",
		Help:       "Replies with a channel's founders.",
		Pattern:    foundersCommandPattern,
		Handler:    founders,
		PrefixOnly: true,
		Permission: permission.Normal,
	}
	foundersPattern = regexp.MustCompile(foundersCommandPattern.String() + `(\w+).*`)

	logsCommandPattern = basecommand.PrefixPattern("logs")
	logsCommand        = basecommand.Command{
		Name:       "logs",
		Help:       "Replies with a link to a Twitch user's logs in a channel.",
		Usage:      "$logs <channel> <user>",
		Permission: permission.Normal,
		PrefixOnly: true,
		Pattern:    logsCommandPattern,
		Handler:    logs,
	}
	logsPattern = regexp.MustCompile(logsCommandPattern.String() + `(\w+)\s+(\w+).*`)

	modsCommandPattern = basecommand.PrefixPattern("mods")
	modsCommand        = basecommand.Command{
		Name:       "mods",
		Help:       "Replies with a channel's mods.",
		Usage:      "$mods [user]",
		Permission: permission.Normal,
		PrefixOnly: true,
		Pattern:    modsCommandPattern,
		Handler:    mods,
	}
	modsPattern = regexp.MustCompile(modsCommandPattern.String() + `(\w+).*`)

	nameColorCommandPattern = basecommand.PrefixPattern("namecolor")
	nameColorCommand        = basecommand.Command{
		Name:       "namecolor",
		Help:       "Replies with a user's name color.",
		Usage:      "$namecolor [user]",
		Permission: permission.Normal,
		PrefixOnly: true,
		Pattern:    nameColorCommandPattern,
		Handler:    nameColor,
	}
	nameColorPattern = regexp.MustCompile(nameColorCommandPattern.String() + `(\w+).*`)

	titleCommandPattern = basecommand.PrefixPattern("title")
	titleCommand        = basecommand.Command{
		Name:       "title",
		Help:       "Replies with a channel's title.",
		Usage:      "$title [user]",
		Permission: permission.Normal,
		PrefixOnly: true,
		Pattern:    titleCommandPattern,
		Handler:    title,
	}
	titlePattern = regexp.MustCompile(titleCommandPattern.String() + `(\w+).*`)

	verifiedBotCommandPattern = basecommand.PrefixPattern("(?:verifiedbot|vb)")
	verifiedBotCommand        = basecommand.Command{
		Name:           "verifiedbot",
		AlternateNames: []string{"vb", "something"},
		Help:           "Replies whether a user is a verified bot.",
		Usage:          "$verifiedbot [user]",
		Permission:     permission.Normal,
		PrefixOnly:     true,
		Pattern:        verifiedBotCommandPattern,
		Handler:        verifiedBot,
	}
	verifiedBotPattern = regexp.MustCompile(verifiedBotCommandPattern.String() + `(\w+).*`)

	vipsCommandPattern = basecommand.PrefixPattern("vips")
	vipsCommand        = basecommand.Command{
		Name:       "vips",
		Help:       "Replies with a channel's VIPs.",
		Usage:      "$vips [user]",
		Permission: permission.Normal,
		PrefixOnly: true,
		Pattern:    vipsCommandPattern,
		Handler:    vips,
	}
	vipsPattern = regexp.MustCompile(vipsCommandPattern.String() + `(\w+).*`)
)

func banReason(msg *message.IncomingMessage) ([]*message.Message, error) {
	targetUser := basecommand.ParseTarget(msg, banReasonPattern)

	users, err := ivr.FetchUsers(targetUser)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return []*message.Message{
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

	return []*message.Message{
		{
			Channel: msg.Message.Channel,
			Text:    resp,
		},
	}, nil
}

func currentGame(msg *message.IncomingMessage) ([]*message.Message, error) {
	targetChannel := basecommand.ParseTarget(msg, currentGamePattern)

	tw := twitchplatform.Instance
	if tw == nil {
		return nil, fmt.Errorf("twitch platform connection not initialized")
	}

	channel, err := tw.Channel(targetChannel)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve channel info for %s: %w", targetChannel, err)
	}

	if channel.GameName == "" {
		return []*message.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("%s is not currenly playing anything", channel.BroadcasterName),
			},
		}, nil
	}

	return []*message.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("%s is currenly playing %s", channel.BroadcasterName, channel.GameName),
		},
	}, nil
}

func founders(msg *message.IncomingMessage) ([]*message.Message, error) {
	targetChannel := basecommand.ParseTarget(msg, foundersPattern)

	founders, err := ivr.FetchFounders(targetChannel)
	if err != nil {
		if strings.Contains(err.Error(), "Specified user has no founders.") {
			return []*message.Message{
				{
					Channel: msg.Message.Channel,
					Text:    fmt.Sprintf("%s has no founders", targetChannel),
				},
			}, nil
		}

		return nil, err
	}

	if len(founders.Founders) == 0 {
		return []*message.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("%s has no founders", targetChannel),
			},
		}, nil
	}

	foundersGroups := chunkBy(founders.Founders, maxUsersPerMessage)

	var messages []*message.Message

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
		messages = append(messages, &message.Message{Channel: msg.Message.Channel, Text: text})
	}

	return messages, nil
}

func logs(msg *message.IncomingMessage) ([]*message.Message, error) {
	matches := logsPattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) != 3 {
		return []*message.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Usage: %slogs <channel> <user>", msg.Prefix),
			},
		}, nil
	}
	targetChannel := strings.ToLower(matches[1])
	targetUser := strings.ToLower(matches[2])

	return []*message.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("%s's logs in %s's chat: https://logs.ivr.fi/?channel=%s&username=%s", targetUser, targetChannel, targetChannel, targetUser),
		},
	}, nil
}

func mods(msg *message.IncomingMessage) ([]*message.Message, error) {
	targetChannel := basecommand.ParseTarget(msg, modsPattern)

	modsAndVIPs, err := ivr.FetchModsAndVIPs(targetChannel)
	if err != nil {
		return nil, err
	}

	if len(modsAndVIPs.Mods) == 0 {
		return []*message.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("%s has no mods", targetChannel),
			},
		}, nil
	}

	modGroups := chunkBy(modsAndVIPs.Mods, maxUsersPerMessage)

	var messages []*message.Message

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
		messages = append(messages, &message.Message{Channel: msg.Message.Channel, Text: text})
	}

	return messages, nil
}

func nameColor(msg *message.IncomingMessage) ([]*message.Message, error) {
	targetUser := basecommand.ParseTarget(msg, nameColorPattern)

	users, err := ivr.FetchUsers(targetUser)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return []*message.Message{
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

	return []*message.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("%s's name color is %s", user.DisplayName, user.ChatColor),
		},
	}, nil
}

func title(msg *message.IncomingMessage) ([]*message.Message, error) {
	targetChannel := basecommand.ParseTarget(msg, titlePattern)

	tw := twitchplatform.Instance
	if tw == nil {
		return nil, fmt.Errorf("twitch platform connection not initialized")
	}

	channel, err := tw.Channel(targetChannel)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve channel info for %s: %w", targetChannel, err)
	}

	return []*message.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("%s's title: %s", channel.BroadcasterName, channel.Title),
		},
	}, nil
}

func verifiedBot(msg *message.IncomingMessage) ([]*message.Message, error) {
	targetUser := basecommand.ParseTarget(msg, verifiedBotPattern)

	users, err := ivr.FetchUsers(targetUser)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return []*message.Message{
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

	return []*message.Message{
		{
			Channel: msg.Message.Channel,
			Text:    resp,
		},
	}, nil
}

func vips(msg *message.IncomingMessage) ([]*message.Message, error) {
	targetChannel := basecommand.ParseTarget(msg, vipsPattern)

	modsAndVIPs, err := ivr.FetchModsAndVIPs(targetChannel)
	if err != nil {
		return nil, err
	}

	if len(modsAndVIPs.VIPs) == 0 {
		return []*message.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("%s has no VIPs", targetChannel),
			},
		}, nil
	}

	vipGroups := chunkBy(modsAndVIPs.VIPs, maxUsersPerMessage)

	var messages []*message.Message
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
		messages = append(messages, &message.Message{Channel: msg.Message.Channel, Text: text})
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

func chunkBy[T any](items []T, chunkSize int) (chunks [][]T) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	return append(chunks, items)
}
