// Package kick provides an API client to the Kick API.
package kick

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Danny-Dasilva/CycleTLS/cycletls"
)

var (
	// Base URL for API requests. Should only be changed for testing.
	BaseURL = "https://kick.com"
	// ErrChannelNotFound is returned when a channel is not found.
	ErrChannelNotFound = errors.New("channel not found")
	// Token to use for Kick calls.
	Token string
	// UserToken to use for Kick calls.
	UserToken string

	cycleTLSClient = cycletls.Init()
	errNotFound    = errors.New("404 not found")
)

// Channel represents a Kick channel.
type Channel struct {
	// ID is the unique ID of this channel.
	ID int `json:"id"`
	// Name is the name of this channel.
	Name string `json:"slug"`
	// UserID is the ID of the user owning this channel.
	UserID int `json:"user_id"`
	// User is the user that owns this channel.
	User User `json:"user"`
	// Chatroom is this channel's chatroom.
	Chatroom Chatroom `json:"chatroom"`
	// IsBanned is whether the channel is banned.
	IsBanned bool `json:"is_banned"`
	// VODEnabled is whether the channel has VODs enabled.
	VODEnabled bool `json:"vod_enabled"`
	// SubscriptionEnabled is whether users can subscribe to this channel.
	SubscriptionEnabled bool `json:"subscription_enabled"`
	// FollowersCount is the number of followers this channel has.
	FollowersCount int `json:"followers_count"`
	// SubscriberBadges is this channel's sub badges.
	SubscriberBadges []SubscriberBadge `json:"subscriber_badges"`
	// BannerImage is this channel's banner image.
	BannerImage *ChannelBanner `json:"banner_image"`
	// OfflineBanerImage is this channel's banner image used when offline.
	OfflineBannerImage *ChannelBanner `json:"offline_banner_image"`
	// Livestream is this channel's current livestream, if the channel is live.
	Livestream *Livestream `json:"livestream"`
	// PlaybackURL is the URL to stream video from this channel
	PlaybackURL string `json:"playback_url"`
	// Verified is whether this channel is verified.
	Verified bool `json:"verified"`
	// RecentCategories is the categories this channel recently streamed in.
	RecentCategories []ExtendedCategory `json:"recent_categories"`

	// Following is of unknown use.
	// It does not appear to be reliably set.
	Following *bool `json:"following"`
	// CanHost is of unknown use.
	// (it's either whether this channel can host others or be hosted. not sure which one)
	CanHost bool `json:"can_host"`
	// Role is of unknown use.
	Role any `json:"role"`
	// Subscription is of unknown use.
	Subscription any `json:"subscription"`
	// Muted is of unknown use.
	Muted bool `json:"muted"`
	// FollowerBadges is of unknown use.
	FollowerBadges []any `json:"follower_badges"`
}

// SubscriberBadge represents a possible sub badge.
type SubscriberBadge struct {
	// ID is the unique ID of this sub badge.
	ID int `json:"id"`
	// ChannelID is the ID of the channel this sub badge is in.
	ChannelID int `json:"channel_id"`
	// Months is how many months a user would need to be subscribed
	// to have this badge.
	Months int `json:"months"`
	// Image is the image for this badge.
	Image BadgeImage `json:"badge_image"`
}

// BadgeImage represents a chat badge image.
type BadgeImage struct {
	// URL is the URL of this badge image.
	URL string `json:"src"`

	// SourceSet is of unknown use.
	// It appears to be unset.
	SourceSet string `json:"srcset"`
}

// ChannelBanner represents a channel banner image.
type ChannelBanner struct {
	// URL is the URL of the banner image.
	URL string `json:"url"`
}

// ExtendedCategory represents a stream category with additional information.
type ExtendedCategory struct {
	// ID is the unique ID of the extended category.
	ID int `json:"id"`
	// ID is the category ID of this extended category.
	CategoryID int `json:"category_id"`
	// DisplayName is the human-readable name of this extended category.
	// example: "Counter-Strike: Global Offensive"
	DisplayName string `json:"name"`
	// Slug is the unique string for this extended category (?).
	// example: "counter-strike-global-offensive"
	Slug string `json:"slug"`
	// Tags are the tags for this extended category.
	// example: ["FPS","Shooter","Action"]
	Tags []string `json:"tags"`
	// Description is a long-form description of the extended category
	// for display to users.
	// Example: "Call of Duty: Modern Warfare II drops players into an unprecedented global conflict that features the return of the iconic Operators of Task Force 141."
	Description string `json:"description"`
	// DeleteTime is when the extended category was deleted.
	DeletedAt *time.Time `json:"deleted_at"`
	// CurrentViewers is how many viewers the extended category currently has.
	CurrentViewers int `json:"viewers"`
	// Banner is this extended category's banner.
	Banner *CategoryBanner `json:"banner"`
	// Category is this extended category's category.
	Category Category `json:"category"`
}

