// Package database handles connections to the database.
package database

import (
	"fmt"
	"strings"

	"github.com/airforce270/airbot/database/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Instance is the connection to the database.
// It should be set by main.
var Instance *gorm.DB

// Connect creates a connection to the database.
func Connect(dbname, user, password string) (*gorm.DB, error) {
	settings := map[string]string{
		"host":     "database",
		"dbname":   dbname,
		"user":     user,
		"password": password,
		"port":     "5432",
		"sslmode":  "disable",
		"TimeZone": "UTC",
	}
	dsn := formatDSN(settings)
	return gorm.Open(postgres.Open(dsn))
}

// Migrate performs GORM auto-migrations for all data models.
func Migrate(db *gorm.DB) error {
	for _, model := range models.AllModels {
		if err := db.AutoMigrate(&model); err != nil {
			return fmt.Errorf("failed to migrate %v: %w", model, err)
		}
	}
	return nil
}

func LeaveChannel(db *gorm.DB, platformName, channel string) error {
	var channels []models.JoinedChannel
	db.Where(models.JoinedChannel{Platform: platformName, Channel: strings.ToLower(channel)}).Find(&channels)

	if len(channels) == 0 {
		return fmt.Errorf("bot is not in channel %s", channel)
	}

	for _, c := range channels {
		db.Delete(&c)
	}
	return nil
}

// formatDSN formats settings into a DSN for a Postgres GORM connection.
func formatDSN(settings map[string]string) string {
	parts := make([]string, len(settings))
	for key, value := range settings {
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(parts, " ")
}
