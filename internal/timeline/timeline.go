// Package timeline provides timeline management for incidents.
package timeline

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/fishnix/ohshift/internal/incident"
	"github.com/fishnix/ohshift/internal/logger"
	"github.com/slack-go/slack"
)

// Entry represents a single entry in the timeline
type Entry struct {
	ID        string // Unique identifier to prevent duplicates
	Timestamp time.Time
	Type      string // "incident_start", "message", "image", "reaction", "bot_interaction"
	UserID    string // Slack user ID (e.g., "U0123456")
	Username  string // Slack username (e.g., "thatopsguy")
	Content   string
	Metadata  map[string]interface{}
}

// Timeline represents an incident timeline
type Timeline struct {
	IncidentID  string
	ChannelID   string
	LastUpdated time.Time
	Entries     []Entry
	mu          sync.RWMutex
}

// Manager handles timeline operations
type Manager struct {
	api       *slack.Client
	logger    *slog.Logger
	timelines map[string]*Timeline
	userCache map[string]string // userID -> username cache
	mu        sync.RWMutex
}

// NewManager creates a new timeline manager
func NewManager(api *slack.Client) *Manager {
	return &Manager{
		api:       api,
		logger:    logger.With("component", "timeline_manager"),
		timelines: make(map[string]*Timeline),
		userCache: make(map[string]string),
	}
}

// resolveUsername resolves a user ID to a username using the Slack API
func (m *Manager) resolveUsername(userID string) string {
	// Check cache first
	m.mu.RLock()
	if username, exists := m.userCache[userID]; exists {
		m.mu.RUnlock()
		return username
	}
	m.mu.RUnlock()

	// If not in cache, fetch from Slack API
	user, err := m.api.GetUserInfo(userID)
	if err != nil {
		m.logger.Warn("Failed to get user info, using user ID as fallback",
			"error", err,
			"user_id", userID)
		// Cache the user ID as fallback to avoid repeated API calls
		m.mu.Lock()
		m.userCache[userID] = userID
		m.mu.Unlock()
		return userID
	}

	// Cache the username
	m.mu.Lock()
	m.userCache[userID] = user.Name
	m.mu.Unlock()

	m.logger.Debug("Resolved user ID to username",
		"user_id", userID,
		"username", user.Name)

	return user.Name
}

// resolveUserInfo resolves a user ID to both user ID and username
func (m *Manager) resolveUserInfo(userID string) (string, string) {
	username := m.resolveUsername(userID)
	return userID, username
}

// CreateTimeline creates a new timeline for an incident
func (m *Manager) CreateTimeline(inc *incident.Incident, channelID string) (*Timeline, error) {
	m.logger.Info("Creating timeline for incident",
		"incident_id", inc.ID,
		"channel_id", channelID,
		"severity", inc.Severity,
		"title", inc.Title)

	// Resolve user ID to username for the incident starter
	userID, username := m.resolveUserInfo(inc.StartedBy)

	initialEntry := Entry{
		ID:        fmt.Sprintf("incident_start_%s", inc.ID),
		Timestamp: inc.StartedAt,
		Type:      "incident_start",
		UserID:    userID,
		Username:  username,
		Content:   fmt.Sprintf("🚨 %s Incident Started", inc.Severity),
		Metadata: map[string]interface{}{
			"severity":    inc.Severity,
			"title":       inc.Title,
			"description": inc.Description,
		},
	}

	timeline := &Timeline{
		IncidentID:  inc.ID,
		ChannelID:   channelID,
		LastUpdated: time.Now(),
		Entries:     []Entry{initialEntry},
	}

	m.mu.Lock()
	m.timelines[inc.ID] = timeline
	m.mu.Unlock()

	m.logger.Info("Timeline created in memory",
		"incident_id", inc.ID,
		"initial_entries", 1)

	// Don't post timeline message to channel initially since incident creation already displays the information
	m.logger.Info("Timeline created successfully (not posted to channel initially)",
		"incident_id", inc.ID,
		"channel_id", channelID)

	return timeline, nil
}

// GetTimeline retrieves a timeline by incident ID
func (m *Manager) GetTimeline(incidentID string) (*Timeline, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	timeline, exists := m.timelines[incidentID]

	if exists {
		m.logger.Debug("Timeline retrieved",
			"incident_id", incidentID,
			"entries_count", len(timeline.Entries),
			"last_updated", timeline.LastUpdated)
	} else {
		m.logger.Debug("Timeline not found",
			"incident_id", incidentID)
	}

	return timeline, exists
}

