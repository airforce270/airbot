// Package twitch handles Twitch-specific logic.
package twitch

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/airforce270/airbot/apiclients/ivr"
	"github.com/airforce270/airbot/config"
	"github.com/airforce270/airbot/database/model"
	"github.com/airforce270/airbot/message"

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
	// id is the Twitch ID of the account the bot is running as.
	id string
	// isVerifiedBot is whether the user running as is a verified bot on Twitch.
	// See https://dev.twitch.tv/docs/irc#verified-bots
	isVerifiedBot bool
	// channels is the Twitch channels to join.
	channels []*twitchChannel
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

func (t *Twitch) Send(msg message.Message) error {
	var channel *twitchChannel
	for _, c := range t.channels {
		if strings.EqualFold(c.Name, msg.Channel) {
			channel = c
		}
	}
	if channel == nil {
		return fmt.Errorf("can't send message to unjoined channel %q", msg.Channel)
	}

	// If the bot has "normal permissions" (not verified, mod, or VIP),
	// sending the message too quickly will get it held back.
	if !t.isVerifiedBot && !channel.BotIsModerator && !channel.BotIsVIP {
		time.Sleep(time.Millisecond * 100)
	}

	go t.persistUserAndMessage(t.id, t.username, msg.Text, msg.Channel, msg.Time)
	t.i.Say(msg.Channel, msg.Text)
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
	log.Printf("Creating Twitch IRC client...")
	i := twitchirc.NewClient(t.username, fmt.Sprintf("oauth:%s", t.accessToken))
	t.i = i

	t.i.OnUserNoticeMessage(func(msg twitchirc.UserNoticeMessage) { log.Printf("[Twitch] USERNOTICE: %s", msg.Raw) })
	t.i.OnNoticeMessage(func(msg twitchirc.NoticeMessage) { log.Printf("[Twitch] NOTICE: %s", msg.Raw) })

	ivrUsers, err := ivr.FetchUsers(t.username)
	if err != nil {
		fmt.Printf("Failed to fetch info about %s from IVR, assuming not a verified bot: %v", t.username, err)
		t.isVerifiedBot = false
	} else if len(ivrUsers) != 1 {
		fmt.Printf("IVR API returned %d users for %s, assuming not a verified bot: %v", len(ivrUsers), t.username, ivrUsers)
		t.isVerifiedBot = false
	} else {
		t.isVerifiedBot = ivrUsers[0].IsVerifiedBot
	}

	if t.isVerifiedBot {
		log.Printf("[Twitch] Bot user %s is a verified bot, using increased rate limit", t.username)
		t.i.SetJoinRateLimiter(twitchirc.CreateVerifiedRateLimiter())
	}

	log.Printf("Connecting to Twitch IRC...")
	twitchIRCReady := make(chan bool)
	t.i.OnConnect(func() { twitchIRCReady <- true })
	go func() {
		if err := t.i.Connect(); err != nil {
			panic(fmt.Sprintf("failed to connect to twitch IRC: %v", err))
		}
	}()
	<-twitchIRCReady

	for _, channel := range t.channels {
		log.Printf("Joining Twitch channel %s...", channel.Name)
		i.Join(channel.Name)
	}

	log.Printf("Connecting to Twitch API...")
	h, err := helix.NewClient(&helix.Options{
		ClientID:        t.clientID,
		UserAccessToken: t.accessToken,
	})
	if err != nil {
		return err
	}
	t.h = h

	botUser, err := t.User(t.username)
	if err != nil {
		return err
	}
	if botUser == nil {
		return fmt.Errorf("no user returned for bot (%s): %w", t.username, err)
	}
	t.id = botUser.ID

	go t.listenForModAndVIPChanges()

	return nil
}

func (t *Twitch) Disconnect() error {
	for _, channel := range t.channels {
		log.Printf("Leaving Twitch channel %s...", channel.Name)
		t.i.Depart(channel.Name)
	}
	log.Printf("Disconnecting from Twitch IRC...")
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
	if len(resp.Data.Channels) == 0 {
		return nil, fmt.Errorf("no channels found for %s", channel)
	}
	if len(resp.Data.Channels) > 1 {
		log.Printf("more than one channel found for %s, using the first", channel)
	}
	return &resp.Data.Channels[0], nil
}

func (t *Twitch) prefix(channel string) string {
	p, ok := t.prefixes[strings.ToLower(channel)]
	if !ok {
		log.Printf("No prefix found for channel %s", channel)
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

func (t *Twitch) listenForModAndVIPChanges() {
	for {
		for _, channel := range t.channels {
			modsAndVIPs, err := ivr.FetchModsAndVIPs(channel.Name)
			if err != nil {
				log.Printf("Failed to look up mods and VIPs for %s: %v", channel.Name, err)
				break
			}
			go t.updateModStatusForChannel(channel, modsAndVIPs.Mods)
			go t.updateVIPStatusForChannel(channel, modsAndVIPs.VIPs)
		}

		time.Sleep(time.Second * 30)
	}
}

func (t *Twitch) updateModStatusForChannel(channel *twitchChannel, mods []*ivr.ModOrVIPUser) {
	for _, mod := range mods {
		if mod.ID == t.id || t.username == channel.Name {
			log.Printf("[Twitch] Determined bot is a mod in channel %q", channel.Name)
			channel.BotIsModerator = true
			return
		}
	}
	log.Printf("[Twitch] Determined bot is not a mod in channel %q", channel.Name)
	channel.BotIsModerator = false
}

func (t *Twitch) updateVIPStatusForChannel(channel *twitchChannel, vips []*ivr.ModOrVIPUser) {
	for _, vip := range vips {
		if vip.ID == t.id {
			log.Printf("[Twitch] Determined bot is a VIP in channel %q", channel.Name)
			channel.BotIsVIP = true
			return
		}
	}
	log.Printf("[Twitch] Determined bot is not a VIP in channel %q", channel.Name)
	channel.BotIsVIP = false
}

// New creates a new Twitch connection.
func New(username string, channels []config.TwitchChannelConfig, clientID, accessToken string, db *gorm.DB) *Twitch {
	return &Twitch{
		username:    username,
		channels:    buildChannels(channels),
		prefixes:    *buildPrefixes(channels),
		clientID:    clientID,
		accessToken: accessToken,
		db:          db,
	}
}

// New creates a new Twitch connection for testing.
func NewForTesting(url string) *Twitch {
	helixClient, err := helix.NewClient(&helix.Options{
		ClientID:        "fake-client-id",
		UserAccessToken: "fake-access-token",
		APIBaseURL:      url,
	})
	if err != nil {
		panic(err)
	}
	return &Twitch{
		username:    "fake-username",
		channels:    nil,
		prefixes:    nil,
		clientID:    "fake-client-id",
		accessToken: "fake-access-token",
		db:          nil,
		h:           helixClient,
	}
}

type twitchChannel struct {
	Name           string
	Prefix         string
	BotIsModerator bool
	BotIsVIP       bool
}

func buildChannels(channels []config.TwitchChannelConfig) []*twitchChannel {
	var builtChannels []*twitchChannel
	for _, channel := range channels {
		builtChannels = append(builtChannels, &twitchChannel{
			Name:           channel.Name,
			Prefix:         channel.Prefix,
			BotIsModerator: false,
			BotIsVIP:       false,
		})
	}
	return builtChannels
}

func buildPrefixes(channels []config.TwitchChannelConfig) *map[string]string {
	prefixes := map[string]string{}
	for _, channel := range channels {
		prefixes[strings.ToLower(channel.Name)] = channel.Prefix
	}
	return &prefixes
}
