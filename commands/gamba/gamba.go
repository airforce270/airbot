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
	"github.com/airforce270/airbot/permission"

	"gorm.io/gorm"
)

// Commands contains this package's commands.
var Commands = [...]basecommand.Command{
	pointsCommand,
	rouletteCommand,
}

var (
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
		Name:            "roulette",
		AlternateNames:  []string{"r"},
		Help:            "Roulettes some points.",
		Usage:           "$roulette <amount|percent%|all>",
		Permission:      permission.Normal,
		ChannelCooldown: time.Duration(5) * time.Second,
		PrefixOnly:      true,
		Pattern:         rouletteCommandPattern,
		Handler:         roulette,
	}
	roulettePattern = regexp.MustCompile(rouletteCommandPattern.String() + `(all|\d+%|\d+).*`)
)

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