// AddEntry adds a new entry to the timeline
func (m *Manager) AddEntry(incidentID string, entry Entry) error {
	m.logger.Info("Adding entry to timeline",
		"incident_id", incidentID,
		"entry_type", entry.Type,
		"entry_id", entry.ID,
		"user", entry.Username,
		"timestamp", entry.Timestamp)

	m.mu.Lock()
	timeline, exists := m.timelines[incidentID]
	m.mu.Unlock()

	if !exists {
		m.logger.Error("Timeline not found for entry",
			"incident_id", incidentID,
			"entry_type", entry.Type,
			"entry_id", entry.ID,
			"user", entry.Username)
		return fmt.Errorf("timeline not found for incident: %s", incidentID)
	}

	// Check for duplicate entry
	timeline.mu.Lock()
	for _, existingEntry := range timeline.Entries {
		if existingEntry.ID == entry.ID {
			timeline.mu.Unlock()
			m.logger.Debug("Duplicate entry skipped",
				"incident_id", incidentID,
				"entry_id", entry.ID,
				"entry_type", entry.Type,
				"user", entry.Username)
			return nil // Skip duplicate entry
		}
	}

	timeline.Entries = append(timeline.Entries, entry)
	timeline.LastUpdated = time.Now()
	entriesCount := len(timeline.Entries)
	timeline.mu.Unlock()

	m.logger.Info("Entry added to timeline in memory",
		"incident_id", incidentID,
		"entry_type", entry.Type,
		"entry_id", entry.ID,
		"total_entries", entriesCount,
		"user", entry.Username)

	// Update timeline in channel
	err := m.postTimelineToChannel(timeline)
	if err != nil {
		m.logger.Error("Failed to update timeline in channel",
			"error", err,
			"incident_id", incidentID,
			"entry_type", entry.Type)

		return err
	}

	// Add white check mark reaction for entries that correspond to Slack messages
	if entry.Type == "message" || entry.Type == "image" || entry.Type == "highlighted" {
		if messageID, ok := entry.Metadata["message_id"].(string); ok {
			if err := m.addReactionToMessage(timeline.ChannelID, messageID, "white_check_mark"); err != nil {
				m.logger.Warn("Failed to add reaction to message (non-critical)",
					"error", err,
					"incident_id", incidentID,
					"entry_type", entry.Type,
					"entry_id", entry.ID,
					"message_id", messageID)
				// Don't return error here as the timeline entry was successfully added
			} else {
				m.logger.Debug("White check mark reaction added to message",
					"incident_id", incidentID,
					"entry_type", entry.Type,
					"entry_id", entry.ID,
					"message_id", messageID)
			}
		}
	}

	m.logger.Info("Timeline updated successfully",
		"incident_id", incidentID,
		"entry_type", entry.Type,
		"entry_id", entry.ID,
		"total_entries", entriesCount)

	return nil
}

// AddMessageEntry adds a message to the timeline
func (m *Manager) AddMessageEntry(incidentID, userID, message, messageID string) error {
	m.logger.Debug("Adding message entry to timeline",
		"incident_id", incidentID,
		"user_id", userID,
		"message_id", messageID,
		"message_length", len(message))

	// Resolve user ID to username
	resolvedUserID, username := m.resolveUserInfo(userID)

	entry := Entry{
		ID:        fmt.Sprintf("message_%s", messageID),
		Timestamp: time.Now(),
		Type:      "message",
		UserID:    resolvedUserID,
		Username:  username,
		Content:   message,
		Metadata: map[string]interface{}{
			"message_id": messageID,
		},
	}

	return m.AddEntry(incidentID, entry)
}

// AddImageEntry adds an image to the timeline
func (m *Manager) AddImageEntry(incidentID, userID, imageURL, caption, messageID string) error {
	m.logger.Debug("Adding image entry to timeline",
		"incident_id", incidentID,
		"user_id", userID,
		"message_id", messageID,
		"image_url", imageURL,
		"caption", caption)

	// Resolve user ID to username
	resolvedUserID, username := m.resolveUserInfo(userID)

	entry := Entry{
		ID:        fmt.Sprintf("image_%s", messageID),
		Timestamp: time.Now(),
		Type:      "image",
		UserID:    resolvedUserID,
		Username:  username,
		Content:   caption,
		Metadata: map[string]interface{}{
			"image_url":  imageURL,
			"message_id": messageID,
		},
	}

	return m.AddEntry(incidentID, entry)
}

