// Package twitch implements Twitch commands.
package twitch

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/airforce270/airbot/apiclients/ivr"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/message"
	twitchplatform "github.com/airforce270/airbot/platforms/twitch"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	banReasonCommand,
	currentGameCommand,
	foundersCommand,
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
		Pattern:    banReasonCommandPattern,
		Handle:     banReason,
		PrefixOnly: true,
	}
	banReasonPattern = regexp.MustCompile(banReasonCommandPattern.String() + `(\w+).*`)

	currentGameCommandPattern = basecommand.PrefixPattern("currentgame")
	currentGameCommand        = basecommand.Command{
		Pattern:    currentGameCommandPattern,
		Handle:     currentGame,
		PrefixOnly: true,
	}
	currentGamePattern = regexp.MustCompile(currentGameCommandPattern.String() + `(\w+).*`)

	foundersCommandPattern = basecommand.PrefixPattern("founders")
	foundersCommand        = basecommand.Command{
		Pattern:    foundersCommandPattern,
		Handle:     founders,
		PrefixOnly: true,
	}
	foundersPattern = regexp.MustCompile(foundersCommandPattern.String() + `(\w+).*`)

	modsCommandPattern = basecommand.PrefixPattern("mods")
	modsCommand        = basecommand.Command{
		Pattern:    modsCommandPattern,
		Handle:     mods,
		PrefixOnly: true,
	}
	modsPattern = regexp.MustCompile(modsCommandPattern.String() + `(\w+).*`)

	nameColorCommandPattern = basecommand.PrefixPattern("namecolor")
	nameColorCommand        = basecommand.Command{
		Pattern:    nameColorCommandPattern,
		Handle:     nameColor,
		PrefixOnly: true,
	}
	nameColorPattern = regexp.MustCompile(nameColorCommandPattern.String() + `(\w+).*`)

	titleCommandPattern = basecommand.PrefixPattern("title")
	titleCommand        = basecommand.Command{
		Pattern:    titleCommandPattern,
		Handle:     title,
		PrefixOnly: true,
	}
	titlePattern = regexp.MustCompile(titleCommandPattern.String() + `(\w+).*`)

	verifiedBotCommandPattern = basecommand.PrefixPattern("(?:verifiedbot|vb)")
	verifiedBotCommand        = basecommand.Command{
		Pattern:    verifiedBotCommandPattern,
		Handle:     verifiedBot,
		PrefixOnly: true,
	}
	verifiedBotPattern = regexp.MustCompile(verifiedBotCommandPattern.String() + `(\w+).*`)

	vipsCommandPattern = basecommand.PrefixPattern("vips")
	vipsCommand        = basecommand.Command{
		Pattern:    vipsCommandPattern,
		Handle:     vips,
		PrefixOnly: true,
	}
	vipsPattern = regexp.MustCompile(vipsCommandPattern.String() + `(\w+).*`)
)

func banReason(msg *message.IncomingMessage) ([]*message.Message, error) {
	matches := banReasonPattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) <= 1 {
		return nil, fmt.Errorf("no channel provided")
	}
	targetChannel := strings.ToLower(matches[1])

	user, err := ivr.FetchUser(targetChannel)
	if err != nil {
		return nil, err
	}

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
	matches := currentGamePattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) <= 1 {
		return nil, fmt.Errorf("no channel provided")
	}
	targetChannel := matches[1]

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
			Text:    fmt.Sprintf("%s is currenly playing %s", channel.BroadcasterName, channel.GameName),
		},
	}, nil
}

func founders(msg *message.IncomingMessage) ([]*message.Message, error) {
	matches := foundersPattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) <= 1 {
		return nil, fmt.Errorf("no channel provided")
	}
	targetChannel := strings.ToLower(matches[1])

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

func mods(msg *message.IncomingMessage) ([]*message.Message, error) {
	matches := modsPattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) <= 1 {
		return nil, fmt.Errorf("no channel provided")
	}
	targetChannel := strings.ToLower(matches[1])

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
	matches := nameColorPattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) <= 1 {
		return nil, fmt.Errorf("no channel provided")
	}
	targetChannel := strings.ToLower(matches[1])

	user, err := ivr.FetchUser(targetChannel)
	if err != nil {
		return nil, err
	}

	return []*message.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("%s's name color is %s", user.DisplayName, user.ChatColor),
		},
	}, nil
}

func title(msg *message.IncomingMessage) ([]*message.Message, error) {
	matches := titlePattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) <= 1 {
		return nil, fmt.Errorf("no channel provided")
	}
	targetChannel := matches[1]

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
	matches := verifiedBotPattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) <= 1 {
		return nil, fmt.Errorf("no channel provided")
	}
	targetChannel := strings.ToLower(matches[1])

	user, err := ivr.FetchUser(targetChannel)
	if err != nil {
		return nil, err
	}

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
	matches := vipsPattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) <= 1 {
		return nil, fmt.Errorf("no channel provided")
	}
	targetChannel := strings.ToLower(matches[1])

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
