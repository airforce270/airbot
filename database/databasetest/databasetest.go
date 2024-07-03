// Package databasetest provides utilities for database testing.
package databasetest

import (
	"testing"

	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/database/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// New creates a new in-memory database for testing.
func New(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"))
	if err != nil {
		t.Fatalf("Failed to create new in-memory DB: %v", err)
	}

	if err := database.Migrate(db); err != nil {
		t.Fatalf("Failed to migrate DB: %v", err)
	}

	for _, user := range []string{"user1", "user2", "user3"} {
		seedTwitchUser(t, db, user)
	}

	return db
}

func seedTwitchUser(t testing.TB, db *gorm.DB, user string) {
	t.Helper()
	err := db.Create(&models.User{TwitchID: user, TwitchName: user}).Error
	if err != nil {
		t.Fatalf("Failed to seed user %s: %v", user, err)
	}
}