// AddReactionEntry adds a reaction to the timeline
func (m *Manager) AddReactionEntry(incidentID, userID, message, reaction, messageID string) error {
	m.logger.Debug("Adding reaction entry to timeline",
		"incident_id", incidentID,
		"user_id", userID,
		"message_id", messageID,
		"reaction", reaction,
		"message_length", len(message))

	// Resolve user ID to username
	resolvedUserID, username := m.resolveUserInfo(userID)

	entry := Entry{
		ID:        fmt.Sprintf("reaction_%s_%s", messageID, reaction),
		Timestamp: time.Now(),
		Type:      "reaction",
		UserID:    resolvedUserID,
		Username:  username,
		Content:   message,
		Metadata: map[string]interface{}{
			"reaction":   reaction,
			"message_id": messageID,
		},
	}

	return m.AddEntry(incidentID, entry)
}

// AddBotInteractionEntry adds a bot interaction to the timeline
func (m *Manager) AddBotInteractionEntry(incidentID, userID, interaction string) error {
	m.logger.Debug("Adding bot interaction entry to timeline",
		"incident_id", incidentID,
		"user_id", userID,
		"interaction", interaction)

	// Resolve user ID to username
	resolvedUserID, username := m.resolveUserInfo(userID)

	entry := Entry{
		ID:        fmt.Sprintf("bot_interaction_%s_%d", userID, time.Now().UnixNano()),
		Timestamp: time.Now(),
		Type:      "bot_interaction",
		UserID:    resolvedUserID,
		Username:  username,
		Content:   interaction,
		Metadata:  map[string]interface{}{},
	}

	return m.AddEntry(incidentID, entry)
}

// AddHighlightedEntry adds a highlighted message to the timeline (e.g., for :point_up: reactions)
func (m *Manager) AddHighlightedEntry(incidentID, userID, message, messageID string, originalTimestamp time.Time) error {
	m.logger.Debug("Adding highlighted entry to timeline",
		"incident_id", incidentID,
		"user_id", userID,
		"message_id", messageID,
		"message_length", len(message),
		"original_timestamp", originalTimestamp)

	// Resolve user ID to username
	resolvedUserID, username := m.resolveUserInfo(userID)

	entry := Entry{
		ID:        fmt.Sprintf("highlighted_%s", messageID),
		Timestamp: originalTimestamp,
		Type:      "highlighted",
		UserID:    resolvedUserID,
		Username:  username,
		Content:   message,
		Metadata: map[string]interface{}{
			"message_id": messageID,
		},
	}

	return m.AddEntry(incidentID, entry)
}

// postTimelineToChannel posts the timeline to the incident channel
func (m *Manager) postTimelineToChannel(timeline *Timeline) error {
	m.logger.Debug("Posting timeline to channel",
		"incident_id", timeline.IncidentID,
		"channel_id", timeline.ChannelID)

	timeline.mu.RLock()
	entries := make([]Entry, len(timeline.Entries))
	copy(entries, timeline.Entries)
	entriesCount := len(entries)

	timeline.mu.RUnlock()

	// Create timeline message
	message := m.formatTimelineMessage(entries)
	messageLength := len(message)

	m.logger.Debug("Timeline message formatted",
		"incident_id", timeline.IncidentID,
		"entries_count", entriesCount,
		"message_length", messageLength)

	// Post to channel
	_, _, err := m.api.PostMessage(timeline.ChannelID, slack.MsgOptionText(message, false))
	if err != nil {
		m.logger.Error("Failed to post timeline message to channel",
			"error", err,
			"incident_id", timeline.IncidentID,
			"channel_id", timeline.ChannelID,
			"message_length", messageLength)

		return err
	}

	m.logger.Debug("Timeline message posted successfully",
		"incident_id", timeline.IncidentID,
		"channel_id", timeline.ChannelID,
		"entries_count", entriesCount)

	return nil
}

