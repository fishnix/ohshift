// Package main is the entry point for the OhShift! Slack incident management bot.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/fishnix/ohshift/config"
	"github.com/fishnix/ohshift/logger"
	"github.com/fishnix/ohshift/slack"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set log level from configuration
	logger.SetLevel(cfg.LogLevel)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		logger.Fatal("Configuration error", "error", err)
	}

	// Create Slack bot
	bot := slack.NewBot(cfg)

	// Set up context with cancel on SIGINT/SIGTERM
	ctx, cancel := context.WithCancel(context.Background())
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("Received shutdown signal")
		cancel()
	}()

	// Start the bot (blocks until shutdown)
	if err := bot.Start(ctx); err != nil {
		logger.Fatal("Bot error", "error", err)
	}

	logger.Info("Bot exited cleanly")
}
