// Package twitch handles Twitch-specific logic.
package twitch

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/airforce270/airbot/apiclients/ivr"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/cache"
	"github.com/airforce270/airbot/cache/cachetest"
	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/database/models"
	"github.com/airforce270/airbot/permission"
	"github.com/airforce270/airbot/utils"

	twitchirc "github.com/gempir/go-twitch-irc/v4"
	"github.com/nicklaw5/helix/v2"
	"gorm.io/gorm"
)

// Name is the unique, human-readable name of the platform.
const Name = "Twitch"

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
	// clientSecret is the OAuth client secret to use when connecting.
	clientSecret string
	// accessToken is the OAuth token to use when connecting.
	// See https://dev.twitch.tv/docs/irc/authenticate-bot#getting-an-access-token
	accessToken string
	// refreshToken is the refresh token to use to refresh the access token.
	refreshToken string
	// irc is the Twitch IRC client.
	irc *twitchirc.Client
	// helix is the Twitch API client.
	helix *helix.Client
	// db is a reference to the database connection.
	db *gorm.DB
	// cdb is a reference to the cache.
	cdb cache.Cache
}

func (t *Twitch) Name() string { return Name }

func (t *Twitch) Username() string { return t.username }

func (t *Twitch) Send(msg base.Message) error {
	return t.Reply(msg, "")
}

