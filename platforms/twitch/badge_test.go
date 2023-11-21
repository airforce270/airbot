package twitch

import (
	"testing"

	twitchirc "github.com/gempir/go-twitch-irc/v4"
)

func TestUserHasBadge(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc string
		user twitchirc.User
		b    badge
		want bool
	}{
		{
			desc: "broadcaster",
			user: twitchirc.User{
				Badges: map[string]int{
					"broadcaster": 1,
				},
			},
			b:    broadcasterBadge,
			want: true,
		},
		{
			desc: "non-broadcaster",
			user: twitchirc.User{
				Badges: map[string]int{},
			},
			b:    broadcasterBadge,
			want: false,
		},
		{
			desc: "subscribed mod",
			user: twitchirc.User{
				Badges: map[string]int{
					"moderator":  1,
					"subscriber": 1,
				},
			},
			b:    moderatorBadge,
			want: true,
		},
		{
			desc: "unsubscribed mod",
			user: twitchirc.User{
				Badges: map[string]int{
					"moderator": 1,
				},
			},
			b:    moderatorBadge,
			want: true,
		},
		{
			desc: "non-mod",
			user: twitchirc.User{
				Badges: map[string]int{},
			},
			b:    moderatorBadge,
			want: false,
		},
		{
			desc: "subscribed vip",
			user: twitchirc.User{
				Badges: map[string]int{
					"vip":        1,
					"subscriber": 1,
				},
			},
			b:    vipBadge,
			want: true,
		},
		{
			desc: "unsubscribed vip",
			user: twitchirc.User{
				Badges: map[string]int{
					"vip": 1,
				},
			},
			b:    vipBadge,
			want: true,
		},
		{
			desc: "non-vip",
			user: twitchirc.User{
				Badges: map[string]int{
					"subscriber": 1,
				},
			},
			b:    vipBadge,
			want: false,
		},
		{
			desc: "founder",
			user: twitchirc.User{
				Badges: map[string]int{
					"founder": 1,
				},
			},
			b:    founderBadge,
			want: true,
		},
		{
			desc: "subscriber",
			user: twitchirc.User{
				Badges: map[string]int{
					"subscriber": 1,
				},
			},
			b:    subscriberBadge,
			want: true,
		},
		{
			desc: "non-subscriber",
			user: twitchirc.User{
				Badges: map[string]int{},
			},
			b:    subscriberBadge,
			want: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			if got := userHasBadge(tc.user, tc.b); got != tc.want {
				t.Errorf("userHasBadge() = %v, want %v", got, tc.want)
			}
		})
	}
}
