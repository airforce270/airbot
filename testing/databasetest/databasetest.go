// Package databasetest provides utilities for database testing.
package databasetest

import (
	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/database/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// NewFakeDB creates a new connection to the in-memory database for testing.
func NewFakeDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"))
	if err != nil {
		panic(err)
	}
	for _, m := range models.AllModels {
		db.Migrator().DropTable(&m)
	}
	database.Migrate(db)
	return db
}
