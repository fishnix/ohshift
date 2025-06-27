// Package incident provides incident management types and utilities for the OhShift! bot.
package incident

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Severity represents the incident severity level
type Severity string

const (
	// Severity0 represents a major customer impact incident (highest priority)
	Severity0 Severity = "SEV0"
	// Severity1 represents a high customer impact incident
	Severity1 Severity = "SEV1"
	// Severity2 represents a low/no customer impact incident
	Severity2 Severity = "SEV2"
	// Severity3 represents a maintenance incident (lowest priority)
	Severity3 Severity = "SEV3"
)

// Incident represents an incident
type Incident struct {
	ID          string
	Title       string
	Description string
	Severity    Severity
	ChannelName string
	StartedBy   string
	StartedAt   time.Time
}

// Command represents a parsed slash command
type Command struct {
	Action      string
	Severity    Severity
	Title       string
	Description string
	Username    string
	UserID      string
}

// ParseCommand parses a slash command string into a Command
func ParseCommand(text string) (*Command, error) {
	parts := strings.Fields(text)
	if len(parts) < 4 {
		return nil, fmt.Errorf("insufficient arguments")
	}

	if parts[0] != "start" {
		return nil, fmt.Errorf("unknown action: %s", parts[0])
	}

	severity := Severity(strings.ToUpper(parts[1]))
	if !isValidSeverity(severity) {
		return nil, fmt.Errorf("invalid severity: %s", parts[1])
	}

	if parts[2] != "incident" {
		return nil, fmt.Errorf("expected 'incident' keyword, got: %s", parts[2])
	}

	// Find the description separator
	textAfterIncident := strings.Join(parts[3:], " ")
	title := textAfterIncident
	description := ""

	// Check if there's a description separator (--)
	if strings.Contains(textAfterIncident, " -- ") {
		parts := strings.SplitN(textAfterIncident, " -- ", 2)
		title = strings.TrimSpace(parts[0])
		description = strings.TrimSpace(parts[1])
	}

	if title == "" {
		return nil, fmt.Errorf("incident title cannot be empty")
	}

	return &Command{
		Action:      "start",
		Severity:    severity,
		Title:       title,
		Description: description,
	}, nil
}

// GenerateChannelName generates a Slack-compatible channel name for an incident
func GenerateChannelName(incident *Incident) string {
	// Format: _inc-YYYYMMDD-HHMMSS-title
	timestamp := incident.StartedAt.Format("20060102-150405")

	// Create slug from title
	slug := createSlug(incident.Title)

	// Construct channel name
	channelName := fmt.Sprintf("_inc-%s-%s", timestamp, slug)

	// Truncate to 64 characters (Slack limit)
	if len(channelName) > 64 {
		channelName = channelName[:64]
		// Ensure we don't cut in the middle of a word if possible
		if lastDash := strings.LastIndex(channelName, "-"); lastDash > 50 {
			channelName = channelName[:lastDash]
		}
	}

	return channelName
}

// createSlug converts a string to a URL-friendly slug
func createSlug(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")

	// Remove leading and trailing hyphens
	s = strings.Trim(s, "-")

	// Limit length to prevent overly long slugs
	if len(s) > 30 {
		s = s[:30]
		// Ensure we don't cut in the middle of a word
		if lastDash := strings.LastIndex(s, "-"); lastDash > 20 {
			s = s[:lastDash]
		}
	}

	return s
}

// isValidSeverity checks if a severity is valid
func isValidSeverity(s Severity) bool {
	validSeverities := []Severity{Severity0, Severity1, Severity2, Severity3}
	for _, valid := range validSeverities {
		if s == valid {
			return true
		}
	}

	return false
}

// GetHelpMessage returns the help message for the slash command
func GetHelpMessage() string {
	return `Usage: /shift start <severity> incident <incident title> [-- <description>]

Examples:
  /shift start SEV0 incident the website is down
  /shift start SEV1 incident database connection issues -- Connection pool exhausted, affecting all users
  /shift start SEV2 incident slow response times -- API response times > 5s, investigating root cause

Valid severities:
  SEV0: Major Customer Impact
  SEV1: High Customer Impact
  SEV2: Low/No Customer Impact
  SEV3: Maintenance

This will create an incident channel and post a notification.`
}

// GenerateIncidentID generates a unique incident ID
func GenerateIncidentID() string {
	// Generate a random 8-byte ID
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return fmt.Sprintf("inc_%d", time.Now().UnixNano())
	}

	return fmt.Sprintf("inc_%x", bytes)
}
