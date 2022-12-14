// Package twitch handles Twitch-specific logic.
package twitch

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/airforce270/airbot/apiclients/ivr"
	"github.com/airforce270/airbot/apiclients/twitchtmi"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/cache"
	"github.com/airforce270/airbot/cache/cachetest"
	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/database/models"
	"github.com/airforce270/airbot/permission"

	twitchirc "github.com/gempir/go-twitch-irc/v3"
	"github.com/nicklaw5/helix/v2"
	"golang.org/x/exp/slices"
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
	// owners contains the usernames of the bot's owners. Usually only one.
	owners []string
	// clientID is the OAuth Client ID to use when connecting.
	clientID string
	// accessToken is the OAuth token to use when connecting.
	// See https://dev.twitch.tv/docs/irc/authenticate-bot#getting-an-access-token
	accessToken string
	// i is the Twitch IRC client.
	i *twitchirc.Client
	// h is the Twitch API client.
	h *helix.Client
	// db is a reference to the database connection.
	db *gorm.DB
	// cdb is a reference to the cache.
	cdb cache.Cache
}

func (t *Twitch) Name() string { return "Twitch" }

func (t *Twitch) Username() string { return t.username }

func (t *Twitch) Send(msg base.Message) error {
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

	// Any newlines in the message causes Twitch to drop the rest of the message.
	text := strings.ReplaceAll(msg.Text, "\n", " ")

	// Bypass 30-second same message detection.
	lastSentMsg, err := t.cdb.FetchString(cache.KeyLastSentTwitchMessage)
	if err != nil {
		log.Printf("Failed to check if message (%q/%q) is in cache: %v", msg.Channel, text, err)
	} else if lastSentMsg == text {
		text = bypassSameMessageDetection(text)
	}

	go t.persistUserAndMessage(t.id, t.username, text, msg.Channel, msg.Time)

	if t.i != nil {
		t.i.Say(msg.Channel, text)
	} else {
		log.Print("Didn't actually send message - IRC client is nil. This is expected in test, but if you see this in production, something's broken!")
	}

	if err := t.cdb.StoreExpiringString(cache.KeyLastSentTwitchMessage, text, lastSentTwitchMessageExpiration); err != nil {
		log.Printf("[Twitch.persistUserAndMessage]: Failed to persist message in cache, %q: %v", text, err)
	}
	return nil
}

func (t *Twitch) Listen() <-chan base.IncomingMessage {
	c := make(chan base.IncomingMessage)
	t.i.OnPrivateMessage(func(msg twitchirc.PrivateMessage) {
		go t.persistUserAndMessage(msg.User.ID, msg.User.DisplayName, msg.Message, msg.Channel, msg.Time)
		c <- base.IncomingMessage{
			Message: base.Message{
				Text:    msg.Message,
				Channel: msg.Channel,
				UserID:  msg.User.ID,
				User:    msg.User.Name,
				Time:    msg.Time,
			},
			Prefix:          t.prefix(msg.Channel),
			PermissionLevel: t.level(&msg),
			Platform:        t,
		}
	})
	return c
}

var ErrBotIsBanned = errors.New("bot is banned from the channel")

func (t *Twitch) Join(channel string, prefix string) error {
	channelInfo, err := t.Channel(channel)
	if err != nil {
		return fmt.Errorf("failed to look up channel: %w", err)
	}

	startTime := time.Now()

	if t.i != nil {
		t.i.Join(channelInfo.BroadcasterName)
	} else {
		log.Printf("Didn't actually join channel %s - IRC client is nil. This is expected in test, but if you see this in production, something's broken!", channelInfo.BroadcasterName)
	}

	// Wait for ban messages to come in.
	// If we're banned, the IRC server will send a ban message about 150ms after we try to join.
	time.Sleep(time.Duration(250) * time.Millisecond)

	var banCount int64
	t.db.Model(&models.BotBan{}).Where("platform = ? AND channel = ? AND banned_at > ?", t.Name(), strings.ToLower(channel), startTime).Count(&banCount)
	if isBanned := banCount > 0; isBanned {
		return ErrBotIsBanned
	}

	t.channels = append(t.channels, &twitchChannel{Name: channelInfo.BroadcasterName, Prefix: prefix})
	return nil
}

func (t *Twitch) Leave(channel string) error {
	if t.i != nil {
		t.i.Depart(channel)
	} else {
		log.Printf("Didn't actually depart channel %s - IRC client is nil. This is expected in test, but if you see this in production, something's broken!", channel)
	}

	var newChannels []*twitchChannel
	for _, ch := range t.channels {
		if strings.EqualFold(ch.Name, channel) {
			continue
		}
		newChannels = append(newChannels, ch)
	}
	t.channels = newChannels

	return nil
}