// Banner represents a category banner image.
type CategoryBanner struct {
	// URL is the URL of the image for the banner.
	URL string `json:"url"`
	// ResponsiveURLs contains various URLs to load an image at a given resolution (?).
	// example: "https://files.kick.com/images/subcategories/2/banner/responsives/93c82fc6-77e2-4dea-a4c9-4b4525541317___banner_600_800.webp 600w, https://files.kick.com/images/subcategories/2/banner/responsives/93c82fc6-77e2-4dea-a4c9-4b4525541317___banner_501_668.webp 501w, ..."
	ResponsiveURLs string `json:"responsive"`
}

// Category represents a stream category.
type Category struct {
	// ID is the category's unique ID.
	ID int `json:"id"`
	// DisplayName is the human-readable name of the category.
	// example: "Games"
	DisplayName string `json:"name"`
	// Slug is the unique slug of the category (?).
	// example: "games"
	Slug string `json:"slug"`
	// Icon is an emoji for the category.
	// example: "🕹️"
	Icon string `json:"icon"`
}

// Livestream represents a single instance of a Kick livestream.
type Livestream struct {
	// ID is the unique ID of the stream.
	ID int `json:"id"`
	// ChannelID is the ID of the channel the stream is on.
	ChannelID int `json:"channel_id"`
	// Title is the title of the stream.
	Title string `json:"session_title"`
	// Slug is a unique string representing the stream (?).
	// example: "021bb-deo-cookoff-may-the-best-dish-win-deo4l"
	Slug string `json:"slug"`
	// CreateTime is when the stream was created.
	// It appears to (usually) be within 5 seconds of (after) StartTime.
	// example: "2023-09-01 14:41:56", appears to be UTC.
	CreateTime string `json:"created_at"`
	// StartTime is when the stream was started.
	// It appears to (usually) be within 5 seconds of (before) CreateTime.
	// i.e. "2023-09-01 14:41:56", appears to be UTC.
	StartTime string `json:"start_time"`
	// IsLive is whether the livestream is currently live.
	IsLive bool `json:"is_live"`
	// Language is the language the stream is in.
	// example: "English"
	Language string `json:"language"`
	// IsMature is whether the stream is intended for mature (18+) audiences.
	IsMature bool `json:"is_mature"`
	// ViewerCount is the stream's current viewer count.
	ViewerCount int `json:"viewer_count"`
	// Thumbnail is the most recent thumbnail of the stream.
	Thumbnail Thumbnail `json:"thumbnail"`
	// Categories is the categories of the stream.
	Categories []ExtendedCategory `json:"categories"`
	// Duration is how long the stream was, in seconds (?).
	// It appears to be 0 when the stream is live.
	Duration int `json:"duration"`

	// Tags is of unknown use.
	// It appears to be unset.
	Tags []any `json:"tags"`
	// RiskLevelID is of unknown use.
	// It appears to be unset.
	RiskLevelID any `json:"risk_level_id"`
	// Source is of unknown use.
	// It appears to be unset.
	Source any `json:"source"`
	// TwitchChannel is of unknown use.
	// It appears to be unset.
	TwitchChannel any `json:"twitch_channel"`
}

// Thumbnail represents a thumbnail image.
type Thumbnail struct {
	// URL is the URL of the image.
	URL string `json:"url"`
}

