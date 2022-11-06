// Package gamba handles gamba-related things.
package gamba

import (
	"log"
	"time"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/database/models"
	"golang.org/x/exp/slices"

	"gorm.io/gorm"
)

var (
	grantInterval       = time.Duration(10) * time.Minute
	activeGrantAmount   = 10
	inactiveGrantAmount = 3
)

func StartGrantingPoints(ps map[string]base.Platform, db *gorm.DB) {
	for {
		go grantPoints(ps, db)
		time.Sleep(grantInterval)
	}
}

type grant struct {
	User     models.User
	IsActive bool
}

// Persist persists the grant in the database. This is not idempotent.
func (g grant) Persist(db *gorm.DB) error {
	amount := activeGrantAmount
	if !g.IsActive {
		amount = inactiveGrantAmount
	}
	result := db.Create(&models.GambaTransaction{
		User:  g.User,
		Game:  "AutomaticGrant",
		Delta: int64(amount),
	})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GrantPoints performs a single point grant to all active and inactive users.
func grantPoints(ps map[string]base.Platform, db *gorm.DB) {
	var grants []grant

	for _, inactiveUser := range getInactiveUsers(ps, db) {
		grants = append(grants, grant{
			User:     inactiveUser,
			IsActive: false,
		})
	}

	activeUsers, err := getActiveUsers(db)
	if err != nil {
		log.Printf("Failed to fetch users: %v", err)
	}
	for _, activeUser := range activeUsers {
		grants = append(grants, grant{
			User:     activeUser,
			IsActive: true,
		})
	}

	grants = deduplicateByUser(grants)

	for _, g := range grants {
		err := g.Persist(db)
		if err != nil {
			log.Printf("Failed to grant points - failed to persist grant: %v", err)
			return
		}
	}
}

func getInactiveUsers(ps map[string]base.Platform, db *gorm.DB) []models.User {
	var users []models.User
	for _, p := range ps {
		allUsers, err := p.Users()
		if err != nil {
			log.Printf("Failed to retrieve users from %s: %v", p.Name(), err)
			continue
		}
		for _, u := range allUsers {
			var user models.User
			result := db.Where(models.User{TwitchID: u}).First(&user)
			if result.Error != nil {
				log.Printf("Failed to look up %s user %s in database: %v", p.Name(), u, result.Error)
				continue
			}
			users = append(users, user)
		}
	}
	return users
}

func getActiveUsers(db *gorm.DB) ([]models.User, error) {
	var recentMessagesUniqueByUser []models.Message
	result := db.Select("user_id").Distinct("user_id").Where("time > ?", time.Now().Add(-grantInterval)).Find(&recentMessagesUniqueByUser)
	if result.Error != nil {
		return nil, result.Error
	}
	var recentUserIDs []uint
	for _, m := range recentMessagesUniqueByUser {
		recentUserIDs = append(recentUserIDs, m.UserID)
	}

	var activeUsers []models.User
	result = db.Where("id IN ?", recentUserIDs).Find(&activeUsers)
	if result.Error != nil {
		return nil, result.Error
	}
	return activeUsers, nil
}

func deduplicateByUser(grants []grant) []grant {
	sorted := grants
	slices.SortStableFunc(sorted, func(g1, g2 grant) bool {
		return g1.IsActive && !g2.IsActive
	})
	var deduped []grant
	var grantedIDs []uint
	for _, g := range sorted {
		if slices.Contains(grantedIDs, g.User.ID) {
			continue
		}
		grantedIDs = append(grantedIDs, g.User.ID)
		deduped = append(deduped, g)
	}
	return deduped
}
