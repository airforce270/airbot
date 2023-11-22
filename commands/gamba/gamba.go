// Package gamba implements gamba commands.
package gamba

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/base/arg"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/database/models"
	"github.com/airforce270/airbot/gamba"
	"github.com/airforce270/airbot/permission"

	"gorm.io/gorm"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	acceptCommand,
	declineCommand,
	duelCommand,
	givePointsCommand,
	pointsCommand,
	rouletteCommand,
}

var (
	acceptCommand = basecommand.Command{
		Name:       "accept",
		Desc:       "Accepts a duel.",
		Permission: permission.Normal,
		Handler:    accept,
	}

	declineCommand = basecommand.Command{
		Name:       "decline",
		Desc:       "Declines a duel.",
		Permission: permission.Normal,
		Handler:    decline,
	}

	duelCommand = basecommand.Command{
		Name: "duel",
		Desc: fmt.Sprintf("Duels another chatter. They have %d seconds to accept or decline.", duelPendingSecs),
		Params: []arg.Param{
			{Name: "user", Type: arg.Username, Required: true},
			{Name: "amount", Type: arg.Int, Required: true},
		},
		Permission:   permission.Normal,
		UserCooldown: 5 * time.Second,
		Handler:      duel,
	}

	givePointsCommand = basecommand.Command{
		Name:    "givepoints",
		Aliases: []string{"gp"},
		Desc:    "Give points to another chatter.",
		Params: []arg.Param{
			{Name: "user", Type: arg.Username, Required: true},
			{Name: "amount", Type: arg.Int, Required: true},
		},
		Permission: permission.Normal,
		Handler:    givePoints,
	}

	pointsCommand = basecommand.Command{
		Name:       "points",
		Aliases:    []string{"p"},
		Desc:       "Checks how many points someone has.",
		Params:     []arg.Param{{Name: "user", Type: arg.Username, Required: false}},
		Permission: permission.Normal,
		Handler:    points,
	}

	rouletteCommand = basecommand.Command{
		Name:         "roulette",
		Aliases:      []string{"r"},
		Desc:         "Roulettes some points.",
		Params:       []arg.Param{{Name: "amount", Type: arg.String, Required: true, Usage: "amount|percent%|all"}},
		Permission:   permission.Normal,
		UserCooldown: 5 * time.Second,
		Handler:      roulette,
	}
)

const (
	duelPendingSecs     = 30
	duelPendingDuration = duelPendingSecs * time.Second
)

func accept(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	user, err := msg.Resources.Platform.User(msg.Message.User)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s user %s: %w", msg.Resources.Platform.Name(), msg.Message.User, err)
	}

	pendingDuels, err := gamba.InboundPendingDuels(&user, duelPendingDuration, msg.Resources.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch inbound pending duels for %s user %s: %w", msg.Resources.Platform.Name(), msg.Message.User, err)
	}
	if len(pendingDuels) == 0 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "There are no duels pending against you.",
			},
		}, nil
	}

	var outMsgs []*base.Message

	for _, pendingDuel := range pendingDuels {
		randInt, err := rand.Int(msg.Resources.Rand.Reader, big.NewInt(2))
		if err != nil {
			return nil, fmt.Errorf("failed to read random number: %w", err)
		}

		var winner, loser *models.User
		if initiatorWin := randInt.Int64() == 1; initiatorWin {
			winner = &pendingDuel.User
			loser = &pendingDuel.Target
			pendingDuel.Won = true
		} else {
			winner = &pendingDuel.Target
			loser = &pendingDuel.User
			pendingDuel.Won = false
		}

		pendingDuel.Accepted = true
		pendingDuel.Pending = false
		if err := msg.Resources.DB.Save(&pendingDuel).Error; err != nil {
			log.Printf("failed to persist duel acceptance: %v", err)
		}

		err = msg.Resources.DB.Create(&[]models.GambaTransaction{
			{
				Game:  "Duel",
				User:  *winner,
				Delta: pendingDuel.Amount,
			},
			{
				Game:  "Duel",
				User:  *loser,
				Delta: -pendingDuel.Amount,
			},
		}).Error
		if err != nil {
			log.Printf("failed to insert gamba transactions: %v", err)
		}

		outMsgs = append(outMsgs, &base.Message{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("%s won the duel with %s and wins %d points!", winner.TwitchName, loser.TwitchName, pendingDuel.Amount),
		})
	}

	return outMsgs, nil
}

