// Package databasetest provides utilities for database testing.
package databasetest

import (
	"database/sql"
	"log"
	"testing"

	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/database/models"

	"gorm.io/gorm"
)

// New creates a new in-memory database for testing.
func New(t *testing.T) *gorm.DB {
	t.Helper()
	ctx := t.Context()

	db, err := database.Connect(ctx, log.Default(), ":memory:")
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

// New2 creates a new in-memory database for testing.
func New2(t *testing.T) *database.Queries {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create new in-memory DB: %v", err)
	}

	q := database.New(db)

	for _, user := range []string{"user1", "user2", "user3"} {
		seedTwitchUser2(t, q, user)
	}

	return q
}

func FirstTwitchUserOrInsert(t testing.TB, db *database.Queries, id, name, string) database.User {
	t.Helper()
  u, err := db.SelectTwitchUser(t.Context(), sql.NullString{String: id, Valid: true}, sql.NullString{String: name, Valid: true})
	return user1
}

func seedTwitchUser(t testing.TB, db *gorm.DB, user string) {
	t.Helper()
	err := db.Create(&models.User{TwitchID: user, TwitchName: user}).Error
	if err != nil {
		t.Fatalf("Failed to seed user %s: %v", user, err)
	}
}
func seedTwitchUser2(t testing.TB, db *database.Queries, user string) {
	t.Helper()
	err := db.CreateTwitchUser(t.Context(), database.CreateTwitchUserParams{
		TwitchID:   sql.NullString{String: user, Valid: true},
		TwitchName: sql.NullString{String: user, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to seed user %s: %v", user, err)
	}
}
