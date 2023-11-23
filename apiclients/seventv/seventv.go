// Package seventv provides a client to the 7tv API.
// https://7tv.io/docs
package seventv

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// DefaultBaseURL is the default base URL for the 7TV API.
const DefaultBaseURL = "https://7tv.io/"

// NewDefaultClient returns a new default 7TV API client.
func NewDefaultClient() *Client { return NewClient(DefaultBaseURL) }

// NewClient creates a new 7TV API client.
func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

// Client is a 7TV client.
type Client struct {
	baseURL string
}

// FetchUserConnectionByTwitchUserId fetches a 7tv user+connection
// given a Twitch userid.
func (c *Client) FetchUserConnectionByTwitchUserId(uid string) (*PlatformConnection, error) {
	rawResp, err := http.Get(fmt.Sprintf("%s/v3/users/twitch/%s", c.baseURL, uid))
	if err != nil {
		return nil, fmt.Errorf("failed to get 7tv user connection for twitch user %q: %w", uid, err)
	}
	if rawResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from getting 7tv user connection for twitch user %q: %v", uid, rawResp)
	}
	defer rawResp.Body.Close()

	respBody, err := io.ReadAll(rawResp.Body)
	if err != nil {
		return nil, fmt.Errorf("unreadable response from getting 7tv user connection for twitch user %q: %w", uid, err)
	}

	var resp PlatformConnection
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from getting 7tv user connection for twitch user %q: %w", uid, err)
	}

	return &resp, nil
}

// PlatformConnection is a connection between a 7TV account
// and an external platform account.
type PlatformConnection struct {
	// Platform is the platform connected to.
	// Examples: "TWITCH", "DISCORD", "KICK"
	Platform string `json:"platform"`
	// ID is the user's ID on the platform.
	// Example: "181950834"
	ID string `json:"id"`
	// Username is the user's username on the platform.
	// Example: "airforce2700"
	Username string `json:"username"`
	// DisplayName is the user's display name on the platform.
	DisplayName string `json:"display_name"`
	// LinkedAt is when the connection was created.
	LinkedAt UnixTimeMs `json:"linked_at"`
	// EmoteCapacity is the number of emotes that can be
	// used at one time on the connection.
	EmoteCapacity int `json:"emote_capacity"`
	// EmoteSetID is the unique ID of the active emote set on this connection.
	// It appears to typically be unset.
	EmoteSetID *string `json:"emote_set_id"`
	// EmoteSet is the active emote set on this connection.
	EmoteSet EmoteSet `json:"emote_set"`
	// User is the 7TV user for this connection.
	User User `json:"user"`
}

// User is a 7TV user.
type User struct {
	// ID is the user's ID.
	// Example: "621f13b614f489808df5d58e"
	ID string `json:"id"`
	// Username is the user's username.
	// Example: "airforce2700"
	Username string `json:"username"`
	// DisplayName is the user's display name.
	DisplayName string `json:"display_name"`
	// CreateTime is when the user created was created.
	CreateTime UnixTimeMs `json:"created_at"`
	// AvatarURL is the URL of the user's avatar.
	// Notably, does not include the scheme and starts with `//`.
	// Example: "//cdn.7tv.app/user/<id>/<img>/3x.webp"
	// If the user is not a 7TV subscriber or does not have an avatar,
	// this value will be "//cdn.7tv.app/"
	AvatarURL string `json:"avatar_url"`
	// Bio is this user's bio.
	Bio string `json:"biography"`
	// Style is this user's style information.
	Style Style `json:"style"`
	// RoleIDs are the IDs of the roles the user has.
	// Example: "62b48deb791a15a25c2a0354"
	RoleIDs []string `json:"roles"`
	// Connections are the user's platform connections.
	// Note that the `EmoteSet`s in each connection
	// do not include the emotes themselves.
	Connections []PlatformConnection `json:"connections"`
}

// Owner is the owner of an emote.
// It's a subset of the fields from `User`.
type Owner struct {
	// ID is the user's ID.
	// Example: "621f13b614f489808df5d58e"
	ID string `json:"id"`
	// Username is the user's username.
	// Example: "airforce2700"
	Username string `json:"username"`
	// DisplayName is the user's display name.
	DisplayName string `json:"display_name"`
	// AvatarURL is the URL of the user's avatar.
	// Notably, does not include the scheme and starts with `//`.
	// Example: "//cdn.7tv.app/user/<id>/<img>/3x.webp"
	// If the user is not a 7TV subscriber or does not have an avatar,
	// this value will be "//cdn.7tv.app/"
	AvatarURL string `json:"avatar_url"`
	// Style is this user's style information.
	Style Style `json:"style"`
	// RoleIDs are the IDs of the roles the user has.
	// Example: "62b48deb791a15a25c2a0354"
	RoleIDs []string `json:"roles"`
}

// Style is style information (?)
// Use is unknown.
type Style struct {
	// Paint is the user's active name paint.
	// It appears to always be negative.
	Paint int `json:"color,omitempty"`
}