func decline(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	user, err := msg.Resources.Platform.User(msg.Message.User)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s user %s: %w", msg.Resources.Platform.Name(), msg.Message.User, err)
	}

	pendingDuels, err := gamba.InboundPendingDuels(&user, duelPendingDuration, msg.Resources.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch inbound pending duels for %s user %s: %w", msg.Resources.Platform.Name(), msg.Message.User, err)
	}
	if len(pendingDuels) == 0 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "There are no duels pending against you.",
			},
		}, nil
	}

	for _, pendingDuel := range pendingDuels {
		pendingDuel.Accepted = false
		pendingDuel.Pending = false
		if err := msg.Resources.DB.Save(&pendingDuel).Error; err != nil {
			log.Printf("failed to persist duel declining: %v", err)
		}
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    "Declined duel.",
		},
	}, nil
}

func duel(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	targetArg, pointsArg := args[0], args[1]
	if !targetArg.Present || !pointsArg.Present {
		return nil, basecommand.ErrBadUsage
	}
	target, points := targetArg.StringValue, pointsArg.IntValue

	if target == msg.Message.User {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "You can't duel yourself Pepega",
			},
		}, nil
	}

	if points == 0 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "You must duel at least 1 point.",
			},
		}, nil
	}
	if points < 1 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "nice try forsenCD",
			},
		}, nil
	}

	targetUser, err := msg.Resources.Platform.User(target)
	if err != nil {
		if errors.Is(err, base.ErrUserUnknown) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    fmt.Sprintf("%s has never been seen by %s", target, msg.Resources.Platform.Username()),
				},
			}, nil
		}
		return nil, fmt.Errorf("user %s on %s is unknown to the bot: %w", target, msg.Resources.Platform.Name(), err)
	}
	user, err := msg.Resources.Platform.User(msg.Message.User)
	if err != nil {
		if errors.Is(err, base.ErrUserUnknown) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    fmt.Sprintf("%s has never been seen by %s", target, msg.Resources.Platform.Username()),
				},
			}, nil
		}
		return nil, fmt.Errorf("user %s on %s is unknown to the bot: %w", msg.Message.User, msg.Resources.Platform.Name(), err)
	}

	userPoints, err := FetchUserPoints(msg.Resources.DB, user)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user points for user %d: %w", user.ID, err)
	}
	if int64(points) > userPoints {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("You don't have enough points for that duel (you have %d points)", userPoints),
			},
		}, nil
	}

	targetUserPoints, err := FetchUserPoints(msg.Resources.DB, targetUser)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user points for user %d: %w", targetUser.ID, err)
	}
	if int64(points) > targetUserPoints {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("%s don't have enough points for that duel (they have %d points)", target, targetUserPoints),
			},
		}, nil
	}

	outboundPendingDuels, err := gamba.OutboundPendingDuels(&user, duelPendingDuration, msg.Resources.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch outbound pending duels for user %d: %w", user.ID, err)
	}
	if len(outboundPendingDuels) > 0 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "You already have a duel pending.",
			},
		}, nil
	}

	inboundPendingDuels, err := gamba.InboundPendingDuels(&targetUser, duelPendingDuration, msg.Resources.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch inbound pending duels for user %d: %w", targetUser.ID, err)
	}
	if len(inboundPendingDuels) > 0 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "That chatter already has a duel pending.",
			},
		}, nil
	}

	err = msg.Resources.DB.Create(&models.Duel{
		UserID:   user.ID,
		User:     user,
		TargetID: targetUser.ID,
		Target:   targetUser,
		Amount:   int64(points),
		Pending:  true,
		Accepted: false,
		Won:      false,
	}).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create pending duel: %w", err)
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("@%s, %s has started a duel for %d points! Type %saccept or %sdecline in the next %d seconds!", target, msg.Message.User, points, msg.Prefix, msg.Prefix, duelPendingSecs),
		},
	}, nil
}

func givePoints(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	targetArg, pointsArg := args[0], args[1]
	if !targetArg.Present || !pointsArg.Present {
		return nil, basecommand.ErrBadUsage
	}
	target, points := targetArg.StringValue, pointsArg.IntValue

	if len(args) < 2 {
		return nil, basecommand.ErrBadUsage
	}

	if target == msg.Message.User {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "You can't give points to yourself Pepega",
			},
		}, nil
	}

	if points == 0 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "You must give at least 1 point.",
			},
		}, nil
	}
	if points < 1 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "nice try forsenCD",
			},
		}, nil
	}

	targetUser, err := msg.Resources.Platform.User(target)
	if err != nil {
		if errors.Is(err, base.ErrUserUnknown) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    fmt.Sprintf("%s has never been seen by %s", target, msg.Resources.Platform.Username()),
				},
			}, nil
		}
		return nil, fmt.Errorf("user %s on %s is unknown to the bot: %w", target, msg.Resources.Platform.Name(), err)
	}
	user, err := msg.Resources.Platform.User(msg.Message.User)
	if err != nil {
		if errors.Is(err, base.ErrUserUnknown) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    fmt.Sprintf("%s has never been seen by %s", target, msg.Resources.Platform.Username()),
				},
			}, nil
		}
		return nil, fmt.Errorf("failed to retrieve db user %s: %w", msg.Message.User, err)
	}

	userPoints, err := FetchUserPoints(msg.Resources.DB, user)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch points for user %d: %w", user.ID, err)
	}
	if int64(points) > userPoints {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("You can't give more points than you have (you have %d points)", userPoints),
			},
		}, nil
	}

	err = msg.Resources.DB.Create(&[]models.GambaTransaction{
		{
			Game:  "GivePoints",
			User:  user,
			Delta: -int64(points),
		},
		{
			Game:  "GivePoints",
			User:  targetUser,
			Delta: int64(points),
		},
	}).Error
	if err != nil {
		log.Printf("failed to insert gamba transactions: %v", err)
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("%s gave %d points to %s FeelsOkayMan <3", user.TwitchName, points, targetUser.TwitchName),
		},
	}, nil
}