// formatTimelineMessage formats the timeline entries into a readable message
func (m *Manager) formatTimelineMessage(entries []Entry) string {
	if len(entries) == 0 {
		m.logger.Debug("Formatting empty timeline message")
		return "📋 *Timeline*\nNo entries yet."
	}

	m.logger.Debug("Formatting timeline message",
		"entries_count", len(entries))

	message := "📋 *Incident Timeline*\n\n"

	for i, entry := range entries {
		timestamp := entry.Timestamp.Format("15:04:05")
		icon := m.getEntryIcon(entry.Type)

		// Ensure we have a username to display (backward compatibility)
		displayName := m.ensureUsername(entry)

		message += fmt.Sprintf("%s *%s* - @%s\n", icon, timestamp, displayName)
		message += fmt.Sprintf("   %s\n", entry.Content)

		// Add metadata if present
		if len(entry.Metadata) > 0 {
			for key, value := range entry.Metadata {
				message += fmt.Sprintf("   • %s: %v\n", key, value)
			}
		}

		if i < len(entries)-1 {
			message += "\n"
		}
	}

	m.logger.Debug("Timeline message formatted successfully",
		"entries_count", len(entries),
		"message_length", len(message))

	return message
}

// ensureUsername ensures we have a username to display, handling backward compatibility
func (m *Manager) ensureUsername(entry Entry) string {
	// If we already have a username, use it
	if entry.Username != "" {
		return entry.Username
	}

	// If we have a user ID, resolve it to a username
	if entry.UserID != "" {
		return m.resolveUsername(entry.UserID)
	}

	// Fallback to a generic name if neither is available
	return "unknown_user"
}

// getEntryIcon returns the appropriate icon for an entry type
func (m *Manager) getEntryIcon(entryType string) string {
	switch entryType {
	case "incident_start":
		return "🚨"
	case "message":
		return "💬"
	case "image":
		return "🖼️"
	case "reaction":
		return "👆"
	case "bot_interaction":
		return "🤖"
	default:
		return "📝"
	}
}

// ExportTimeline exports the timeline as JSON
func (m *Manager) ExportTimeline(incidentID string) ([]byte, error) {
	m.logger.Info("Exporting timeline",
		"incident_id", incidentID)

	timeline, exists := m.GetTimeline(incidentID)
	if !exists {
		m.logger.Error("Timeline not found for export",
			"incident_id", incidentID)
		return nil, fmt.Errorf("timeline not found for incident: %s", incidentID)
	}

	timeline.mu.RLock()
	defer timeline.mu.RUnlock()

	data, err := json.MarshalIndent(timeline, "", "  ")
	if err != nil {
		m.logger.Error("Failed to marshal timeline for export",
			"error", err,
			"incident_id", incidentID)
		return nil, err
	}

	m.logger.Info("Timeline exported successfully",
		"incident_id", incidentID,
		"data_size", len(data),
		"entries_count", len(timeline.Entries))

	return data, nil
}

// GetEntries returns a copy of the timeline entries
func (t *Timeline) GetEntries() []Entry {
	t.mu.RLock()
	defer t.mu.RUnlock()

	entries := make([]Entry, len(t.Entries))
	copy(entries, t.Entries)

	return entries
}

// GetLastUpdated returns the last updated timestamp
func (t *Timeline) GetLastUpdated() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.LastUpdated
}

// HasEntry checks if a timeline already has an entry with the given ID
func (m *Manager) HasEntry(incidentID, entryID string) bool {
	timeline, exists := m.GetTimeline(incidentID)
	if !exists {
		return false
	}

	timeline.mu.RLock()
	defer timeline.mu.RUnlock()

	for _, entry := range timeline.Entries {
		if entry.ID == entryID {
			return true
		}
	}
	return false
}

// addReactionToMessage adds a reaction to a message in Slack
func (m *Manager) addReactionToMessage(channelID, messageTimestamp, reaction string) error {
	m.logger.Debug("Adding reaction to message",
		"channel_id", channelID,
		"message_timestamp", messageTimestamp,
		"reaction", reaction)

	err := m.api.AddReaction(reaction, slack.ItemRef{
		Channel:   channelID,
		Timestamp: messageTimestamp,
	})
	if err != nil {
		m.logger.Error("Failed to add reaction to message",
			"error", err,
			"channel_id", channelID,
			"message_timestamp", messageTimestamp,
			"reaction", reaction)
		return err
	}

	m.logger.Debug("Reaction added successfully",
		"channel_id", channelID,
		"message_timestamp", messageTimestamp,
		"reaction", reaction)
	return nil
}