// https://dev.twitch.tv/docs/irc/msg-id#notice-message-ids
const twitchMsgIdBanned = "msg_banned"

func (t *Twitch) Connect() error {
	log.Printf("Initializing channel data...")
	t.initializeJoinedChannels()

	log.Printf("Creating Twitch IRC client...")
	i := twitchirc.NewClient(t.username, fmt.Sprintf("oauth:%s", t.accessToken))
	t.i = i

	t.setUpIRCHandlers()

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
		log.Printf("[%s] Bot user %s is a verified bot, using increased rate limit", t.Name(), t.username)
		t.i.SetJoinRateLimiter(twitchirc.CreateVerifiedRateLimiter())
	}

	log.Printf("Connecting to Twitch IRC...")
	t.connectIRC()

	log.Printf("Connecting to Twitch API...")
	h, err := helix.NewClient(&helix.Options{
		ClientID:        t.clientID,
		UserAccessToken: t.accessToken,
	})
	if err != nil {
		return err
	}
	t.h = h

	botUser, err := t.FetchUser(t.username)
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

func (t *Twitch) connectIRC() {
	var wg sync.WaitGroup
	wg.Add(1)
	t.i.OnConnect(func() { wg.Done() })
	go func() {
		if err := t.i.Connect(); err != nil {
			log.Fatalf("failed to connect to twitch IRC: %v", err)
		}
	}()
	wg.Wait()

	for _, channel := range t.channels {
		log.Printf("Joining Twitch channel %s...", channel.Name)
		t.i.Join(channel.Name)
	}
}

func (t *Twitch) Disconnect() error {
	for _, channel := range t.channels {
		log.Printf("Leaving Twitch channel %s...", channel.Name)
		t.i.Depart(channel.Name)
	}
	log.Printf("Disconnecting from Twitch IRC...")
	return t.i.Disconnect()
}

func (t *Twitch) SetPrefix(channel, prefix string) error {
	for _, c := range t.channels {
		if strings.EqualFold(c.Name, channel) {
			c.Prefix = prefix
			return nil
		}
	}
	return fmt.Errorf("channel %s not joined", channel)
}

func (t *Twitch) User(username string) (models.User, error) {
	var user models.User
	result := t.db.Where("LOWER(twitch_name) = ?", strings.ToLower(username)).Limit(1).Find(&user)
	if result.Error != nil {
		return models.User{}, fmt.Errorf("failed to retrieve twitch user %s from db: %w", username, result.Error)
	}
	if user.ID == 0 {
		return models.User{}, fmt.Errorf("twitch user %s has never been seen by the bot: %w", username, base.ErrUserUnknown)
	}
	return user, nil
}

func (t *Twitch) CurrentUsers() ([]string, error) {
	var allChatters []string
	for _, c := range t.channels {
		chatters, err := twitchtmi.FetchChatters(c.Name)
		if err != nil {
			return nil, fmt.Errorf("[%s] failed to fetch chatters for %s: %w", t.Name(), c.Name, err)
		}
		for _, chatter := range chatters.AllChatters() {
			if slices.Contains(allChatters, chatter) {
				continue
			}
			allChatters = append(allChatters, chatter)
		}
	}
	return allChatters, nil
}

func (t *Twitch) Timeout(username, channel string, duration time.Duration) error {
	return t.Send(base.Message{
		Channel: channel,
		Text:    fmt.Sprintf("/timeout %s %.f", username, duration.Seconds()),
	})
}

