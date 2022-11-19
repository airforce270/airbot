// Package twitch implements Twitch commands.
package twitch

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/airforce270/airbot/apiclients/ivr"
	"github.com/airforce270/airbot/base"
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
	banReasonCommandPattern = basecommand.PrefixPattern("(?:banreason|br)")
	banReasonCommand        = basecommand.Command{
		Name:       "banreason",
		Aliases:    []string{"br"},
		Help:       "Replies with the reason someone was banned on Twitch.",
		Usage:      "$banreason <user>",
		PrefixOnly: true,
		Permission: permission.Normal,
		Pattern:    banReasonCommandPattern,
		Handler:    banReason,
	}
	banReasonPattern = regexp.MustCompile(banReasonCommandPattern.String() + `@?(\w+).*`)

	currentGameCommandPattern = basecommand.PrefixPattern("currentgame")
	currentGameCommand        = basecommand.Command{
		Name:       "currentgame",
		Help:       "Replies with the game that's currently being streamed on a channel.",
		Usage:      "$currentgame <channel>",
		PrefixOnly: true,
		Permission: permission.Normal,
		Pattern:    currentGameCommandPattern,
		Handler:    currentGame,
	}
	currentGamePattern = regexp.MustCompile(currentGameCommandPattern.String() + `@?(\w+).*`)

	foundersCommandPattern = basecommand.PrefixPattern("founders")
	foundersCommand        = basecommand.Command{
		Name:       "founders",
		Help:       "Replies with a channel's founders. If no channel is provided, the current channel will be used.",
		Usage:      "$founders [channel]",
		PrefixOnly: true,
		Permission: permission.Normal,
		Pattern:    foundersCommandPattern,
		Handler:    founders,
	}
	foundersPattern = regexp.MustCompile(foundersCommandPattern.String() + `@?(\w+).*`)

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
	logsPattern = regexp.MustCompile(logsCommandPattern.String() + `@?(\w+)\s+@?(\w+).*`)

	modsCommandPattern = basecommand.PrefixPattern("mods")
	modsCommand        = basecommand.Command{
		Name:       "mods",
		Help:       "Replies with a channel's mods. If no channel is provided, the current channel will be used.",
		Usage:      "$mods [channel]",
		Permission: permission.Normal,
		PrefixOnly: true,
		Pattern:    modsCommandPattern,
		Handler:    mods,
	}
	modsPattern = regexp.MustCompile(modsCommandPattern.String() + `@?(\w+).*`)

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
	nameColorPattern = regexp.MustCompile(nameColorCommandPattern.String() + `@?(\w+).*`)

	titleCommandPattern = basecommand.PrefixPattern("title")
	titleCommand        = basecommand.Command{
		Name:       "title",
		Help:       "Replies with a channel's title. If no channel is provided, the current channel will be used.",
		Usage:      "$title [channel]",
		Permission: permission.Normal,
		PrefixOnly: true,
		Pattern:    titleCommandPattern,
		Handler:    title,
	}
	titlePattern = regexp.MustCompile(titleCommandPattern.String() + `@?(\w+).*`)

	verifiedBotCommandPattern = regexp.MustCompile(`\s*(?:verifiedbot|vb)(?:\s+|$)`)
	verifiedBotCommand        = basecommand.Command{
		Name:       "verifiedbot",
		Aliases:    []string{"vb"},
		Help:       "Replies whether a user is a verified bot.",
		Usage:      "$verifiedbot [user]",
		Permission: permission.Normal,
		PrefixOnly: true,
		Pattern:    verifiedBotCommandPattern,
		Handler:    verifiedBot,
	}
	verifiedBotPattern = regexp.MustCompile(verifiedBotCommandPattern.String() + `@?(\w+).*`)

	verifiedBotQuietCommandPattern = basecommand.PrefixPattern(`(?:verifiedbot|vb)(?:q(?:uiet)?)`)
	verifiedBotQuietCommand        = basecommand.Command{
		Name:       "verifiedbotquiet",
		Aliases:    []string{"vbq"},
		Help:       "Replies whether a user is a verified bot, but responds quietly.",
		Usage:      "$verifiedbotquiet [user]",
		Permission: permission.Normal,
		PrefixOnly: true,
		Pattern:    verifiedBotQuietCommandPattern,
		Handler:    verifiedBotQuiet,
	}
	verifiedBotQuietPattern = regexp.MustCompile(verifiedBotQuietCommandPattern.String() + `@?(\w+).*`)

	vipsCommandPattern = basecommand.PrefixPattern("vips")
	vipsCommand        = basecommand.Command{
		Name:       "vips",
		Help:       "Replies with a channel's VIPs. If no channel is provided, the current channel will be used.",
		Usage:      "$vips [channel]",
		Permission: permission.Normal,
		PrefixOnly: true,
		Pattern:    vipsCommandPattern,
		Handler:    vips,
	}
	vipsPattern = regexp.MustCompile(vipsCommandPattern.String() + `@?(\w+).*`)
)

func banReason(msg *base.IncomingMessage) ([]*base.Message, error) {
	targetUser := basecommand.ParseTarget(msg, banReasonPattern)

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

func currentGame(msg *base.IncomingMessage) ([]*base.Message, error) {
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

func founders(msg *base.IncomingMessage) ([]*base.Message, error) {
	targetChannel := basecommand.ParseTargetWithDefault(msg, foundersPattern, msg.Message.Channel)

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

func logs(msg *base.IncomingMessage) ([]*base.Message, error) {
	matches := logsPattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) != 3 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("Usage: %slogs <channel> <user>", msg.Prefix),
			},
		}, nil
	}
	targetChannel := strings.ToLower(matches[1])
	targetUser := strings.ToLower(matches[2])

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("%s's logs in %s's chat: https://logs.ivr.fi/?channel=%s&username=%s", targetUser, targetChannel, targetChannel, targetUser),
		},
	}, nil
}

func mods(msg *base.IncomingMessage) ([]*base.Message, error) {
	targetChannel := basecommand.ParseTargetWithDefault(msg, modsPattern, msg.Message.Channel)

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

func nameColor(msg *base.IncomingMessage) ([]*base.Message, error) {
	targetUser := basecommand.ParseTarget(msg, nameColorPattern)

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

func title(msg *base.IncomingMessage) ([]*base.Message, error) {
	targetChannel := basecommand.ParseTargetWithDefault(msg, titlePattern, msg.Message.Channel)

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

func verifiedBot(msg *base.IncomingMessage) ([]*base.Message, error) {
	targetUser := basecommand.ParseTarget(msg, verifiedBotPattern)

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

func verifiedBotQuiet(msg *base.IncomingMessage) ([]*base.Message, error) {
	targetUser := basecommand.ParseTarget(msg, verifiedBotQuietPattern)

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

func vips(msg *base.IncomingMessage) ([]*base.Message, error) {
	targetChannel := basecommand.ParseTargetWithDefault(msg, vipsPattern, msg.Message.Channel)

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
