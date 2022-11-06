// Package gamba implements gamba commands.
package gamba

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"regexp"
	"strconv"
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

	rouletteCommandPattern = basecommand.PrefixPattern("roulette")
	rouletteCommand        = basecommand.Command{
		Name:            "roulette",
		AlternateNames:  []string{"r"},
		Help:            "Roulettes some points.",
		Usage:           "$roulette <amount>",
		Permission:      permission.Normal,
		ChannelCooldown: time.Duration(5) * time.Second,
		PrefixOnly:      true,
		Pattern:         rouletteCommandPattern,
		Handler:         roulette,
	}
	roulettePattern = regexp.MustCompile(rouletteCommandPattern.String() + `(\d+).*`)
)

func points(msg *base.IncomingMessage) ([]*base.Message, error) {
	db := database.Instance
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	target := basecommand.ParseTarget(msg, pointsPattern)
	var user models.User
	db.Where(models.User{TwitchID: msg.Message.UserID}).First(&user)

	pointsCount := fetchUserPoints(db, user)

	return []*base.Message{
		{
			Channel: msg.Message.Channel,
			Text:    fmt.Sprintf("GAMBA %s has %d points", target, pointsCount),
		},
	}, nil
}

func roulette(msg *base.IncomingMessage) ([]*base.Message, error) {
	amountStr := basecommand.ParseTarget(msg, roulettePattern)
	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse roulette amount %q: %w", amountStr, err)
	}
	if amount < 0 {
		return []*base.Message{
			{
				Channel: msg.Message.Channel,
				Text:    "nice try forsenCD",
			},
		}, nil
	}

	db := database.Instance
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	var user models.User
	result := db.Where(models.User{TwitchID: msg.Message.UserID}).Find(&user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to lookup user in db: %w", result.Error)
	}

	points := fetchUserPoints(db, user)
	if amount > points {
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
