// Package main contains the main logic for airbot.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/airforce270/airbot/apiclients/supinic"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/cache"
	"github.com/airforce270/airbot/config"
	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/gamba"
	"github.com/airforce270/airbot/platforms"
	"github.com/airforce270/airbot/utils/cleanup"
	"github.com/airforce270/airbot/utils/restart"
)

const (
	waitForCancelFuncs         = 100 * time.Millisecond
	waitForContextCancellation = 100 * time.Millisecond
)

type postStartupResources struct {
	platforms map[string]base.Platform
	cache     cache.Cache
}

func initialStart(ctx context.Context) (cleanup.Cleaner, error) {
	cleaner, _, err := start(ctx)
	return cleaner, err
}

func reStart(ctx context.Context) (cleanup.Cleaner, error) {
	cleaner, resources, err := start(ctx)
	if err != nil {
		return nil, err
	}
	if err := restart.Notify(resources.cache, resources.platforms); err != nil {
		log.Printf("Failed to notify restart: %v", err)
	}
	return cleaner, err
}

func start(ctx context.Context) (cleanup.Cleaner, postStartupResources, error) {
	cleaner := cleanup.NewCleaner()

	log.Print("Reading config...")
	configSrc, err := config.DefaultNewConfigSource()
	if err != nil {
		return nil, postStartupResources{}, fmt.Errorf("failed to open config source: %w", err)
	}
	cfg, err := config.Read(configSrc)
	if err != nil {
		configSrc.Close()
		return nil, postStartupResources{}, fmt.Errorf("failed to read config: %w", err)
	}
	configSrc.Close()

	log.Printf("Connecting to database...")
	db, err := database.Connect(ctx, os.Getenv("POSTGRES_DB"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"))
	if err != nil {
		return nil, postStartupResources{}, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Printf("Connecting to cache...")
	cdb := cache.NewRedis()

	log.Printf("Performing database migrations...")
	if err = database.Migrate(db); err != nil {
		return nil, postStartupResources{}, fmt.Errorf("failed to perform database migrations: %w", err)
	}

	log.Printf("Preparing chat connections...")
	ps, err := platforms.Build(cfg, db, &cdb)
	if err != nil {
		return nil, postStartupResources{}, fmt.Errorf("failed to build platforms: %w", err)
	}

	for _, p := range ps {
		log.Printf("Connecting to %s...", p.Name())
		if err := p.Connect(ctx); err != nil {
			return cleaner, postStartupResources{}, fmt.Errorf("failed to connect to %s: %w", p.Name(), err)
		}

		log.Printf("Starting to handle messages on %s...", p.Name())
		go platforms.StartHandling(ctx, p, db, &cdb, cfg, ps, cfg.LogIncoming, cfg.LogOutgoing)
		cleaner.Register(cleanup.Func{Name: p.Name(), F: p.Disconnect})
	}

	go gamba.StartGrantingPoints(ctx, ps, db)

	if cfg.Supinic.IsConfigured() && cfg.Supinic.ShouldPingAPI {
		log.Println("Starting to ping the Supinic API...")
		supinicClient := supinic.NewClient(cfg.Supinic.UserID, cfg.Supinic.APIKey)
		go supinicClient.StartPinging(ctx)
	}

	return cleaner, postStartupResources{cache: &cdb, platforms: ps}, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		cancel()
		time.Sleep(waitForContextCancellation)
		os.Exit(1)
	}()

	cleaner, err := initialStart(ctx)
	if err != nil {
		log.Fatalf("Failed to start: %v", err)
	}
	log.Printf("Airbot is now running.")

	for {
		select {
		case <-restart.C:
			log.Printf("Restarting...")

			if err := cleaner.Cleanup(); err != nil {
				log.Printf("Cleanup failed: %v", err)
			}
			time.Sleep(waitForCancelFuncs)
			cancel()
			time.Sleep(waitForContextCancellation)

			ctx, cancel = context.WithCancel(context.Background())

			cleaner, err = reStart(ctx)
			if err != nil {
				log.Fatalf("Failed to start: %v", err)
			}
			log.Printf("Airbot is now running (restarted).")
		case <-ctx.Done():
			log.Printf("Context cancelled, Airbot shutting down.")
			return
		}
	}
}

// send message that says "Restarted" once bot is restarted
