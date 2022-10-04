package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"airbot/config"
	"airbot/logs"
	"airbot/platforms"
)

// configFileName contains the name of the config file to be read.
const configFileName = "config.json"

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
			err := f.f()
			if err != nil {
				logs.Printf("cleanup function %s failed: %v", f.name, err)
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
	logs.Printf("Reading config from %s...", configFileName)
	cfg, err := config.Read(configFileName)
	if err != nil {
		logs.Fatalf("failed to read config from %s: %v\n", configFileName, err)
	}

	logs.Printf("Building platforms...")
	ps, err := platforms.Build(cfg)
	if err != nil {
		logs.Fatalf("failed to build platforms: %v\n", err)
	}

	for _, p := range ps {
		logs.Printf("Connecting to platform %s...", p.Name())
		p.Connect()

		logs.Printf("Starting to handle messages on platform %s...", p.Name())
		go platforms.StartHandling(p, cfg.LogIncoming, cfg.LogOutgoing)
		cleanupFuncs = append(cleanupFuncs, cleanupFunc{name: p.Name(), f: p.Disconnect})
	}

	startListeningForSigterm()

	logs.Printf("Airbot is now running.")
	wait()
}
