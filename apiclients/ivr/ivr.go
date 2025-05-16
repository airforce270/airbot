// Package ivr provides an API client to the ivr.fi API.
package ivr

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	// ErrUserNotFound is returned from some methods when the user couldn't be found.
	ErrUserNotFound = errors.New("user was not found")
	// ErrUserNotFound is returned from some methods when the channel couldn't be found.
	ErrChannelNotFound = errors.New("channel was not found")
)

// New DefaultClient returns a default IVR API client.
func NewDefaultClient() *Client { return NewClient("https://api.ivr.fi") }

// NewClient creates a new IVR API client.
func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

// Client is a client for the IVR API.
type Client struct {
	baseURL string
}

// FetchUsers fetches a user info from the IVR API.
func (c *Client) FetchUsers(username string) ([]*TwitchUsersResponseItem, error) {
	body, err := get(fmt.Sprintf("%s/v2/twitch/user?login=%s", c.baseURL, username))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users from IVR: %w", err)
	}

	resp := []*TwitchUsersResponseItem{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from IVR API: %w", err)
	}

	return resp, nil
}

// FetchModsAndVIPs fetches the mods and VIPs for a given a Twitch channel.
func (c *Client) FetchModsAndVIPs(channel string) (*ModsAndVIPsResponse, error) {
	body, err := get(fmt.Sprintf("%s/v2/twitch/modvip/%s", c.baseURL, channel))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch mods and vips from IVR: %w", err)
	}

	resp := ModsAndVIPsResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from IVR API: %w", err)
	}

	return &resp, nil
}

// FetchFounders fetches the list of founders for a given Twitch channel.
func (c *Client) FetchFounders(channel string) (*FoundersResponse, error) {
	reqURL := fmt.Sprintf("%s/v2/twitch/founders/%s", c.baseURL, channel)
	httpResp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch founders from IVR: %w", err)
	}
	// The IVR API responds with a 404 when the user has no founders.
	if httpResp.StatusCode == http.StatusNotFound {
		return &FoundersResponse{}, nil
	}
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from IVR API (URL:%s): %v", reqURL, httpResp)
	}

	if httpResp.Body == nil {
		return nil, fmt.Errorf("no data returned from IVR API: %v", httpResp)
	}
	defer func() { _ = httpResp.Body.Close() }() // ignore error

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response from IVR API: %w", err)
	}

	resp := FoundersResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from IVR API: %w", err)
	}

	return &resp, nil
}

// FetchSubAge fetches subscription information
// for a Twitch user to a given Twitch channel.
// If a user or channel was not found,
// ErrUserNotFound or ErrChannelNotFound will be returned, respectively.
func (c *Client) FetchSubAge(user, channel string) (*SubAgeResponse, error) {
	reqURL := fmt.Sprintf("%s/v2/twitch/subage/%s/%s", c.baseURL, user, channel)
	httpResp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sub age from IVR: %w", err)
	}
	// The IVR API responds with a 404 when a user or channel was not found.
	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusNotFound {
		return nil, fmt.Errorf("bad response from IVR API (URL:%s): %v", reqURL, httpResp)
	}

	if httpResp.Body == nil {
		return nil, fmt.Errorf("no data returned from IVR API: %v", httpResp)
	}
	defer func() { _ = httpResp.Body.Close() }() // ignore error

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response from IVR API: %w", err)
	}

	// The IVR API responds with a 404 when a user or channel was not found.
	if httpResp.StatusCode == http.StatusNotFound {
		if strings.Contains(string(body), "User was not found") {
			return nil, ErrUserNotFound
		}
		if strings.Contains(string(body), "Channel was not found") {
			return nil, ErrChannelNotFound
		}
		return nil, fmt.Errorf("bad response from IVR API (URL:%s): %v", reqURL, httpResp)
	}

	resp := SubAgeResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from IVR API: %w", err)
	}

	return &resp, nil
}