func points(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	target := basecommand.FirstArgOrUsername(args, msg)
	user, err := msg.Resources.Platform.User(target)
	if err != nil {
		if errors.Is(err, base.ErrUserUnknown) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    fmt.Sprintf("%s has never been seen by %s", target, msg.Resources.Platform.Username()),
				},
			}, nil
		}
		return nil, fmt.Errorf("user %s on %s is unknown to the bot: %w", target, msg.Resources.Platform.Name(), err)
	}

	pointsCount, err := FetchUserPoints(msg.Resources.DB, user)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch points for user %d: %w", user.ID, err)
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("GAMBA %s has %d points", target, pointsCount),
		},
	}, nil
}

func roulette(msg *base.IncomingMessage, args []arg.Arg) ([]*base.Message, error) {
	amountArg := args[0]
	if !amountArg.Present {
		return nil, basecommand.ErrBadUsage
	}

	user, err := msg.Resources.Platform.User(msg.Message.User)
	if err != nil {
		if errors.Is(err, base.ErrUserUnknown) {
			// This should never happen, as the incoming message should have been logged already
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    fmt.Sprintf("%s has never been seen by %s", msg.Message.User, msg.Resources.Platform.Username()),
				},
			}, nil
		}
		return nil, fmt.Errorf("user %s on %s is unknown to the bot: %w", msg.Message.User, msg.Resources.Platform.Name(), err)
	}

	points, err := FetchUserPoints(msg.Resources.DB, user)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch points for user %d: %w", user.ID, err)
	}

	var amount int64
	amountStr := amountArg.StringValue
	if amountStr == "all" {
		amount = points
	} else if strings.HasSuffix(amountStr, "%") {
		percent, err := strconv.ParseInt(strings.Replace(amountStr, "%", "", 1), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse roulette amount percent %q: %w", amountStr, err)
		}
		amount = int64(math.Floor(float64(points) * (float64(percent) / 100)))
	} else {
		var err error
		amount, err = strconv.ParseInt(amountStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse roulette amount %q: %w", amountStr, err)
		}
	}
	if amount < 0 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "nice try forsenCD",
			},
		}, nil
	} else if amount > points {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("%s: You don't have enough points for that (current: %d)", msg.Message.User, points),
			},
		}, nil
	}

	randInt, err := rand.Int(msg.Resources.Rand.Reader, big.NewInt(2))
	if err != nil {
		return nil, fmt.Errorf("failed to read random number: %w", err)
	}

	win := randInt.Int64() == 1
	delta := amount
	if !win {
		delta = -amount
	}
	newPoints := points + delta

	go func() {
		txn := models.GambaTransaction{
			Game:  "Roulette",
			User:  user,
			Delta: delta,
		}
		if err := msg.Resources.DB.Create(&txn).Error; err != nil {
			log.Printf("failed to insert gamba transaction: %v", err)
		}
	}()

	outMsg := &base.Message{Channel: msg.Message.Channel}
	if win {
		outMsg.Text = fmt.Sprintf("GAMBA %s won %d points in roulette and now has %d points!", msg.Message.User, delta, newPoints)
	} else {
		outMsg.Text = fmt.Sprintf("GAMBA %s lost %d points in roulette and now has %d points!", msg.Message.User, -delta, newPoints)
	}
	return []*base.Message{outMsg}, nil
}

// FetchUserPoints fetches user points. Only exported for testing, do not use.
func FetchUserPoints(db *gorm.DB, user models.User) (int64, error) {
	var points int64
	if err := db.Model(&models.GambaTransaction{}).Select("COALESCE(SUM(delta), 0)").Where(models.GambaTransaction{UserID: user.ID}).Scan(&points).Error; err != nil {
		return 0, fmt.Errorf("failed to fetch points for user %d: %w", user.ID, err)
	}
	return points, nil
}
