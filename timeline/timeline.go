// Package timeline provides timeline management for incidents.
package timeline

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/fishnix/ohshift/incident"
	"github.com/fishnix/ohshift/logger"
	"github.com/slack-go/slack"
)

// Entry represents a single entry in the timeline
type Entry struct {
	Timestamp time.Time
	Type      string // "incident_start", "message", "image", "reaction", "bot_interaction"
	User      string
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
	mu        sync.RWMutex
}

// NewManager creates a new timeline manager
func NewManager(api *slack.Client) *Manager {
	return &Manager{
		api:       api,
		logger:    logger.With("component", "timeline_manager"),
		timelines: make(map[string]*Timeline),
	}
}

// CreateTimeline creates a new timeline for an incident
func (m *Manager) CreateTimeline(inc *incident.Incident, channelID string) (*Timeline, error) {
	m.logger.Info("Creating timeline for incident",
		"incident_id", inc.ID,
		"channel_id", channelID,
		"severity", inc.Severity,
		"title", inc.Title)

	initialEntry := Entry{
		Timestamp: inc.StartedAt,
		Type:      "incident_start",
		User:      inc.StartedBy,
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
		"user", entry.User,
		"timestamp", entry.Timestamp)

	m.mu.Lock()
	timeline, exists := m.timelines[incidentID]
	m.mu.Unlock()

	if !exists {
		m.logger.Error("Timeline not found for entry",
			"incident_id", incidentID,
			"entry_type", entry.Type,
			"user", entry.User)

		return fmt.Errorf("timeline not found for incident: %s", incidentID)
	}

	timeline.mu.Lock()
	timeline.Entries = append(timeline.Entries, entry)
	timeline.LastUpdated = time.Now()
	entriesCount := len(timeline.Entries)
	timeline.mu.Unlock()

	m.logger.Info("Entry added to timeline in memory",
		"incident_id", incidentID,
		"entry_type", entry.Type,
		"total_entries", entriesCount,
		"user", entry.User)

	// Update timeline in channel
	err := m.postTimelineToChannel(timeline)
	if err != nil {
		m.logger.Error("Failed to update timeline in channel",
			"error", err,
			"incident_id", incidentID,
			"entry_type", entry.Type)

		return err
	}

	m.logger.Info("Timeline updated in channel successfully",
		"incident_id", incidentID,
		"entry_type", entry.Type,
		"total_entries", entriesCount)

	return nil
}

// AddMessageEntry adds a message to the timeline
func (m *Manager) AddMessageEntry(incidentID, user, message string) error {
	m.logger.Debug("Adding message entry to timeline",
		"incident_id", incidentID,
		"user", user,
		"message_length", len(message))

	entry := Entry{
		Timestamp: time.Now(),
		Type:      "message",
		User:      user,
		Content:   message,
		Metadata:  map[string]interface{}{},
	}

	return m.AddEntry(incidentID, entry)
}

// AddImageEntry adds an image to the timeline
func (m *Manager) AddImageEntry(incidentID, user, imageURL, caption string) error {
	m.logger.Debug("Adding image entry to timeline",
		"incident_id", incidentID,
		"user", user,
		"image_url", imageURL,
		"caption", caption)

	entry := Entry{
		Timestamp: time.Now(),
		Type:      "image",
		User:      user,
		Content:   caption,
		Metadata: map[string]interface{}{
			"image_url": imageURL,
		},
	}

	return m.AddEntry(incidentID, entry)
}

// AddReactionEntry adds a reaction to the timeline
func (m *Manager) AddReactionEntry(incidentID, user, message, reaction string) error {
	m.logger.Debug("Adding reaction entry to timeline",
		"incident_id", incidentID,
		"user", user,
		"reaction", reaction,
		"message_length", len(message))

	entry := Entry{
		Timestamp: time.Now(),
		Type:      "reaction",
		User:      user,
		Content:   message,
		Metadata: map[string]interface{}{
			"reaction": reaction,
		},
	}

	return m.AddEntry(incidentID, entry)
}

// AddBotInteractionEntry adds a bot interaction to the timeline
func (m *Manager) AddBotInteractionEntry(incidentID, user, interaction string) error {
	m.logger.Debug("Adding bot interaction entry to timeline",
		"incident_id", incidentID,
		"user", user,
		"interaction", interaction)

	entry := Entry{
		Timestamp: time.Now(),
		Type:      "bot_interaction",
		User:      user,
		Content:   interaction,
		Metadata:  map[string]interface{}{},
	}

	return m.AddEntry(incidentID, entry)
}

// AddHighlightedEntry adds a highlighted message to the timeline (e.g., for :point_up: reactions)
func (m *Manager) AddHighlightedEntry(incidentID, user, message string, originalTimestamp time.Time) error {
	m.logger.Debug("Adding highlighted entry to timeline",
		"incident_id", incidentID,
		"user", user,
		"message_length", len(message),
		"original_timestamp", originalTimestamp)

	entry := Entry{
		Timestamp: originalTimestamp,
		Type:      "highlighted",
		User:      user,
		Content:   message,
		Metadata:  map[string]interface{}{},
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

		message += fmt.Sprintf("%s *%s* - @%s\n", icon, timestamp, entry.User)
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