func (t *Twitch) Reply(msg base.Message, replyToID string) error {
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

	if t.irc != nil {
		if replyToID != "" {
			t.irc.Reply(msg.Channel, replyToID, text)
		} else {
			t.irc.Say(msg.Channel, text)
		}
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
	t.irc.OnPrivateMessage(func(msg twitchirc.PrivateMessage) {
		go t.persistUserAndMessage(msg.User.ID, msg.User.DisplayName, msg.Message, msg.Channel, msg.Time)
		c <- base.IncomingMessage{
			Message: base.Message{
				Text:    msg.Message,
				Channel: msg.Channel,
				ID:      msg.ID,
				UserID:  msg.User.ID,
				User:    msg.User.Name,
				Time:    msg.Time,
			},
			Prefix:          t.prefix(msg.Channel),
			PermissionLevel: t.level(&msg),
			Resources: base.Resources{
				Platform: t,
			},
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

	if t.irc != nil {
		t.irc.Join(channelInfo.BroadcasterName)
	} else {
		log.Printf("Didn't actually join channel %s - IRC client is nil. This is expected in test, but if you see this in production, something's broken!", channelInfo.BroadcasterName)
	}

	// Wait for ban messages to come in.
	// If we're banned, the IRC server will send a ban message about 150ms after we try to join.
	time.Sleep(250 * time.Millisecond)

	var banCount int64
	err = t.db.Model(&models.BotBan{}).Where("platform = ? AND channel = ? AND banned_at > ?", t.Name(), strings.ToLower(channel), startTime).Count(&banCount).Error
	if err != nil {
		return fmt.Errorf("failed to retrieve %s channels bot is banned in: %w", t.Name(), err)
	}
	if isBanned := banCount > 0; isBanned {
		return fmt.Errorf("bot is banned from %s/%s: %w", t.Name(), channel, ErrBotIsBanned)
	}

	t.channels = append(t.channels, &twitchChannel{
		ID:     channelInfo.BroadcasterID,
		Name:   channelInfo.BroadcasterName,
		Prefix: prefix,
	})
	return nil
}

func (t *Twitch) Leave(channel string) error {
	if t.irc != nil {
		t.irc.Depart(channel)
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

func (t *Twitch) Connect(ctx context.Context) error {
	log.Printf("[%s] Initializing channel data...", t.Name())
	if err := t.ensureSelfIsJoined(); err != nil {
		return fmt.Errorf("[%s] failed to join self: %w", t.Name(), err)
	}
	if err := t.populateInMemoryJoinedChannelCache(); err != nil {
		return fmt.Errorf("[%s] failed to populate in-memory joined channel cache: %w", t.Name(), err)
	}

	log.Printf("[%s] Creating IRC client...", t.Name())
	i := twitchirc.NewClient(t.username, "oauth:"+t.accessToken)
	t.irc = i

	log.Printf("[%s] Connecting to Twitch API...", t.Name())
	h, err := helix.NewClient(&helix.Options{
		ClientID:        t.clientID,
		ClientSecret:    t.clientSecret,
		UserAccessToken: t.accessToken,
		RefreshToken:    t.refreshToken,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to Twitch API: %w", err)
	}
	t.helix = h

	t.helix.OnUserAccessTokenRefreshed(func(userAccessToken, refreshToken string) {
		t.accessToken = userAccessToken
		t.refreshToken = refreshToken
		t.irc.SetIRCToken("oauth:" + userAccessToken)
	})

	// Make sure to do this before connecting to IRC.
	// This makes an API call to Twitch which automatically refreshes the token if needed.
	// That way we have a valid token before connecting to IRC.
	log.Printf("[%s] Refetching bot's ID...", t.Name())
	botUser, err := t.FetchUser(t.username)
	if err != nil {
		return fmt.Errorf("failed to fetch user %s: %w", t.username, err)
	}
	if botUser == nil {
		return fmt.Errorf("no user returned for bot (%s): %w", t.username, err)
	}
	t.id = botUser.ID

	log.Printf("[%s] Updating cached joined channels...", t.Name())
	if _, err := t.updateCachedJoinedChannels(); err != nil {
		return fmt.Errorf("failed to update cached joined channels: %w", err)
	}
	go t.startWatchingForChannelRenames(ctx)

	t.setUpIRCHandlers()

	log.Printf("[%s] Checking if the bot is a verified bot...", t.Name())
	ivrClient := ivr.NewDefaultClient()
	ivrUsers, err := ivrClient.FetchUsers(t.username)
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
		t.irc.SetJoinRateLimiter(twitchirc.CreateVerifiedRateLimiter())
	}

	log.Printf("[%s] Connecting to Twitch IRC...", t.Name())
	t.connectIRC()

	go t.listenForModAndVIPChanges(ctx, ivrClient)

	return nil
}

func (t *Twitch) connectIRC() {
	var wg sync.WaitGroup
	wg.Add(1)
	t.irc.OnConnect(func() { wg.Done() })
	go func() {
		if err := t.irc.Connect(); err != nil && !errors.Is(err, twitchirc.ErrClientDisconnected) {
			log.Fatalf("failed to connect to twitch IRC: %v", err)
		}
	}()
	wg.Wait()
	t.irc.OnConnect(nil)

	for _, channel := range t.channels {
		log.Printf("[%s] Joining channel %s...", t.Name(), channel.Name)
		t.irc.Join(channel.Name)
	}
}

func (t *Twitch) Disconnect() error {
	for _, channel := range t.channels {
		log.Printf("[%s] Leaving channel %s...", t.Name(), channel.Name)
		t.irc.Depart(channel.Name)
	}
	log.Printf("[%s] Disconnecting from Twitch IRC...", t.Name())
	return t.irc.Disconnect()
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
	err := t.db.Where("LOWER(twitch_name) = ?", strings.ToLower(username)).Limit(1).Find(&user).Error
	if err != nil {
		return models.User{}, fmt.Errorf("failed to retrieve twitch user %s from db: %w", username, err)
	}
	if user.ID == 0 {
		return models.User{}, fmt.Errorf("twitch user %s has never been seen by the bot: %w", username, base.ErrUserUnknown)
	}
	return user, nil
}

func (t *Twitch) CurrentUsers() ([]string, error) {
	var allChatters []string
	for _, c := range t.channels {
		pageToken := "<unset>" + strconv.Itoa(rand.Int())
		for pageToken != "" {
			req := &helix.GetChatChattersParams{
				BroadcasterID: c.Name,
				ModeratorID:   t.username,
				First:         "1000", // maximum is 1000
				After:         pageToken,
			}
			resp, err := t.helix.GetChannelChatChatters(req)
			if err != nil {
				return nil, fmt.Errorf("[%s] failed to fetch chatters for %s: %w", t.Name(), c.Name, err)
			}
			pageToken = resp.Data.Pagination.Cursor

			for _, chatter := range resp.Data.Chatters {
				if slices.Contains(allChatters, chatter.Username) {
					continue
				}
				allChatters = append(allChatters, chatter.Username)
			}
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
	users, err := t.helix.GetUsers(&helix.UsersParams{Logins: []string{channel}})
	if err != nil {
		return nil, fmt.Errorf("failed to get user %s from Helix: %w", channel, err)
	}
	if users.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("twitch GetUsers call for %q failed, resp:%v", channel, users)
	}
	if len(users.Data.Users) != 1 {
		return nil, fmt.Errorf("wrong number of users returned (should be 1): %w", err)
	}
	return &users.Data.Users[0], nil
}

var ErrChannelNotFound = errors.New("channel not found")

func (t *Twitch) Channel(channel string) (*helix.ChannelInformation, error) {
	user, err := t.FetchUser(channel)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user %s: %w", channel, err)
	}
	resp, err := t.helix.GetChannelInformation(&helix.GetChannelInformationParams{
		BroadcasterIDs: []string{user.ID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get info for channel %s from Helix: %w", channel, err)
	}
	if len(resp.Data.Channels) == 0 {
		return nil, fmt.Errorf("channel %s not found: %w", channel, ErrChannelNotFound)
	}
	if len(resp.Data.Channels) > 1 {
		log.Printf("more than one channel found for %s, using the first", channel)
	}
	return &resp.Data.Channels[0], nil
}

func (t *Twitch) startWatchingForChannelRenames(ctx context.Context) {
	const checkInterval = 2 * time.Minute
	ticker := time.NewTicker(checkInterval)
	for {
		select {
		case <-ticker.C:
			renamed, err := t.updateCachedJoinedChannels()
			if err != nil {
				log.Printf("[%s] Failed to update cached joined channels: %v", t.Name(), err)
			}
			for _, channel := range renamed {
				if err := t.Join(channel.Channel, channel.Prefix); err != nil {
					log.Printf("[%s] Failed to rejoin renamed channel %s (%s)", t.Name(), channel.Channel, channel.ChannelID)
				}
			}
		case <-ctx.Done():
			log.Printf("[%s] Stopping watching for channel renames, context cancelled", t.Name())
			ticker.Stop()
			return
		}
	}
}

// updateCachedJoinedChannels updates the joined channel caches.
//
// It:
//  1. Checks the joined channel records for name updates
//  2. Updates the DB records if any were found
//  3. Updates the now-obsolete entries in the in-memory cache (t.channels)
//  4. Returns the channels that were renamed
//
// Importantly, renamed channels are not joined.
// The caller is responsible for rejoining those channels.
func (t *Twitch) updateCachedJoinedChannels() ([]*models.JoinedChannel, error) {
	var renamed []*models.JoinedChannel

	var joinedChannels []*models.JoinedChannel
	if err := t.db.Find(&joinedChannels).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve joined channels (???): %w", err)
	}

	helixUsersByID := map[string]helix.User{}
	for _, channelBatch := range utils.Chunk(joinedChannels, 100) {
		ids := make([]string, 0, 100)
		logins := make([]string, 0, 100)
		for _, joinedChannel := range channelBatch {
			if joinedChannel.ChannelID != "" {
				ids = append(ids, joinedChannel.ChannelID)
			} else {
				logins = append(ids, joinedChannel.Channel)
			}
		}
		resp, err := t.helix.GetUsers(&helix.UsersParams{IDs: ids, Logins: logins})
		if err != nil || resp.Error != "" {
			return nil, fmt.Errorf("failed to get users from Helix: %d %s %s %w", resp.ErrorStatus, resp.Error, resp.ErrorMessage, err)
		}
		for _, user := range resp.Data.Users {
			helixUsersByID[user.ID] = user
		}
	}

	for _, joinedChannel := range joinedChannels {
		if joinedChannel.ChannelID == "" {
			log.Printf("[%s] Detected joined channel with no ID (%s) - probably the bot. Attempting to fix.\n", t.Name(), joinedChannel.Channel)
			found := false
			for id, user := range helixUsersByID {
				if user.Login == joinedChannel.Channel {
					log.Printf("[%s] Found ID for %s: %s", t.Name(), joinedChannel.Channel, id)
					joinedChannel.ChannelID = id
					renamed = append(renamed, joinedChannel)
					found = true
					break
				}
			}
			if found {
				continue
			}
			log.Printf("[%s] Couldn't find ID for channel %s", t.Name(), joinedChannel.Channel)
		}

		helixUser, ok := helixUsersByID[joinedChannel.ChannelID]
		if !ok {
			log.Printf("[%s] User %s (%s) not found in Helix lookup", t.Name(), joinedChannel.ChannelID, joinedChannel.Channel)
			continue
		}
		if joinedChannel.Channel != helixUser.Login {
			log.Printf("[%s] Detected renamed channel (%s): %s->%s", t.Name(), helixUser.ID, joinedChannel.Channel, helixUser.Login)
			joinedChannel.Channel = helixUser.Login
			renamed = append(renamed, joinedChannel)
		}
	}

	for _, renamedChannel := range renamed {
		if err := t.db.Save(renamedChannel).Error; err != nil {
			log.Printf("[%s] Failed to persist rename of %s to %s: %v", t.Name(), renamedChannel.ChannelID, renamedChannel.Channel, err)
		}
	}

	for _, channel := range t.channels {
		for _, renamedChannel := range renamed {
			if renamedChannel.ChannelID == channel.ID {
				channel.Name = renamedChannel.Channel
				break
			}
		}
	}

	return renamed, nil
}

// populateInMemoryJoinedChannelCache populates the in-memory joined channel
// data using the latest joined channel data from the database.
func (t *Twitch) populateInMemoryJoinedChannelCache() error {
	var dbChannels []models.JoinedChannel
	if err := t.db.Where(models.JoinedChannel{Platform: t.Name()}).Find(&dbChannels).Error; err != nil {
		return fmt.Errorf("failed to fetch channels for %s from DB: %w", t.Name(), err)
	}

	var channels []*twitchChannel
	for _, dbChannel := range dbChannels {
		channels = append(channels, &twitchChannel{
			ID:     dbChannel.ChannelID,
			Name:   dbChannel.Channel,
			Prefix: dbChannel.Prefix,
		})
	}

	t.channels = channels

	return nil
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

func (t *Twitch) ensureSelfIsJoined() error {
	var botChannel models.JoinedChannel
	result := t.db.Where(models.JoinedChannel{Platform: t.Name(), Channel: strings.ToLower(t.username)}).
		Attrs(models.JoinedChannel{Prefix: defaultBotPrefix, JoinedAt: time.Now()}).
		FirstOrCreate(&botChannel)
	if err := result.Error; err != nil {
		return fmt.Errorf("failed to fetch/create DB row for %s/%s: %w", t.Name(), strings.ToLower(t.username), err)
	}
	return nil
}

func (t *Twitch) setUpIRCHandlers() {
	t.irc.OnClearMessage(func(msg twitchirc.ClearMessage) {
		log.Printf("[%s] CLEAR: %s", t.Name(), msg.Raw)
	})
	t.irc.OnClearChatMessage(func(msg twitchirc.ClearChatMessage) {
		log.Printf("[%s] CLEARCHAT: %s", t.Name(), msg.Raw)
	})
	// OnConnect is set within Twitch.Connect()
	t.irc.OnGlobalUserStateMessage(func(msg twitchirc.GlobalUserStateMessage) {})
	t.irc.OnNoticeMessage(func(msg twitchirc.NoticeMessage) {
		log.Printf("[%s] NOTICE: %s", t.Name(), msg.Raw)

		// This fires when we the bot tries to join a channel it's banned in.
		if msg.MsgID == twitchMsgIdBanned {
			t.handleBannedFromChannel(msg.Channel)
		}
	})
	t.irc.OnPingMessage(func(msg twitchirc.PingMessage) {})
	// OnPrivateMessage is set within Twitch.Connect()
	t.irc.OnPongMessage(func(msg twitchirc.PongMessage) {})
	t.irc.OnReconnectMessage(func(msg twitchirc.ReconnectMessage) {
		log.Printf("[%s] Reconnect requested, reconnecting...", t.Name())
		if err := t.Disconnect(); err != nil {
			log.Printf("[%s] Disconnect failed: %v", t.Name(), err)
		}
		t.connectIRC()
	})
	t.irc.OnRoomStateMessage(func(msg twitchirc.RoomStateMessage) {})
	t.irc.OnSelfJoinMessage(func(msg twitchirc.UserJoinMessage) {})
	t.irc.OnSelfPartMessage(func(msg twitchirc.UserPartMessage) {
		log.Printf("[%s] SELFPART: %s", t.Name(), msg.Raw)
	})
	t.irc.OnUnsetMessage(func(msg twitchirc.RawMessage) {
		log.Printf("[%s] UNSET: %s", t.Name(), msg.Raw)
	})
	t.irc.OnUserJoinMessage(func(msg twitchirc.UserJoinMessage) {
		log.Printf("[%s] USERJOIN: %s", t.Name(), msg.Raw)
	})
	t.irc.OnUserNoticeMessage(func(msg twitchirc.UserNoticeMessage) {
		log.Printf("[%s] USERNOTICE: %s", t.Name(), msg.Raw)
	})
	t.irc.OnUserPartMessage(func(msg twitchirc.UserPartMessage) {
		log.Printf("[%s] USERPART: %s", t.Name(), msg.Raw)
	})
	t.irc.OnUserStateMessage(func(msg twitchirc.UserStateMessage) {})
	t.irc.OnWhisperMessage(func(msg twitchirc.WhisperMessage) {
		log.Printf("[%s] WHISPER: %s", t.Name(), msg.Raw)
		t.persistUserAndMessage(msg.User.ID, msg.User.Name, msg.Message, "whisper-"+t.Username(), time.Now())
	})
}

func (t *Twitch) persistUserAndMessage(twitchID, twitchName, message, channel string, sentTime time.Time) {
	var user models.User
	result := t.db.Where(models.User{TwitchID: twitchID}).Assign(models.User{TwitchName: twitchName}).FirstOrCreate(&user)
	if err := result.Error; err != nil {
		log.Printf("[Twitch.persistUserAndMessage]: Failed to find/create user, twitchName:%q %v", twitchName, err)
	}
	result = t.db.Create(&models.Message{
		Text:    message,
		Channel: channel,
		User:    user,
		Time:    sentTime,
	})
	if err := result.Error; err != nil {
		log.Printf("[Twitch.persistUserAndMessage]: Failed to persist message in database, %q/%q: %v", channel, message, err)
	}
}

func (t *Twitch) listenForModAndVIPChanges(ctx context.Context, ivrClient *ivr.Client) {
	const checkInterval = 30 * time.Second
	ticker := time.NewTicker(checkInterval)
	for {
		select {
		case <-ticker.C:
			for _, channel := range t.channels {
				modsAndVIPs, err := ivrClient.FetchModsAndVIPs(channel.Name)
				if err != nil {
					log.Printf("Failed to look up mods and VIPs for %s: %v", channel.Name, err)
					break
				}
				go t.updateModStatusForChannel(channel, modsAndVIPs.Mods)
				go t.updateVIPStatusForChannel(channel, modsAndVIPs.VIPs)
			}
		case <-ctx.Done():
			log.Printf("[%s] Context cancelled, stopping listening for mod and VIP changes", t.Name())
			ticker.Stop()
			return
		}
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
		err := t.db.Create(&models.BotBan{
			Platform: t.Name(),
			Channel:  strings.ToLower(channel),
			BannedAt: time.Now(),
		}).Error
		if err != nil {
			log.Printf("failed to create bot ban DB entry: %v", err)
		}
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
func New(username string, owners []string, clientID, clientSecret, accessToken, refreshToken string, db *gorm.DB, cdb cache.Cache) *Twitch {
	return &Twitch{
		username:     username,
		owners:       lowercaseAll(owners),
		clientID:     clientID,
		clientSecret: clientSecret,
		accessToken:  accessToken,
		refreshToken: refreshToken,
		db:           db,
		cdb:          cdb,
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
		irc:         nil,
		helix:       helixClient,
		db:          db,
		cdb:         cachetest.NewInMemory(),
	}
}

type twitchChannel struct {
	ID             string
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
const lastSentTwitchMessageExpiration = 30 * time.Second

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
