// Package gamba implements gamba commands.
package gamba

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/commands/basecommand"
	"github.com/airforce270/airbot/database"
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
	pointsCommand,
	rouletteCommand,
}

var (
	acceptCommand = basecommand.Command{
		Name:       "accept",
		Help:       "Accepts a duel.",
		Usage:      "$accept",
		Permission: permission.Normal,
		PrefixOnly: true,
		Pattern:    basecommand.PrefixPattern("accept"),
		Handler:    accept,
	}

	declineCommand = basecommand.Command{
		Name:       "decline",
		Help:       "Declines a duel.",
		Usage:      "$decline",
		Permission: permission.Normal,
		PrefixOnly: true,
		Pattern:    basecommand.PrefixPattern("decline"),
		Handler:    decline,
	}

	duelPendingSecs     = 30
	duelPendingDuration = time.Duration(duelPendingSecs) * time.Second
	duelCommandPattern  = basecommand.PrefixPattern("duel")
	duelCommandUsage    = "$points <user>"
	duelCommand         = basecommand.Command{
		Name:         "duel",
		Help:         fmt.Sprintf("Duels another chatter. They have %d seconds to accept or decline.", duelPendingSecs),
		Usage:        duelCommandUsage,
		Permission:   permission.Normal,
		UserCooldown: time.Duration(5) * time.Second,
		PrefixOnly:   true,
		Pattern:      duelCommandPattern,
		Handler:      duel,
	}
	duelPattern = regexp.MustCompile(duelCommandPattern.String() + `@?(\w+)\s+(\d+).*`)

	pointsCommandPattern = basecommand.PrefixPattern("(?:p(?: |$)|points)")
	pointsCommand        = basecommand.Command{
		Name:           "points",
		AlternateNames: []string{"p"},
		Help:           "Checks how many points you have.",
		Usage:          "$points [user]",
		Permission:     permission.Normal,
		PrefixOnly:     true,
		Pattern:        pointsCommandPattern,
		Handler:        points,
	}
	pointsPattern = regexp.MustCompile(pointsCommandPattern.String() + `@?(\w+).*`)

	rouletteCommandPattern = basecommand.PrefixPattern("(?:r(?: |$)|roulette)")
	rouletteCommand        = basecommand.Command{
		Name:           "roulette",
		AlternateNames: []string{"r"},
		Help:           "Roulettes some points.",
		Usage:          "$roulette <amount|percent%|all>",
		Permission:     permission.Normal,
		UserCooldown:   time.Duration(5) * time.Second,
		PrefixOnly:     true,
		Pattern:        rouletteCommandPattern,
		Handler:        roulette,
	}
	roulettePattern = regexp.MustCompile(rouletteCommandPattern.String() + `(all|\d+%|\d+).*`)
)

func accept(msg *base.IncomingMessage) ([]*base.Message, error) {
	db := database.Instance
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	user, err := msg.Platform.User(msg.Message.User)
	if err != nil {
		return nil, err
	}

	pendingDuels, err := gamba.InboundPendingDuels(&user, duelPendingDuration, db)
	if err != nil {
		return nil, err
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
		randInt, err := rand.Int(base.RandReader, big.NewInt(2))
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
		result := db.Save(&pendingDuel)
		if result.Error != nil {
			log.Printf("failed to persist duel acceptance: %v", result.Error)
		}

		result = db.Create(&models.GambaTransaction{
			Game:  "Duel",
			User:  *winner,
			Delta: pendingDuel.Amount,
		})
		if result.Error != nil {
			log.Printf("failed to insert gamba transaction: %v", result.Error)
		}
		result = db.Create(&models.GambaTransaction{
			Game:  "Duel",
			User:  *loser,
			Delta: -pendingDuel.Amount,
		})
		if result.Error != nil {
			log.Printf("failed to insert gamba transaction: %v", result.Error)
		}

		outMsgs = append(outMsgs, &base.Message{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("%s won the duel with %s and wins %d points!", winner.TwitchName, loser.TwitchName, pendingDuel.Amount),
		})
	}

	return outMsgs, nil
}

func decline(msg *base.IncomingMessage) ([]*base.Message, error) {
	db := database.Instance
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	user, err := msg.Platform.User(msg.Message.User)
	if err != nil {
		return nil, err
	}

	pendingDuels, err := gamba.InboundPendingDuels(&user, duelPendingDuration, db)
	if err != nil {
		return nil, err
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
		result := db.Save(&pendingDuel)
		if result.Error != nil {
			log.Printf("failed to persist duel declining: %v", result.Error)
		}
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    "Declined duel.",
		},
	}, nil
}