// EmoteSet is an emote set.
type EmoteSet struct {
	// ID is the unique ID of the set.
	// Example: "621f13b614f489808df5d58e"
	ID string `json:"id"`
	// Name is the display name of the set.
	// Example: "airforce2700's Emotes"
	Name string `json:"name"`
	// Flags are flags set on the set (?)
	// Appears to always be 0.
	// Use is unknown.
	Flags int `json:"flags"`
	// Tags are tags on the emote set.
	// Appears to always be empty.
	// Use is unknown.
	Tags []string `json:"tags"`
	// Immutable is whether the set is immutable.
	// Appears to always be false.
	// Use is unknown.
	Immutable bool `json:"immutable"`
	// Privileged is whether the set is privileged.
	// Appears to always be false.
	// Use is unknown.
	Privileged bool `json:"privileged"`
	// Emotes are the emotes in the set.
	Emotes []Emote `json:"emotes"`
	// EmoteCount is the number of emotes in the set.
	EmoteCount int `json:"emote_count"`
	// Capacity is the capacity of the set.
	Capacity int `json:"capacity"`
	// Owner is the owner of the set.
	Owner Owner `json:"owner"`
}

// Emote is an emote.
type Emote struct {
	// ID is the unique ID of the emote.
	// Example: "6535d68eaf0fd607b5e8e98f"
	ID string `json:"id"`
	// Name is the written name of the emote.
	// Example: "librarySecurity"
	Name string `json:"name"`
	// Flags are flags set on the set (?)
	// Appears to always be 0.
	// Use is unknown.
	Flags int `json:"flags"`
	// UpdateTime is the last update time of the emote (?)
	UpdateTime UnixTimeMs `json:"timestamp"`
	// Creator is the ID of the creator of the emote (?)
	Creator string `json:"actor_id"`
	// Data is extended data about this emote.
	Data EmoteData `json:"data,omitempty"`
}

// EmoteData is extended data about an emote.
type EmoteData struct {
	// ID is the unique ID of the emote.
	// Example: "6535d68eaf0fd607b5e8e98f"
	ID string `json:"id"`
	// Name is the written name of the emote.
	// Example: "librarySecurity"
	Name string `json:"name"`
	// Flags are flags set on the set (?)
	// Appears to always be 0.
	// Use is unknown.
	Flags int `json:"flags"`
	// Flags are flags set on the set (?)
	// Appears to always be 0.
	// Use is unknown.
	Tags []string `json:"tags"`
	// Lifecycle is where this emote is in its lifecycle (?)
	// Mapping to states/strings is unknown.
	// Example: 3
	Lifecycle int `json:"lifecycle"`
	// States describe extended attributes about this emote.
	// Example: "LISTED", "NO_PERSONAL"
	States []string `json:"state"`
	// Listed is whether the emote is listed.
	Listed bool `json:"listed"`
	// Animated is whether the emote is animated.
	Animated bool `json:"animated"`
	// Owner is the owner of the set.
	Owner Owner `json:"owner"`
	// Host is information about where to load the emote from.
	Host Host `json:"host"`
}

// Host is information about where to load an emote from.
type Host struct {
	// BaseURL is the base URL to use for loading the emote.
	// Example: "//cdn.7tv.app/emote/6535d68eaf0fd607b5e8e98f"
	BaseURL string `json:"url"`
	// Files are the individual files to use when loading the emote.
	// There are generally multiple formats and sizes.
	Files []File `json:"files"`
}

// File is information on how to load an emote.
type File struct {
	// Name is the file name.
	// Example: "1x.avif"
	Name string `json:"name"`
	// Static name is the "static name of the emote".
	// Example: "1x_static.avif"
	// Use is unknown.
	StaticName string `json:"static_name"`
	// Width is the width of the emote.
	// Example: 32
	Width int `json:"width"`
	// Height is the height of the emote.
	// Example: 32
	Height int `json:"height"`
	// FrameCount is the number of frames in the emote.
	// If the emote is not animated, it is 1.
	FrameCount int `json:"frame_count"`
	// Size is the size of the file in bytes.
	Size int `json:"size"`
	// Format is the format the image is in.
	// Examples: "AVIF", "WEBP"
	Format string `json:"format"`
}

// UnixTimeMs is the unix timestamp in milliseconds.
// It is the format the 7TV API regularly returns timestamps in.
type UnixTimeMs time.Time

// MarshalJSON marshals this type into JSON.
func (t UnixTimeMs) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(time.Time(t).Unix()*1000, 10)), nil
}

// UnmarshalJSON unmarshals this type from JSON.
func (t *UnixTimeMs) UnmarshalJSON(in []byte) error {
	inInt, err := strconv.ParseInt(string(in), 10, 64)
	if err != nil {
		return err
	}
	*(*time.Time)(t) = time.UnixMilli(inInt).UTC()
	return nil
}