// TwitchUsersResponseItem /is the type that the IVR API responds with
// for calls to /v2/twitch/users.
type TwitchUsersResponseItem struct {
	// IsBanned is whether the user is banned on Twitch or not.
	IsBanned bool `json:"banned"`
	// BanReason is the reason the user was banned.
	BanReason string `json:"banReason"`
	// DisplayName is the user's name as it's displayed.
	DisplayName string `json:"displayName"`
	// Username is the user's login username. Usually is DisplayName lowercased.
	Username string `json:"login"`
	// ID is the user's Twitch id.
	ID string `json:"id"`
	// Bio is the user's bio.
	Bio string `json:"bio"`
	// FollowCount is how many channels the user follows.
	FollowCount int `json:"follows"`
	// FollowersCount is how many followers the user has.
	FollowersCount int `json:"followers"`
	// ProfileViewCount is how many times the user's profile has been viewed.
	ProfileViewCount int `json:"profileViewCount"`
	// ChatColor is the hex color of the user's name in chat.
	// Example: #FDFF00
	ChatColor string `json:"chatColor"`
	// ProfilePictureURL is the URL to the user's profile picture on Twitch.
	ProfilePictureURL string `json:"logo"`
	// BannerURL is the URL to the user's banner on Twitch.
	BannerURL string `json:"banner"`
	// IsVerifiedBot is whether the user is a verified bot on Twitch.
	// See https://dev.twitch.tv/docs/irc#verified-bots
	IsVerifiedBot bool `json:"verifiedBot"`
	// CreatedAt is when the user's account was created.
	CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt is when the user's account was last updated.
	UpdatedAt time.Time `json:"updatedAt"`
	// EmotePrefix is the user's emote prefix.
	EmotePrefix string `json:"emotePrefix"`
	// Roles contains information about the user's global Twitch roles.
	Roles RolesInfo `json:"roles"`
	// Badges contains the user's Twitch badges.
	// Note: does not include FFZ, 7TV, etc. badges.
	Badges []BadgeInfo `json:"badges"`
	// ChatSettings contains information about this user's chat.
	ChatSettings ChatSettingsInfo `json:"chatSettings"`
	// Stream contains info about the user's current stream.
	Stream *StreamInfo `json:"stream"`
	// LastBroadcast contains info about this user's last stream.
	LastBroadcast LastBroadcastInfo `json:"lastBroadcast"`
	// Panels contains information about the panels on the user's about page.
	Panels []PanelInfo `json:"panels"`
}

// RolesInfo contains information about a user's global Twitch roles.
type RolesInfo struct {
	// IsAffiliate is whether the user is a Twitch Affiliate.
	IsAffiliate bool `json:"isAffiliate"`
	// IsPartner is whether the user is a Twitch Partner.
	IsPartner bool `json:"isPartner"`
	// IsAffiliate is whether the user is a Twitch staff member.
	IsStaff bool `json:"isStaff"`
}

// BadgeInfo represents a Twitch badge.
type BadgeInfo struct {
	// Set is the set this badge is in.
	Set string `json:"setID"`
	// Title is the title of this badge.
	Title string `json:"title"`
	// Descripion is a verbose description of this badge.
	Description string `json:"description"`
	// Version is this badge's version.
	Version string `json:"version"`
}

// ChatSettingsInfo contains information about a user's chat.
type ChatSettingsInfo struct {
	// ChatDelayMs is the delay before messages are sent.
	ChatDelayMs int `json:"chatDelayMs"`
	// FollowersOnlyDurationMinutes is the minimum amount of time a user must be following for
	// before they can send chat messages.
	FollowersOnlyDurationMinutes int `json:"followersOnlyDurationMinutes"`
	// SlowModeDurationSeconds is the minimum amount of time between when users can send chat messages.
	SlowModeDurationSeconds int `json:"slowModeDurationSeconds"`
	// BlockLinks is whether links are blocked in chat messages.
	BlockLinks bool `json:"blockLinks"`
	// IsSubscribersOnlyModeEnabled is whether subs-only mode is enabled in chat.
	// If so, only subscribers to the channel can send messages.
	IsSubscribersOnlyModeEnabled bool `json:"isSubscribersOnlyModeEnabled"`
	// IsEmoteOnlyModeEnabled is whether emote-only mode is enabled in chat.
	// If so, messages sent can only contain Twitch emotes.
	IsEmoteOnlyModeEnabled bool `json:"isEmoteOnlyModeEnabled"`
	// IsFastSubsModeEnabled is whether fast-subs-only mode is enabled.
	IsFastSubsModeEnabled bool `json:"isFastSubsModeEnabled"`
	// IsUniqueChatModeEnabled is whether unique-chat mode is enabled.
	// If so, messages must be unique (within the last 30 seconds)
	IsUniqueChatModeEnabled bool `json:"isUniqueChatModeEnabled"`
	// RequireVerifiedAccount is whether a verified Twitch account is required to send chat messages.
	RequireVerifiedAccount bool `json:"requireVerifiedAccount"`
	// Rules contains the chat rules.
	Rules []string `json:"rules"`
}

// StreamInfo contains info about a user's current stream.
type StreamInfo struct {
	// Title is the title of the stream.
	Title string `json:"title"`
	// ID is the ID of the stream.
	ID string `json:"id"`
	// StartTime is when the stream started.
	StartTime time.Time `json:"createdAt"`
	// Type is the type of the stream.
	// When live, the value is "live".
	// Unsure what additional values this can contain.
	Type string `json:"type"`
	// ViewersCount is the number of current viewers to the stream.
	ViewersCount int `json:"viewersCount"`
	// Game contains info about the game being played on stream.
	Game GameInfo `json:"game"`
}

// GameInfo contains info about a game being played on stream.
type GameInfo struct {
	// DisplayName is the display name of the game.
	DisplayName string `json:"displayName"`
}

// LastBroadcastInfo contains info about a user's last stream.
type LastBroadcastInfo struct {
	// StartTime is when the user's last stream started.
	StartTime time.Time `json:"startedAt"`
	// Title is the title of the user's last stream.
	Title string `json:"title"`
}

// PanelInfo contains information about a panel on a user's about page.
type PanelInfo struct {
	// ID is the id of the panel.
	ID string `json:"id"`
}

