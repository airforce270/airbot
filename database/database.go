// Package database handles connections to the database.
//
//go:generate sqlfluff fix .
//go:generate sqlfluff format .
//go:generate sqlfluff fix .
//go:generate sqlfluff format .
//go:generate sqlc generate
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/airforce270/airbot/utils/ptrs"

	_ "github.com/glebarez/sqlite"
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
func Connect(ctx context.Context, logger *log.Logger, dbFile string) (*sql.DB, *Queries, error) {
	db, err := sql.Open("sqlite", dbFile+formatPragmas(pragmas))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open DB connection: %w", err)
	}
	queries := New(db)

	context.AfterFunc(ctx, func() {
		if err := close(db); err != nil {
			logger.Printf("failed to close DB cleanly: %v", err)
		}
	})

	// Remove once running in production
	if _, err := db.Exec("ALTER TABLE cache_bool_items RENAME COLUMN `key` TO keyy"); err != nil {
		log.Printf("Failed to rename cache_bool_items.key, was it already migrated?: %v", err)
	}
	if _, err := db.Exec("ALTER TABLE cache_string_items RENAME COLUMN `key` TO keyy"); err != nil {
		log.Printf("Failed to rename cache_string_items.key, was it already migrated?: %v", err)
	}

	return db, queries, nil
}

func close(db *sql.DB) error {
	if _, err := db.Exec("PRAGMA analysis_limit = 400;"); err != nil {
		return fmt.Errorf("failed to set analysis_limit: %w", err)
	}
	if _, err := db.Exec("PRAGMA optimize;"); err != nil {
		return fmt.Errorf("failed to run optimize: %w", err)
	}

	return nil
}

// Migrate performs GORM auto-migrations for all data models.
// func Migrate(db *gorm.DB) error {
// 	for _, model := range models.AllModels {
// 		if err := db.AutoMigrate(&model); err != nil {
// 			return fmt.Errorf("failed to migrate %+v: %w", model, err)
// 		}
// 	}
// 	return nil
// }

// LeaveChannel leaves a channel.
func LeaveChannel(ctx context.Context, queries *Queries, platformName, channel string) error {
	affectedRows, err := queries.LeaveChannel(ctx, LeaveChannelParams{
		Platform: ptrs.StringNil(platformName),
		Channel:  ptrs.StringNil(strings.ToLower(channel)),
	})
	if err != nil {
		return fmt.Errorf("failed to leave %s/%s: %w", platformName, channel, err)
	}

	if affectedRows == 0 {
		return fmt.Errorf("bot is not in channel %s", channel)
	}

	return nil
}

// CreateGambaTransactions creates gamba transactions.
func CreateGambaTransactions(ctx context.Context, db *sql.DB, queries *Queries, reqs []CreateGambaTransactionParams) ([]GambaTransaction, error) {
	return WithinTxn(db, queries, func(queries *Queries) ([]GambaTransaction, error) {
		var txns []GambaTransaction
		for _, req := range reqs {
			if txn, err := queries.CreateGambaTransaction(ctx, req); err != nil {
				return nil, fmt.Errorf("failed to insert gamba transaction: %w", err)
			} else {
				txns = append(txns, txn)
			}
		}
		return txns, nil
	})
}

// SelectOrCreateTwitchUser selects a twitch user
// or creates a new one if it doesn't exist.
func SelectOrCreateTwitchUser(ctx context.Context, db *sql.DB, queries *Queries, id, name string) (User, error) {
	return WithinTxn(db, queries, func(queries *Queries) (User, error) {
		u, err := queries.SelectTwitchUser(ctx, SelectTwitchUserParams{
			TwitchID:   ptrs.StringNil(id),
			TwitchName: ptrs.StringNil(name),
		})
		if err != nil {
			return User{}, fmt.Errorf("failed to select user %s: %w", id, err)
		}
		if u.ID != 0 {
			return u, nil
		}

		u, err = queries.CreateTwitchUser(ctx, CreateTwitchUserParams{
			TwitchID:   ptrs.StringNil(id),
			TwitchName: ptrs.StringNil(name),
		})
		if err != nil {
			return User{}, fmt.Errorf("failed to create user %s: %w", id, err)
		}

		return u, nil
	})
}

