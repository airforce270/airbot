package twitch

import (
	"github.com/airforce270/airbot/permission"

	twitchirc "github.com/gempir/go-twitch-irc/v4"
)

type badge string

func (b badge) String() string { return string(b) }

const (
	broadcasterBadge badge = "broadcaster"
	moderatorBadge   badge = "moderator"
	vipBadge         badge = "vip"
	founderBadge     badge = "founder"
	subscriberBadge  badge = "subscriber"
)

var badgeLevels = map[badge]permission.Level{
	broadcasterBadge: permission.Admin,
	moderatorBadge:   permission.Mod,
	vipBadge:         permission.VIP,
	founderBadge:     permission.AboveNormal,
	subscriberBadge:  permission.AboveNormal,
}

func userHasBadge(u twitchirc.User, b badge) bool {
	for userBadge := range u.Badges {
		if userBadge == b.String() {
			return true
		}
	}
	return false
}