// ModsAndVIPsResponse is the type that the IVR API responds with
// for calls to /v2/twitch/modvip/{user}.
type ModsAndVIPsResponse struct {
	// Mods is the moderators of the channel.
	Mods []*ModOrVIPUser `json:"mods"`
	// VIPs is the VIPs of the channel.
	VIPs []*ModOrVIPUser `json:"vips"`
}

// ModOrVipUser contains information about a mod/VIP user.
type ModOrVIPUser struct {
	// ID is the user's Twitch ID.
	ID string `json:"id"`
	// Username is the user's Twitch username.
	Username string `json:"login"`
	// DisplayName is the user's display name.
	DisplayName string `json:"displayName"`
	// GrantedAt is when the user was made a mod/VIP.
	GrantedAt time.Time `json:"grantedAt"`
}

// FoundersResponse is the type that the IVR API responds with
// for calls to /v2/twitch/founders/{user}.
type FoundersResponse struct {
	Founders []*Founder `json:"founders"`
}

// Founder contains information about a channel's founder.
type Founder struct {
	// ID is the user's Twitch ID.
	ID string `json:"id"`
	// Username is the user's twitch username.
	Username string `json:"login"`
	// DisplayName is the user's display name.
	DisplayName string `json:"displayName"`
	// InitiallySubbedAt is when the user initially subbed to the channel.
	InitiallySubbedAt time.Time `json:"entitlementStart"`
	// IsSubscribed is whether the user is currently subscribed.
	IsSubscribed bool `json:"isSubscribed"`
}

// SubAgeResponse is the type that the IVR API responds with
// for calls to /v2/twitch/subage/{user}/{channel}.
type SubAgeResponse struct {
	// User is the user in question.
	User SubAgeUser `json:"user"`
	// Channel is the channel in question.
	Channel SubAgeUser `json:"channel"`
	// StatusHidden is whether the user has hidden the status
	// of their subscription to this channel.
	StatusHidden bool `json:"statusHidden"`
	// FollowTime is when the user followed the channel.
	FollowTime time.Time `json:"followedAt"`
	// Streak contains information about the user's
	// subscription streak to this channel.
	Streak *SubAgeDuration `json:"streak"`
	// Streak contains information about the user's
	// cumulative subscription time to this channel.
	Cumulative *SubAgeDuration `json:"cumulative"`
	// Metadata contains metadata about the user's subscription to this channel.
	Metadata *SubAgeMetadata `json:"meta"`
}

// SubAgeUser represents a user/channel used in subscription age lookups.
type SubAgeUser struct {
	// ID is the user's Twitch ID.
	ID string `json:"id"`
	// Username is the user's Twitch username.
	Username string `json:"login"`
	// DisplayName is the user's name as it's displayed.
	DisplayName string `json:"displayName"`
}

// SubAgeDuration contains information
// about the duration of a user's subscription.
type SubAgeDuration struct {
	// ElapsedDays is the number of days elapsed in the *current month*
	// (or latest month, if an expired subscription) of a subscription.
	// Note that this is NOT the TOTAL elapsed days of a subscription.
	ElapsedDays int `json:"elapsedDays"`
	// DaysRemaining is how many days are remaining in the subscription.
	DaysRemaining int `json:"daysRemaining"`
	// Months is the number of cumulative months in the subscription.
	// If this is part of a streak, the number of months in the streak.
	Months int `json:"months"`
	// StartTime is when the current month of the subscription started.
	StartTime time.Time `json:"start"`
	// EndTime is when the current month of the subscription will end.
	EndTime time.Time `json:"end"`
}

// SubAgeMetadata contains metadata about a user's subscription to a channel.
type SubAgeMetadata struct {
	// Type is the subscription type.
	// Either "paid", "prime", or "gift"
	Type string `json:"type"`
	// Tier is the tier the user is subscribed at.
	// Either "1", "2", or "3".
	Tier string `json:"tier"`
	// EndTime is when the user's subscription ends.
	EndTime time.Time `json:"endsAt"`
	// RenewTime is when the user's subscription is set to renew.
	// Only set if the user has auto-renew enabled for the subscription.
	RenewTime *time.Time `json:"renewsAt"`
	// GiftInfo contains information about the gift, if the subscription is gifted.
	// Only set if Type is "gift".
	GiftInfo *SubAgeGiftMetadata `json:"giftMeta"`
}

// GiftInfo contains metadata about a gifted subscription.
type SubAgeGiftMetadata struct {
	// GiftTime is when the subscription was gifted.
	GiftTime time.Time `json:"giftDate"`
	// Gifter is the user that gifted the subscription.
	// Only present if the gifting user did not perform an anonymous gift.
	Gifter *SubAgeUser `json:"gifter"`
}

func get(reqURL string) (respBody []byte, err error) {
	httpResp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("get request to IVR failed (URL:%s): %w", reqURL, err)
	}
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from IVR API (URL:%s): %v", reqURL, httpResp)
	}
	defer func() { _ = httpResp.Body.Close() }() // ignore error

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response from IVR API: %w", err)
	}

	return body, nil
}