func (t *Twitch) FetchUser(channel string) (*helix.User, error) {
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

var ErrChannelNotFound = errors.New("channel not found")

func (t *Twitch) Channel(channel string) (*helix.ChannelInformation, error) {
	user, err := t.FetchUser(channel)
	if err != nil {
		return nil, err
	}
	resp, err := t.h.GetChannelInformation(&helix.GetChannelInformationParams{BroadcasterIDs: []string{user.ID}})
	if err != nil {
		return nil, err
	}
	if len(resp.Data.Channels) == 0 {
		return nil, fmt.Errorf("channel %s not found: %w", channel, ErrChannelNotFound)
	}
	if len(resp.Data.Channels) > 1 {
		log.Printf("more than one channel found for %s, using the first", channel)
	}
	return &resp.Data.Channels[0], nil
}

// updateCachedJoinedChannels updates the in-memory joined channel data
// using the latest joined channel data from the database.
func (t *Twitch) updateCachedJoinedChannels() {
	var dbChannels []models.JoinedChannel
	t.db.Where(models.JoinedChannel{Platform: t.Name()}).Find(&dbChannels)

	var channels []*twitchChannel
	for _, dbChannel := range dbChannels {
		channels = append(channels, &twitchChannel{
			Name:   dbChannel.Channel,
			Prefix: dbChannel.Prefix,
		})
	}

	t.channels = channels
}

func (t *Twitch) prefix(channel string) string {
	for _, c := range t.channels {
		if !strings.EqualFold(c.Name, channel) {
			continue
		}
		return c.Prefix
	}
	log.Printf("No prefix found for channel %s", channel)
	return ""
}

// level returns the permission level of the user that sent a message.
func (t *Twitch) level(msg *twitchirc.PrivateMessage) permission.Level {
	if slices.Contains(t.owners, strings.ToLower(msg.User.Name)) {
		return permission.Owner
	}

	for _, channel := range t.channels {
		if !strings.EqualFold(channel.Name, msg.Channel) {
			continue
		}
		for badgeType, level := range badgeLevels {
			if userHasBadge(msg.User, badgeType) {
				return level
			}
		}
	}
	return permission.Normal
}

func (t *Twitch) initializeJoinedChannels() {
	var botChannel models.JoinedChannel
	botChannelResp := t.db.FirstOrCreate(&botChannel, models.JoinedChannel{
		Platform: t.Name(),
		Channel:  strings.ToLower(t.username),
	})

	if botChannelResp.RowsAffected != 0 {
		t.db.Model(&botChannel).Updates(models.JoinedChannel{
			Prefix:   defaultBotPrefix,
			JoinedAt: time.Now(),
		})
	}

	t.updateCachedJoinedChannels()
}

func (t *Twitch) setUpIRCHandlers() {
	t.i.OnClearMessage(func(msg twitchirc.ClearMessage) {
		log.Printf("[%s] CLEAR: %s", t.Name(), msg.Raw)
	})
	t.i.OnClearChatMessage(func(msg twitchirc.ClearChatMessage) {
		log.Printf("[%s] CLEARCHAT: %s", t.Name(), msg.Raw)
	})
	// OnConnect is set within Twitch.Connect()
	t.i.OnGlobalUserStateMessage(func(msg twitchirc.GlobalUserStateMessage) {})
	t.i.OnNoticeMessage(func(msg twitchirc.NoticeMessage) {
		log.Printf("[%s] NOTICE: %s", t.Name(), msg.Raw)

		// This fires when we the bot tries to join a channel it's banned in.
		if msg.MsgID == twitchMsgIdBanned {
			t.handleBannedFromChannel(msg.Channel)
		}
	})
	t.i.OnPingMessage(func(msg twitchirc.PingMessage) {})
	// OnPrivateMessage is set within Twitch.Connect()
	t.i.OnPongMessage(func(msg twitchirc.PongMessage) {})
	t.i.OnReconnectMessage(func(msg twitchirc.ReconnectMessage) {
		log.Printf("[%s] Reconnect requested, reconnecting...", t.Name())
		if err := t.Disconnect(); err != nil {
			log.Printf("[%s] Disconnect failed: %v", t.Name(), err)
		}
		t.connectIRC()
	})
	t.i.OnRoomStateMessage(func(msg twitchirc.RoomStateMessage) {})
	t.i.OnSelfJoinMessage(func(msg twitchirc.UserJoinMessage) {})
	t.i.OnSelfPartMessage(func(msg twitchirc.UserPartMessage) {
		log.Printf("[%s] SELFPART: %s", t.Name(), msg.Raw)
	})
	t.i.OnUnsetMessage(func(msg twitchirc.RawMessage) {
		log.Printf("[%s] UNSET: %s", t.Name(), msg.Raw)
	})
	t.i.OnUserJoinMessage(func(msg twitchirc.UserJoinMessage) {
		log.Printf("[%s] USERJOIN: %s", t.Name(), msg.Raw)
	})
	t.i.OnUserNoticeMessage(func(msg twitchirc.UserNoticeMessage) {
		log.Printf("[%s] USERNOTICE: %s", t.Name(), msg.Raw)
	})
	t.i.OnUserPartMessage(func(msg twitchirc.UserPartMessage) {
		log.Printf("[%s] USERPART: %s", t.Name(), msg.Raw)
	})
	t.i.OnUserStateMessage(func(msg twitchirc.UserStateMessage) {})
	t.i.OnWhisperMessage(func(msg twitchirc.WhisperMessage) {
		log.Printf("[%s] WHISPER: %s", t.Name(), msg.Raw)
		t.persistUserAndMessage(msg.User.ID, msg.User.Name, msg.Message, fmt.Sprintf("whisper-%s", t.Username()), time.Now())
	})
}

func (t *Twitch) persistUserAndMessage(twitchID, twitchName, message, channel string, sentTime time.Time) {
	var user models.User
	result := t.db.FirstOrCreate(&user, models.User{TwitchID: twitchID})
	if result.Error != nil {
		log.Printf("[Twitch.persistUserAndMessage]: Failed to find/create user, twitchName:%q %v", twitchName, result.Error)
	}
	result = t.db.Model(&user).Updates(models.User{TwitchName: twitchName})
	if result.Error != nil {
		log.Printf("[Twitch.persistUserAndMessage]: Failed to update user, twitchName:%q %v", twitchName, result.Error)
	}
	result = t.db.Create(&models.Message{
		Text:    message,
		Channel: channel,
		User:    user,
		Time:    sentTime,
	})
	if result.Error != nil {
		log.Printf("[Twitch.persistUserAndMessage]: Failed to persist message in database, %q/%q: %v", channel, message, result.Error)
	}
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
			channel.BotIsModerator = true
			return
		}
	}
	channel.BotIsModerator = false
}