// User represents a Kick user.
type User struct {
	// ID is the unique ID of the user.
	ID int `json:"id"`
	// Username is the user's unique username.
	Username string `json:"username"`
	// Bio is the user's bio.
	Bio string `json:"bio"`
	// InstagramUsername is the user's username on Instagram.
	// example: "xqcow1/"
	InstagramUsername string `json:"instagram"`
	// TwitterProfileURL is the URL of the user's Twitter profile.
	// example: "https://twitter.com/xqc"
	TwitterProfileURL string `json:"twitter"`
	// YouTubeID is the user's YouTube channel ID.
	// examples: "raycondones" or "channel/UCmDTrq0LNgPodDOFZiSbsww"
	YouTubeID string `json:"youtube"`
	// DiscordUsername is the user's Discord username.
	// example: "xqc"
	DiscordUsername string `json:"discord"`
	// TikTokUsername is the user's TikTok username.
	// example: "xqcow1"
	TikTokUsername string `json:"tiktok"`
	// Facebook is the user's Facebook username (?).
	Facebook string `json:"facebook"`
	// ProfilePicURL is the URL of the user's profile picture.
	ProfilePicURL string `json:"profile_pic"`
	// AgreedToTerms is whether the user agreed to the Kick ToS (?).
	AgreedToTerms bool `json:"agreed_to_terms"`
	// EmailVerifiedAt is when the user verified their email.
	EmailVerifiedAt *time.Time `json:"email_verified_at"`

	// City is the user's city (?).
	// It appears to always be unset.
	City string `json:"city"`
	// State is the user's state (?).
	// It appears to always be unset.
	State string `json:"state"`
	// Country is the user's country (?).
	// It appears to always be unset.
	Country string `json:"country"`
}

// Chatroom represents a Kick chatroom.
type Chatroom struct {
	// ID is the unique ID of the chatroom.
	ID int `json:"id"`
	// ChannelID is the ID of the channel this chatroom is for.
	ChannelID int `json:"channel_id"`
	// CreateTime is when this chatroom was created.
	CreateTime time.Time `json:"created_at"`
	// UpdateTime is when this chatroom was last updated.
	UpdateTime time.Time `json:"updated_at"`
	// SlowMode is whether the chat is in slow mode, i.e. whether it is rate-limited.
	SlowMode bool `json:"slow_mode"`
	// SlowModeSeconds is the interval in which chatters can send messages.
	SlowModeSeconds int `json:"message_interval"`
	// FollowersOnlyMode is whether chat is in followers only mode,
	// i.e. whether only followers can chat.
	FollowersOnlyMode bool `json:"followers_mode"`
	// FollowersOnlyMinutes is how long the chatter must be followed for, in minutes, to chat.
	FollowersOnlyMinutes int `json:"following_min_duration"`
	// SubOnlyMode is whether the chat is in sub only mode,
	// i.e. whether only channel subscribers can chat.
	SubOnly bool `json:"subscribers_mode"`
	// EmoteOnlyMode is whether the chat is in emote only mode,
	// i.e. whether only emotes can be sent in chat (no other text).
	EmoteOnly bool `json:"emotes_mode"`

	// ChatMode is the mode this chatroom is in (?).
	// examples: "public", ...?
	// It is of unknown use.
	ChatMode string `json:"chat_mode"`
	// ChatModeOld is the mode this chatroom is in. (deprecated?)
	// examples: "public", ...?
	// It is of unknown use.
	ChatModeOld string `json:"chat_mode_old"`
	// ChatableType is of unknown use.
	// examples: "App\Models\Channel", ...?
	ChatableType string `json:"chatable_type"`
	// ChatableID is of unknown use.
	ChatableID int `json:"chatable_id"`
}

func FetchChannel(channel string) (*Channel, error) {
	body, err := get(fmt.Sprintf("%s/api/v2/channels/%s", BaseURL, strings.ToLower(channel)))
	if err != nil {
		if errors.Is(err, errNotFound) {
			return nil, fmt.Errorf("channel %s was not found: %w %w", channel, ErrChannelNotFound, err)
		}
		return nil, fmt.Errorf("failed to fetch chatters for %s: %w", channel, err)
	}

	resp := Channel{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from Kick API: %w", err)
	}
	return &resp, nil
}

func get(reqURL string) (respBody []byte, err error) {
	httpResp, err := cycleTLSClient.Do(reqURL, cycletls.Options{Ja3: Token, UserAgent: UserToken}, "GET")
	if err != nil {
		return nil, fmt.Errorf("get request to Kick API failed (URL:%s): %w", reqURL, err)
	}
	if httpResp.Status != http.StatusOK {
		if httpResp.Status == http.StatusNotFound {
			return nil, fmt.Errorf("bad response from Kick API (URL:%s): %v, %w", reqURL, httpResp, errNotFound)
		}
		return nil, fmt.Errorf("bad response from Kick API (URL:%s): %v", reqURL, httpResp)
	}

	return []byte(httpResp.Body), nil
}
