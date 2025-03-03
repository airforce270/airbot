// Package databasetest provides utilities for database testing.
package databasetest

import (
	"database/sql"
	"log"
	"testing"

	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/utils/ptrs"
)

// New creates a new in-memory database connection for testing.
func New(t *testing.T) (*sql.DB, *database.Queries) {
	t.Helper()
	ctx := t.Context()

	db, queries, err := database.Connect(ctx, log.Default(), ":memory:")
	if err != nil {
		t.Fatalf("Failed to create new in-memory DB: %v", err)
	}

	// if err := database.Migrate(db); err != nil {
	// 	t.Fatalf("Failed to migrate DB: %v", err)
	// }

	for _, user := range []string{"user1", "user2", "user3"} {
		seedTwitchUser(t, queries, user)
	}

	return db, queries
}

// New creates a new in-memory database connection for testing.
func NewDB(t *testing.T) *sql.DB {
	t.Helper()
	db, _ := New(t)
	return db
}

// New creates a new in-memory database connection for testing.
func NewQueries(t *testing.T) *database.Queries {
	t.Helper()
	_, queries := New(t)
	return queries
}

func FirstTwitchUserOrInsert(t testing.TB, db *database.Queries, id, name string) database.User {
	t.Helper()
	u, err := db.SelectTwitchUser(t.Context(), database.SelectTwitchUserParams{
		TwitchID:   ptrs.StringNil(id),
		TwitchName: ptrs.StringNil(name),
	})
	if err != nil {
		t.Fatalf("Failed to select user %s: %v", id, err)
	}
	return u
}

func seedTwitchUser(t testing.TB, db *database.Queries, user string) {
	t.Helper()
	_, err := db.CreateTwitchUser(t.Context(), database.CreateTwitchUserParams{
		TwitchID:   ptrs.StringNil(user),
		TwitchName: ptrs.StringNil(user),
	})
	if err != nil {
		t.Fatalf("Failed to seed user %s: %v", user, err)
	}
}
