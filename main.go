// Package main contains the main logic for airbot.
package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/airforce270/airbot/config"
	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/platforms"
)

// cleanupFunc is a function that should be called before program exit.
type cleanupFunc struct {
	// name is the function's human-readable name.
	name string
	// f is the function to be called.
	f func() error
}

// cleanupFuncs contains functions to be called to cleanup before program exit.
var cleanupFuncs []cleanupFunc

// startListeningForSigterm starts a goroutine that listens for SIGTERM (ctrl-c)
// and runs the cleanup functions before exiting.
func startListeningForSigterm() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		for _, f := range cleanupFuncs {
			if err := f.f(); err != nil {
				log.Printf("cleanup function %s failed: %v", f.name, err)
			}
		}
		os.Exit(1)
	}()
}

// wait blocks the thread that calls it indefinitely.
func wait() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	log.Printf("Reading config from %s...", config.Path)
	cfg, err := config.Read(config.Path)
	if err != nil {
		log.Fatalf("failed to read config from %s: %v", config.Path, err)
	}
	config.Instance = cfg

	log.Printf("Connecting to database...")
	db, err := database.Connect(os.Getenv("POSTGRES_DB"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	database.Instance = db

	log.Printf("Performing database migrations...")
	if database.Migrate(db); err != nil {
		log.Fatalf("failed to perform database migrations: %v", err)
	}

	log.Printf("Preparing chat connections...")
	ps, err := platforms.Build(cfg, db)
	if err != nil {
		log.Fatalf("Failed to build platforms: %v", err)
	}

	for _, p := range ps {
		log.Printf("Connecting to %s...", p.Name())
		if err := p.Connect(); err != nil {
			log.Fatalf("Failed to connect to %s: %v", p.Name(), err)
		}

		log.Printf("Starting to handle messages on %s...", p.Name())
		go platforms.StartHandling(p, db, cfg.LogIncoming, cfg.LogOutgoing, cfg.EnableNonPrefixCommands)
		cleanupFuncs = append(cleanupFuncs, cleanupFunc{name: p.Name(), f: p.Disconnect})
	}

	startListeningForSigterm()

	log.Printf("Airbot is now running.")
	wait()
}
