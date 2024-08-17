// Package database handles connections to the database.
package database

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/airforce270/airbot/database/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var pragmas = map[string]string{
	"journal_mode": "WAL",
	"synchronous":  "NORMAL",
	"foreign_keys": "ON",

	"user_version": "ON",

	"temp_store": "2",
	"cache_size": "-32000",
}

// Connect creates a connection to the database.
func Connect(ctx context.Context, logger *log.Logger, dbFile string) (*gorm.DB, error) {
	gormDB, err := gorm.Open(sqlite.Open(dbFile + formatPragmas(pragmas)))
	if err != nil {
		return nil, fmt.Errorf("failed to open DB connection: %w", err)
	}
	gormDB.WithContext(ctx)

	context.AfterFunc(ctx, func() {
		if err := close(gormDB); err != nil {
			logger.Printf("failed to close DB cleanly: %v", err)
		}
	})

	db, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB handle: %w", err)
	}
	db.SetMaxOpenConns(100)

	return gormDB, nil
}

func close(db *gorm.DB) error {
	d, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get DB handle: %w", err)
	}

	d.Exec("PRAGMA analysis_limit = 400;")
	d.Exec("PRAGMA optimize;")

	return nil
}

// Migrate performs GORM auto-migrations for all data models.
func Migrate(db *gorm.DB) error {
	for _, model := range models.AllModels {
		if err := db.AutoMigrate(&model); err != nil {
			return fmt.Errorf("failed to migrate %+v: %w", model, err)
		}
	}
	return nil
}

func LeaveChannel(db *gorm.DB, platformName, channel string) error {
	var channels []models.JoinedChannel
	err := db.Where(models.JoinedChannel{Platform: platformName, Channel: strings.ToLower(channel)}).Find(&channels).Error
	if err != nil {
		return fmt.Errorf("failed to leave %s/%s: %w", platformName, channel, err)
	}

	if len(channels) == 0 {
		return fmt.Errorf("bot is not in channel %s", channel)
	}

	for _, c := range channels {
		if err := db.Delete(&c).Error; err != nil {
			return fmt.Errorf("failed to delete channel %s: %w", c.Channel, err)
		}
	}
	return nil
}

func formatPragmas(ps map[string]string) string {
	var out strings.Builder

	var i int
	for p, v := range ps {
		if i == 0 {
			out.WriteString("?")
		} else {
			out.WriteString("&")
		}
		fmt.Fprintf(&out, "_pragma=%s(%s)", p, v)
		i++
	}

	return out.String()
}
