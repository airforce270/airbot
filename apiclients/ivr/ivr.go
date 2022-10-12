// Package ivr provides an API client to the ivr.fi API.
package ivr

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var baseURL = "https://api.ivr.fi"

// twitchUsersResponse is the type that the IVR API responds with
// for calls to /v2/twitch/users.
type twitchUsersResponse []*twitchUsersResponseItem

type twitchUsersResponseItem struct {
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
	Roles rolesInfo `json:"roles"`
	// Badges contains the user's Twitch badges.
	// Note: does not include FFZ, 7TV, etc. badges.
	Badges []badgeInfo `json:"badges"`
	// ChatSettings contains information about this user's chat.
	ChatSettings chatSettingsInfo `json:"chatSettings"`
	// Stream contains info about the user's current stream.
	Stream *streamInfo `json:"stream"`
	// LastBroadcast contains info about this user's last stream.
	LastBroadcast lastBroadcastInfo `json:"lastBroadcast"`
	// Panels contains information about the panels on the user's about page.
	Panels []panelInfo `json:"panels"`
}

// rolesInfo contains information about a user's global Twitch roles.
type rolesInfo struct {
	// IsAffiliate is whether the user is a Twitch Affiliate.
	IsAffiliate bool `json:"isAffiliate"`
	// IsPartner is whether the user is a Twitch Partner.
	IsPartner bool `json:"isPartner"`
	// IsAffiliate is whether the user is a Twitch staff member.
	IsStaff bool `json:"isStaff"`
}

// badgeInfo represents a Twitch badge.
type badgeInfo struct {
	// Set is the set this badge is in.
	Set string `json:"setID"`
	// Title is the title of this badge.
	Title string `json:"title"`
	// Descripion is a verbose description of this badge.
	Description string `json:"description"`
	// Version is this badge's version.
	Version string `json:"version"`
}

// chatSettingsInfo contains information about a user's chat.
type chatSettingsInfo struct {
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

// streamInfo contains info about a user's current stream.
type streamInfo struct {
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
	Game gameInfo `json:"game"`
}

// gameInfo contains info about a game being played on stream.
type gameInfo struct {
	// DisplayName is the display name of the game.
	DisplayName string `json:"displayName"`
}

// lastBroadcastInfo contains info about a user's last stream.
type lastBroadcastInfo struct {
	// StartTime is when the user's last stream started.
	StartTime time.Time `json:"startedAt"`
	// Title is the title of the user's last stream.
	Title string `json:"title"`
}

// panelInfo contains information about a panel on a user's about page.
type panelInfo struct {
	// ID is the id of the panel.
	ID string `json:"id"`
}

// modsAndVIPsResponse is the type that the IVR API responds with
// for calls to /v2/twitch/modvip/{user}.
type modsAndVIPsResponse struct {
	// Mods is the moderators of the channel.
	Mods []*ModOrVIPUser `json:"mods"`
	// VIPs is the VIPs of the channel.
	VIPs []*ModOrVIPUser `json:"vips"`
}

// ModOrVipUser contains information about a mod/VIP user.
type ModOrVIPUser struct {
	// ID is the user's Twitch ID.
	ID string `json:"id"`
	// Username is the user's twitch username.
	Username string `json:"login"`
	// DisplayName is the user's display name.
	DisplayName string `json:"displayName"`
	// GrantedAt is when the user was made a mod/VIP.
	GrantedAt time.Time `json:"grantedAt"`
}

// foundersResponse is the type that the IVR API responds with
// for calls to /v2/twitch/founders/{user}.
type foundersResponse struct {
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

// FetchUser fetches a user's info from the IVR API.
func FetchUser(username string) (*twitchUsersResponseItem, error) {
	body, err := get(fmt.Sprintf("%s/v2/twitch/user?login=%s", baseURL, username))
	if err != nil {
		return nil, err
	}

	resp := twitchUsersResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from IVR API: %w", err)
	}

	return resp[0], nil
}

func FetchModsAndVIPs(channel string) (*modsAndVIPsResponse, error) {
	body, err := get(fmt.Sprintf("%s/v2/twitch/modvip/%s", baseURL, channel))
	if err != nil {
		return nil, err
	}

	resp := modsAndVIPsResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from IVR API: %w", err)
	}

	return &resp, nil
}

func FetchFounders(channel string) (*foundersResponse, error) {
	body, err := get(fmt.Sprintf("%s/v2/twitch/founders/%s", baseURL, channel))
	if err != nil {
		return nil, err
	}

	resp := foundersResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from IVR API: %w", err)
	}

	return &resp, nil
}

func get(reqURL string) (respBody []byte, err error) {
	httpResp, err := http.Get(reqURL)
	if err != nil {
		return nil, err
	}
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from IVR API (URL:%s): %v", reqURL, httpResp)
	}

	if httpResp.Body == nil {
		return nil, fmt.Errorf("no data returned from IVR API: %v", httpResp)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response from IVR API: %w", err)
	}

	return body, nil
}
