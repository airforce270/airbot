// Package databasetest provides utilities for database testing.
package databasetest

import (
	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/database/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var instance *gorm.DB

func init() {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"))
	if err != nil {
		panic(err)
	}
	instance = db
}

// NewFakeDB creates a new connection to the in-memory database for testing.
func NewFakeDBConn() *gorm.DB {
	return instance
}

// NewFakeDB creates a new connection to the in-memory database for testing.
func NewFakeDB() *gorm.DB {
	db := NewFakeDBConn()

	for _, m := range models.AllModels {
		if err := db.Migrator().DropTable(&m); err != nil {
			panic(err)
		}
	}

	if err := database.Migrate(db); err != nil {
		panic(err)
	}

	for _, user := range []string{"user1", "user2", "user3"} {
		if err := seedTwitchUser(db, user); err != nil {
			panic(err)
		}
	}

	return db
}

func seedTwitchUser(db *gorm.DB, user string) error {
	result := db.Create(&models.User{TwitchID: user, TwitchName: user})
	return result.Error
}