// UpdateOrCreateTwitchUser updates a Twitch user
// or creates a new one if it doesn't exist.
func UpdateOrCreateTwitchUser(ctx context.Context, db *sql.DB, queries *Queries, id, name string) (User, error) {
	return WithinTxn(db, queries, func(queries *Queries) (User, error) {
		user, err := queries.UpdateTwitchUserName(ctx, UpdateTwitchUserNameParams{
			TwitchID:   ptrs.StringNil(id),
			TwitchName: ptrs.StringNil(name),
		})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return User{}, fmt.Errorf("failed to update twitch user name %s: %w", id, err)
		}
		if err == nil {
			return user, nil
		}

		user, err = queries.CreateTwitchUser(ctx, CreateTwitchUserParams{
			TwitchID:   ptrs.StringNil(id),
			TwitchName: ptrs.StringNil(name),
		})
		if err != nil {
			return User{}, fmt.Errorf("failed to create twitch user %s: %w", id, err)
		}

		return user, nil
	})
}

// SelectOrCreateChannelCommandCooldown selects a channel command cooldown
// or creates a new one if it doesn't exist.
func SelectOrCreateChannelCommandCooldown(ctx context.Context, db *sql.DB, queries *Queries, channel, command string) (ChannelCommandCooldown, error) {
	return WithinTxn(db, queries, func(queries *Queries) (ChannelCommandCooldown, error) {
		c, err := queries.SelectChannelCommandCooldown(ctx, SelectChannelCommandCooldownParams{
			Channel: ptrs.StringNil(channel),
			Command: ptrs.StringNil(command),
		})
		if err != nil {
			return ChannelCommandCooldown{}, fmt.Errorf("failed to select channel cooldown %s/%s: %w", channel, command, err)
		}
		if c.ID != 0 {
			return c, nil
		}

		c, err = queries.CreateChannelCommandCooldown(ctx, CreateChannelCommandCooldownParams{
			Channel: ptrs.StringNil(channel),
			Command: ptrs.StringNil(command),
		})
		if err != nil {
			return ChannelCommandCooldown{}, fmt.Errorf("failed to create channel cooldown %s/%s: %w", channel, command, err)
		}

		return c, nil
	})
}

// SelectOrCreateUserCommandCooldown selects a user command cooldown
// or creates a new one if it doesn't exist.
func SelectOrCreateUserCommandCooldown(ctx context.Context, db *sql.DB, queries *Queries, user User, command string) (UserCommandCooldown, error) {
	return WithinTxn(db, queries, func(queries *Queries) (UserCommandCooldown, error) {
		u, err := queries.SelectUserCommandCooldown(ctx, SelectUserCommandCooldownParams{
			UserID:  ptrs.Int64Nil(user.ID),
			Command: ptrs.StringNil(command),
		})
		if err != nil {
			return UserCommandCooldown{}, fmt.Errorf("failed to select user cooldown %v/%s: %w", user.TwitchName, command, err)
		}
		if u.ID != 0 {
			return u, nil
		}

		u, err = queries.CreateUserCommandCooldown(ctx, CreateUserCommandCooldownParams{
			UserID:  ptrs.Int64Nil(user.ID),
			Command: ptrs.StringNil(command),
		})
		if err != nil {
			return UserCommandCooldown{}, fmt.Errorf("failed to create user cooldown %v/%s: %w", user.TwitchName, command, err)
		}

		return u, nil
	})
}

// WithinTxn runs a function within a transaction.
func WithinTxn[T any](db *sql.DB, queries *Queries, f func(q *Queries) (T, error)) (T, error) {
	var zero T

	txn, err := db.Begin()
	if err != nil {
		return zero, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer txn.Rollback()
	q := queries.WithTx(txn)

	val, err := f(q)
	if err != nil {
		return zero, fmt.Errorf("failed to run transaction: %w", err)
	}
	if err := txn.Commit(); err != nil {
		return zero, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return val, nil
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