func duel(msg *base.IncomingMessage) ([]*base.Message, error) {
	db := database.Instance
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	matches := duelPattern.FindStringSubmatch(msg.MessageTextWithoutPrefix())
	if len(matches) < 3 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "Usage: " + duelCommandUsage,
			},
		}, nil
	}

	target := matches[1]
	if target == msg.Message.User {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "You can't duel yourself Pepega",
			},
		}, nil
	}

	pointsStr := matches[2]
	points, err := strconv.ParseInt(pointsStr, 10, 64)
	if err != nil {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "Couldn't parse duel amount.",
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

	targetUser, err := msg.Platform.User(target)
	if err != nil {
		if errors.Is(err, base.ErrUserUnknown) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    fmt.Sprintf("%s has never been seen by %s", target, msg.Platform.Username()),
				},
			}, nil
		}
		return nil, err
	}
	user, err := msg.Platform.User(msg.Message.User)
	if err != nil {
		if errors.Is(err, base.ErrUserUnknown) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    fmt.Sprintf("%s has never been seen by %s", target, msg.Platform.Username()),
				},
			}, nil
		}
		return nil, err
	}

	userPoints := fetchUserPoints(db, user)
	if points > userPoints {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("You don't have enough points for that duel (you have %d points)", userPoints),
			},
		}, nil
	}

	targetUserPoints := fetchUserPoints(db, targetUser)
	if points > targetUserPoints {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    fmt.Sprintf("%s don't have enough points for that duel (they have %d points)", target, targetUserPoints),
			},
		}, nil
	}

	var currentPendingDuels []models.Duel
	result := db.Where(models.Duel{Pending: true}).Find(&currentPendingDuels)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to retrieve current pending duels: %w", result.Error)
	}
	for _, currentPendingDuel := range currentPendingDuels {
		if currentPendingDuel.UserID == user.ID {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    "You already have a duel pending.",
				},
			}, nil
		}
		if currentPendingDuel.TargetID == targetUser.ID {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    "That chatter already has a duel pending.",
				},
			}, nil
		}
	}

	result = db.Create(&models.Duel{
		UserID:   user.ID,
		User:     user,
		TargetID: targetUser.ID,
		Target:   targetUser,
		Amount:   points,
		Pending:  true,
		Accepted: false,
		Won:      false,
	})
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create pending duel: %w", result.Error)
	}

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("@%s, %s has started a duel for %d points! Type %saccept or %sdecline in the next %d seconds!", target, msg.Message.User, points, msg.Prefix, msg.Prefix, duelPendingSecs),
		},
	}, nil
}

func points(msg *base.IncomingMessage) ([]*base.Message, error) {
	db := database.Instance
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	target := basecommand.ParseTarget(msg, pointsPattern)
	user, err := msg.Platform.User(target)
	if err != nil {
		if errors.Is(err, base.ErrUserUnknown) {
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    fmt.Sprintf("%s has never been seen by %s", target, msg.Platform.Username()),
				},
			}, nil
		}
		return nil, err
	}

	pointsCount := fetchUserPoints(db, user)

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("GAMBA %s has %d points", target, pointsCount),
		},
	}, nil
}

func roulette(msg *base.IncomingMessage) ([]*base.Message, error) {
	db := database.Instance
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	user, err := msg.Platform.User(msg.Message.User)
	if err != nil {
		if errors.Is(err, base.ErrUserUnknown) {
			// This should never happen, as the incoming message should have been logged already
			return []*base.Message{
				{
					Channel: msg.Message.Channel,
					Text:    fmt.Sprintf("%s has never been seen by %s", msg.Message.User, msg.Platform.Username()),
				},
			}, nil
		}
		return nil, err
	}

	points := fetchUserPoints(db, user)

	var amount int64
	amountStr := basecommand.ParseTarget(msg, roulettePattern)
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

	randInt, err := rand.Int(base.RandReader, big.NewInt(2))
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
		result := db.Create(&txn)
		if result.Error != nil {
			log.Printf("failed to insert gamba transaction: %v", result.Error)
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

func fetchUserPoints(db *gorm.DB, user models.User) int64 {
	var transactions []*models.GambaTransaction
	db.Where(models.GambaTransaction{UserID: user.ID}).Find(&transactions)

	var points int64
	for _, txn := range transactions {
		points += txn.Delta
	}

	return points
}
