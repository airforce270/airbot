// Package twitch handles Twitch-specific logic.
package twitch

import (
	"fmt"
	"strings"
	"time"

	"airbot/config"
	"airbot/database/model"
	"airbot/logs"
	"airbot/message"

	twitchirc "github.com/gempir/go-twitch-irc/v3"
	"github.com/nicklaw5/helix/v2"
	"gorm.io/gorm"
)

// Instance is a connection to Twitch.
var Instance *Twitch

// Twitch implements Platform for a connection to Twitch chat.
type Twitch struct {
	// username is the Twitch username the bot is running as.
	username string
	// isVerifiedBot is whether the user running as is a verified bot on Twitch.
	// See https://dev.twitch.tv/docs/irc#verified-bots
	isVerifiedBot bool
	// channels is the Twitch channels to join.
	channels []config.TwitchChannelConfig
	// prefixes is a map of channel names to prefixes.
	prefixes map[string]string
	// clientID is the OAuth Client ID to use when connecting.
	clientID string
	// accessToken is the OAuth token to use when connecting.
	// See https://dev.twitch.tv/docs/irc/authenticate-bot#getting-an-access-token
	accessToken string
	// i is the Twitch IRC client.
	i *twitchirc.Client
	// h is the Twitch API client.
	h *helix.Client
	// db is a a reference to the database connection.
	db *gorm.DB
}

func (t *Twitch) Name() string { return "Twitch" }

func (t *Twitch) Username() string { return t.username }

func (t *Twitch) Send(m message.Message) error {
	go t.i.Say(m.Channel, m.Text)
	return nil
}

func (t *Twitch) Listen() chan message.IncomingMessage {
	c := make(chan message.IncomingMessage)
	t.i.OnPrivateMessage(func(msg twitchirc.PrivateMessage) {
		go t.persistUserAndMessage(msg.User.ID, msg.User.DisplayName, msg.Message, msg.Channel, msg.Time)
		c <- message.IncomingMessage{
			Message: message.Message{
				Text:    msg.Message,
				Channel: msg.Channel,
				User:    msg.User.Name,
				Time:    msg.Time,
			},
			Prefix: t.prefix(msg.Channel),
		}
	})
	return c
}

func (t *Twitch) Connect() error {
	logs.Printf("Creating Twitch IRC client...")
	i := twitchirc.NewClient(t.username, fmt.Sprintf("oauth:%s", t.accessToken))
	t.i = i

	if t.isVerifiedBot {
		t.i.SetJoinRateLimiter(twitchirc.CreateVerifiedRateLimiter())
	}

	for _, channel := range t.channels {
		logs.Printf("Joining Twitch channel %s...", channel.Name)
		i.Join(channel.Name)
	}

	logs.Printf("Connecting to Twitch IRC...")
	go func() {
		if err := t.i.Connect(); err != nil {
			panic(fmt.Sprintf("failed to connect to twitch IRC: %v", err))
		}
	}()

	logs.Printf("Connecting to Twitch API...")
	h, err := helix.NewClient(&helix.Options{
		ClientID:        t.clientID,
		UserAccessToken: t.accessToken,
	})
	if err != nil {
		return err
	}
	t.h = h

	return nil
}

func (t *Twitch) Disconnect() error {
	for _, channel := range t.channels {
		logs.Printf("Leaving Twitch channel %s...", channel.Name)
		t.i.Depart(channel.Name)
	}
	logs.Printf("Disconnecting from Twitch IRC...")
	return t.i.Disconnect()
}

func (t *Twitch) User(channel string) (*helix.User, error) {
	users, err := t.h.GetUsers(&helix.UsersParams{Logins: []string{channel}})
	if err != nil {
		return nil, err
	}
	if users.StatusCode != 200 {
		return nil, fmt.Errorf("twitch GetUsers call for %q failed, resp:%v", channel, users)
	}
	if len(users.Data.Users) != 1 {
		return nil, fmt.Errorf("wrong number of users returned (should be 1): %v", err)
	}
	return &users.Data.Users[0], nil
}

func (t *Twitch) Channel(channel string) (*helix.ChannelInformation, error) {
	user, err := t.User(channel)
	if err != nil {
		return nil, err
	}
	resp, err := t.h.GetChannelInformation(&helix.GetChannelInformationParams{BroadcasterIDs: []string{user.ID}})
	if err != nil {
		return nil, err
	}
	logs.Printf("channels:%v", resp.Data.Channels)
	if len(resp.Data.Channels) == 0 {
		return nil, fmt.Errorf("no channels found for %s", channel)
	}
	if len(resp.Data.Channels) > 1 {
		logs.Printf("more than one channel found for %s, using the first", channel)
	}
	return &resp.Data.Channels[0], nil
}

func (t *Twitch) prefix(channel string) string {
	p, ok := t.prefixes[strings.ToLower(channel)]
	if !ok {
		logs.Printf("No prefix found for channel %s", channel)
		return ""
	}
	return p
}

func (t *Twitch) persistUserAndMessage(twitchID, twitchName, message, channel string, sentTime time.Time) {
	var user model.User
	t.db.FirstOrCreate(&user, model.User{TwitchID: twitchID})
	t.db.Model(&user).Updates(model.User{TwitchName: twitchName})
	t.db.Create(&model.Message{
		Text:    message,
		Channel: channel,
		User:    user,
		Time:    sentTime,
	})
}

// New creates a new Twitch connection.
func New(username string, channels []config.TwitchChannelConfig, clientID, accessToken string, isVerifiedBot bool, db *gorm.DB) *Twitch {
	return &Twitch{
		username:      username,
		isVerifiedBot: isVerifiedBot,
		channels:      channels,
		prefixes:      *buildPrefixes(channels),
		clientID:      clientID,
		accessToken:   accessToken,
		db:            db,
	}
}

func buildPrefixes(channels []config.TwitchChannelConfig) *map[string]string {
	prefixes := map[string]string{}
	for _, channel := range channels {
		prefixes[strings.ToLower(channel.Name)] = channel.Prefix
	}
	return &prefixes
}
