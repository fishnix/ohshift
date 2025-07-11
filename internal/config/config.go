// Package config provides configuration management for the OhShift! bot.
package config

import (
	"log/slog"
	"os"
	"strings"
)

// Config holds all configuration for the OhShift bot
type Config struct {
	DBURI                    string
	SlackBotToken            string
	SlackSigningSecret       string
	SlackAppToken            string
	SlashCommand             string
	NotificationsChannel     string
	Port                     string
	LogLevel                 slog.Level
	AddAllMessagesToTimeline bool
}

// Load loads configuration from environment variables
func Load() *Config {
	config := &Config{
		DBURI:                    getEnv("DB_URI", ""),
		SlackBotToken:            getEnv("SLACK_BOT_TOKEN", ""),
		SlackSigningSecret:       getEnv("SLACK_SIGNING_SECRET", ""),
		SlackAppToken:            getEnv("SLACK_APP_TOKEN", ""),
		SlashCommand:             getEnv("SLASH_COMMAND", "/shift"),
		NotificationsChannel:     getEnv("NOTIFICATIONS_CHANNEL", "general"),
		Port:                     getEnv("PORT", "8080"),
		LogLevel:                 parseLogLevel(getEnv("LOG_LEVEL", "info")),
		AddAllMessagesToTimeline: getEnvBool("ADD_ALL_MESSAGES_TO_TIMELINE", false),
	}

	return config
}

// parseLogLevel parses a log level string into slog.Level
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Validate checks if required configuration is present
func (c *Config) Validate() error {
	if c.SlackBotToken == "" {
		return &Error{Field: "SLACK_BOT_TOKEN", Message: "Slack bot token is required"}
	}

	if c.SlackSigningSecret == "" {
		return &Error{Field: "SLACK_SIGNING_SECRET", Message: "Slack signing secret is required"}
	}

	if c.SlackAppToken == "" {
		return &Error{Field: "SLACK_APP_TOKEN", Message: "Slack app token is required for Socket Mode"}
	}

	return nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool gets an environment variable and parses it as a boolean
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if strings.ToLower(value) == "true" {
			return true
		}

		if strings.ToLower(value) == "false" {
			return false
		}
	}

	return defaultValue
}

// Error represents a configuration error
type Error struct {
	Field   string
	Message string
}

func (e *Error) Error() string {
	return e.Message
}