func (t *Twitch) updateVIPStatusForChannel(channel *twitchChannel, vips []*ivr.ModOrVIPUser) {
	for _, vip := range vips {
		if vip.ID == t.id {
			channel.BotIsVIP = true
			return
		}
	}
	channel.BotIsVIP = false
}

func (t *Twitch) handleBannedFromChannel(channel string) {
	log.Printf("[%s] Banned from channel %s, leaving it", t.Name(), channel)
	go func() {
		t.db.Create(&models.BotBan{
			Platform: t.Name(),
			Channel:  strings.ToLower(channel),
			BannedAt: time.Now(),
		})
	}()
	go func() {
		if err := database.LeaveChannel(t.db, t.Name(), channel); err != nil {
			log.Printf("[%s] Failed to leave channel %s (database): %v", t.Name(), channel, err)
		}
	}()
	go func() {
		if err := t.Leave(channel); err != nil {
			log.Printf("[%s] Failed to leave channel %s (IRC): %v", t.Name(), channel, err)
		}
	}()
}

const defaultBotPrefix = "$"

// New creates a new Twitch connection.
func New(username string, owners []string, clientID, accessToken string, db *gorm.DB, cdb cache.Cache) *Twitch {
	return &Twitch{
		username:    username,
		owners:      lowercaseAll(owners),
		clientID:    clientID,
		accessToken: accessToken,
		db:          db,
		cdb:         cdb,
	}
}

// New creates a new Twitch connection for testing.
func NewForTesting(url string, db *gorm.DB) *Twitch {
	helixClient, err := helix.NewClient(&helix.Options{
		ClientID:        "fake-client-id",
		UserAccessToken: "fake-access-token",
		APIBaseURL:      url,
	})
	if err != nil {
		panic(err)
	}
	return &Twitch{
		username:      "fake-username",
		id:            "",
		isVerifiedBot: true,
		channels: []*twitchChannel{
			{Name: "user1"},
			{Name: "user2"},
		},
		owners:      nil,
		clientID:    "fake-client-id",
		accessToken: "fake-access-token",
		i:           nil,
		h:           helixClient,
		db:          db,
		cdb:         cachetest.NewInMemoryCache(),
	}
}

type twitchChannel struct {
	Name           string
	Prefix         string
	BotIsModerator bool
	BotIsVIP       bool
}

func lowercaseAll(strs []string) []string {
	lower := make([]string, len(strs))
	for i, str := range strs {
		lower[i] = strings.ToLower(str)
	}
	return lower
}

const messageSpaceSuffix = " \U000E0000"

// lastSentTwitchMessageExpiration is the duration the last sent message should remain in the cache.
// (Twitch blocks messages that are twice in a row in a 30-second period of time)
const lastSentTwitchMessageExpiration = time.Duration(30) * time.Second

// bypassSameMessageDetection manipulates the whitespace in a message
// in such a way to bypass Twitch's 30-second same message detection/blocking.
func bypassSameMessageDetection(message string) string {
	spaceIndex := strings.Index(message, " ")
	if ignoreFirstSpace := strings.HasPrefix(message, "/") || strings.HasPrefix(message, "."); ignoreFirstSpace {
		spaceIndex = strings.Index(message[spaceIndex+1:], " ")
	}

	if spaceIndex == -1 {
		return message + messageSpaceSuffix
	}

	return strings.Replace(message, " ", "  ", 1)
}
