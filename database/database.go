// Package database handles connections to the database.
package database

import (
	"fmt"
	"strings"

	"github.com/airforce270/airbot/database/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Instance() *gorm.DB {
	if Conn == nil {
		panic("database.Conn is nil!")
	}
	return Conn
}

// Conn is the connection to the database.
// It should be set by main.
var Conn *gorm.DB

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
	gormDB, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		return nil, fmt.Errorf("failed to open DB connection: %w", err)
	}

	db, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB handle: %w", err)
	}
	db.SetMaxOpenConns(100)

	return gormDB, nil
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

// formatDSN formats settings into a DSN for a Postgres GORM connection.
func formatDSN(settings map[string]string) string {
	parts := make([]string, len(settings))
	for key, value := range settings {
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(parts, " ")
}
