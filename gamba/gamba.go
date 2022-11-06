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

// StartGrantingPoints starts a loop to grant points to all chatters on an interval.
// This function blocks and should be run within a goroutine.
func StartGrantingPoints(ps map[string]base.Platform, db *gorm.DB) {
	for {
		go grantPoints(ps, db)
		time.Sleep(grantInterval)
	}
}

// grant represents a points grant that may be given to a user.
type grant struct {
	// User is the user the grant is for.
	User models.User
	// IsActive is whether the user is currently active.
	IsActive bool
}

// Persist persists the grant in the database. This is not an idempotent operation.
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

// grantPoints performs a single point grant to all active and inactive users.
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

// getInactiveUsers gets all inactive users to grant points to.
// The users returned are not guaranteed to be inactive, the results returned are overinclusive.
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

// getActiveUsers gets all active users to grant points to.
// The users returned are guaranteed to be active.
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