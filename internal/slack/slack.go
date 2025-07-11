// Package slack provides Slack API integration using Socket Mode for the OhShift! bot.
package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fishnix/ohshift/internal/config"
	"github.com/fishnix/ohshift/internal/incident"
	"github.com/fishnix/ohshift/internal/logger"
	"github.com/fishnix/ohshift/internal/timeline"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

const (
	timelineCommand = "timeline"
	// MsgCommandNotInIncidentChannel is the error message shown when timeline command is used outside incident channels
	MsgCommandNotInIncidentChannel = "❌ This command can only be used in incident channels."
	// MsgTimelineNotFound is the error message shown when timeline is not found for an incident
	MsgTimelineNotFound = "❌ Timeline not found for this incident."
)

// Bot represents the Slack bot
type Bot struct {
	api          *slack.Client
	socketClient *socketmode.Client
	handler      *socketmode.SocketmodeHandler
	config       *config.Config
	logger       *slog.Logger
	timelineMgr  *timeline.Manager
	// TODO: Replace this mapping of channel IDs to
	// incident IDs with something more durable
	channelToIncident map[string]string
	mu                sync.RWMutex
}

// NewBot creates a new Slack bot instance with Socket Mode
func NewBot(cfg *config.Config) *Bot {
	api := slack.New(cfg.SlackBotToken, slack.OptionAppLevelToken(cfg.SlackAppToken))
	socketClient := socketmode.New(api)
	handler := socketmode.NewSocketmodeHandler(socketClient)

	return &Bot{
		api:               api,
		socketClient:      socketClient,
		handler:           handler,
		config:            cfg,
		logger:            logger.With("component", "slack_bot"),
		timelineMgr:       timeline.NewManager(api),
		channelToIncident: make(map[string]string),
	}
}

// Start starts the Socket Mode event loop and blocks until ctx is done
func (b *Bot) Start(ctx context.Context) error {
	b.setupEventHandlers()

	// Run the event loop in a goroutine
	errCh := make(chan error, 1)

	go func() {
		errCh <- b.handler.RunEventLoopContext(ctx)
	}()

	b.logger.Info("Slack bot started with Socket Mode",
		"slash_command", b.config.SlashCommand,
		"notifications_channel", b.config.NotificationsChannel)

	select {
	case <-ctx.Done():
		b.logger.Info("Shutting down Slack bot")
		return nil
	case err := <-errCh:
		return err
	}
}

// setupEventHandlers sets up all the event handlers for the bot
func (b *Bot) setupEventHandlers() {
	b.handler.Handle(socketmode.EventTypeSlashCommand, b.handleSlashCommand)
	b.handler.Handle(socketmode.EventTypeEventsAPI, b.handleEventsAPI)
}

// handleSlashCommand handles incoming slash commands via Socket Mode
func (b *Bot) handleSlashCommand(evt *socketmode.Event, client *socketmode.Client) {
	cmd, ok := evt.Data.(slack.SlashCommand)
	if !ok {
		b.logger.Debug("Failed to parse slash command event", "event_type", evt.Type)
		return
	}

	b.logger.Info("Received slash command",
		"command", cmd.Command,
		"user", cmd.UserName,
		"text", cmd.Text,
		"channel_id", cmd.ChannelID)

	// Check if this is the timeline command
	if cmd.Text == timelineCommand {
		b.handleTimelineCommand(cmd, client, evt)
		return
	}

	// Parse the command for incident creation
	incidentCmd, err := incident.ParseCommand(cmd.Text)
	if err != nil {
		// Send help message
		response := &slack.Msg{
			ResponseType: "ephemeral",
			Text:         fmt.Sprintf("Error: %s\n\n%s", err.Error(), incident.GetHelpMessage()),
		}
		b.sendSlashResponse(client, evt, response)

		return
	}

	// Set user information
	incidentCmd.UserID = cmd.UserID
	incidentCmd.Username = cmd.UserName

	// Create the incident
	if err := b.createIncident(incidentCmd); err != nil {
		b.logger.Error("Failed to create incident", "error", err, "user", cmd.UserName)

		response := &slack.Msg{
			ResponseType: "ephemeral",
			Text:         fmt.Sprintf("Failed to create incident: %v", err),
		}

		b.sendSlashResponse(client, evt, response)

		return
	}

	// Send success response
	response := &slack.Msg{
		ResponseType: "ephemeral",
		Text:         "Incident created successfully! Check the notifications channel for details.",
	}

	b.sendSlashResponse(client, evt, response)
}

// handleTimelineCommand handles the /shift timeline command
func (b *Bot) handleTimelineCommand(cmd slack.SlashCommand, client *socketmode.Client, evt *socketmode.Event) {
	b.logger.Info("Processing timeline command",
		"user", cmd.UserName,
		"channel_id", cmd.ChannelID)

	// Check if this is an incident channel
	incidentID := b.findIncidentIDByChannel(cmd.ChannelID)
	if incidentID == "" {
		b.logger.Warn("Timeline command used in non-incident channel",
			"user", cmd.UserName,
			"channel_id", cmd.ChannelID)

		response := &slack.Msg{
			ResponseType: "ephemeral",
			Text:         MsgCommandNotInIncidentChannel,
		}
		b.sendSlashResponse(client, evt, response)

		return
	}

	// Get the timeline
	timeline, exists := b.timelineMgr.GetTimeline(incidentID)
	if !exists {
		b.logger.Warn("Timeline not found for incident",
			"incident_id", incidentID,
			"user", cmd.UserName,
			"channel_id", cmd.ChannelID)

		response := &slack.Msg{
			ResponseType: "ephemeral",
			Text:         MsgTimelineNotFound,
		}

		b.sendSlashResponse(client, evt, response)

		return
	}

	// Format the timeline nicely
	timelineText := b.formatTimelineForDisplay(timeline)

	response := &slack.Msg{
		ResponseType: "ephemeral",
		Text:         timelineText,
	}

	b.sendSlashResponse(client, evt, response)

	b.logger.Info("Timeline displayed successfully",
		"incident_id", incidentID,
		"user", cmd.UserName,
		"channel_id", cmd.ChannelID,
		"entries_count", len(timeline.Entries))
}

// formatTimelineForDisplay formats the timeline for nice display in Slack
func (b *Bot) formatTimelineForDisplay(timeline *timeline.Timeline) string {
	entries := timeline.GetEntries()
	lastUpdated := timeline.GetLastUpdated()

	if len(entries) == 0 {
		return "📋 *Incident Timeline*\n\nNo entries yet."
	}

	// Get incident details from the first entry (incident_start)
	var (
		incidentTitle, incidentSeverity, incidentDescription, incidentStartedBy string
		incidentTime                                                            time.Time
	)

	if len(entries) > 0 && entries[0].Type == "incident_start" {
		if title, ok := entries[0].Metadata["title"].(string); ok {
			incidentTitle = title
		}

		if severity, ok := entries[0].Metadata["severity"].(string); ok {
			incidentSeverity = severity
		}

		if desc, ok := entries[0].Metadata["description"].(string); ok {
			incidentDescription = desc
		}

		incidentStartedBy = entries[0].Username
		incidentTime = entries[0].Timestamp
	}

	message := "📋 *Incident Timeline*\n\n"

	// Always show incident creation block
	message += ":new: *Incident Created*\n"
	message += fmt.Sprintf("*Date/Time:* %s\n", incidentTime.Format("2006-01-02 15:04:05"))
	message += fmt.Sprintf("*Severity:* %s\n", incidentSeverity)
	message += fmt.Sprintf("*Created by:* @%s\n", incidentStartedBy)
	message += fmt.Sprintf("*Title:* %s\n", incidentTitle)
	message += fmt.Sprintf("*Description:* %s\n", incidentDescription)
	message += "\n"

	// Add timeline entries
	for i, entry := range entries {
		// Skip the incident_start entry since we already showed it in the header
		if entry.Type == "incident_start" {
			continue
		}

		timestamp := entry.Timestamp.Format("15:04:05")
		icon := b.getTimelineEntryIcon(entry.Type)

		message += fmt.Sprintf("%s *%s* - @%s\n", icon, timestamp, entry.Username)

		// Format content based on entry type
		switch entry.Type {
		case "message":
			message += fmt.Sprintf("   %s\n", entry.Content)
		case "image":
			if imageURL, ok := entry.Metadata["image_url"].(string); ok {
				message += fmt.Sprintf("   📷 %s\n", entry.Content)
				message += fmt.Sprintf("   <%s|View Image>\n", imageURL)
			}
		case "reaction":
			if reaction, ok := entry.Metadata["reaction"].(string); ok {
				message += fmt.Sprintf("   Reacted with :%s: to:\n", reaction)
				message += fmt.Sprintf("   > %s\n", entry.Content)
			}
		case "bot_interaction":
			message += fmt.Sprintf("   🤖 %s\n", entry.Content)
		default:
			message += fmt.Sprintf("   %s\n", entry.Content)
		}

		if i < len(entries)-1 {
			message += "\n"
		}
	}

	// Add footer
	message += fmt.Sprintf("\n---\n*Timeline last updated: %s*", lastUpdated.Format("2006-01-02 15:04:05"))

	return message
}

// getTimelineEntryIcon returns the appropriate icon for a timeline entry type
func (b *Bot) getTimelineEntryIcon(entryType string) string {
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

// sendSlashResponse sends a response to a slash command via Socket Mode
func (b *Bot) sendSlashResponse(client *socketmode.Client, evt *socketmode.Event, response *slack.Msg) {
	payload := map[string]interface{}{
		"response_type": response.ResponseType,
		"text":          response.Text,
	}
	client.Ack(*evt.Request, payload)
}

// createIncident creates a new incident
func (b *Bot) createIncident(cmd *incident.Command) error {
	// Create incident object
	inc := &incident.Incident{
		ID:          incident.GenerateIncidentID(),
		Title:       cmd.Title,
		Description: cmd.Description,
		Severity:    cmd.Severity,
		StartedBy:   cmd.Username,
		StartedAt:   time.Now(),
	}

	// Generate channel name
	channelName := incident.GenerateChannelName(inc)
	inc.ChannelName = channelName

	// Create the channel with description
	channel, err := b.api.CreateConversation(slack.CreateConversationParams{
		ChannelName: channelName,
		IsPrivate:   false,
	})
	if err != nil {
		return fmt.Errorf("failed to create channel: %v", err)
	}

	// Set the channel topic and purpose after creation
	_, err = b.api.SetTopicOfConversation(channel.ID, fmt.Sprintf("%s Incident: %s", cmd.Severity, cmd.Title))
	if err != nil {
		b.logger.Warn("Failed to set channel topic", "error", err, "channel_id", channel.ID)
	}

	// Use description if provided, otherwise use title
	channelPurpose := cmd.Description
	if channelPurpose == "" {
		channelPurpose = cmd.Title
	}

	_, err = b.api.SetPurposeOfConversation(channel.ID, channelPurpose)
	if err != nil {
		b.logger.Warn("Failed to set channel purpose", "error", err, "channel_id", channel.ID)
	}

	// Invite the user who created the incident to the channel
	_, err = b.api.InviteUsersToConversation(channel.ID, cmd.UserID)
	if err != nil {
		b.logger.Warn("Failed to invite user to incident channel",
			"error", err,
			"user_id", cmd.UserID,
			"channel_id", channel.ID)
	} else {
		b.logger.Info("User invited to incident channel",
			"user_id", cmd.UserID,
			"username", cmd.Username,
			"channel_id", channel.ID)
	}

	// Create timeline for the incident
	_, err = b.timelineMgr.CreateTimeline(inc, channel.ID)
	if err != nil {
		b.logger.Warn("Failed to create timeline", "error", err, "incident_id", inc.ID)
	}

	// Post initial message in the incident channel
	descriptionText := cmd.Description
	if descriptionText == "" {
		descriptionText = cmd.Title
	}

	initialMessage := fmt.Sprintf("🚨 *%s Incident Started*\n\n"+
		"*Severity:* %s\n"+
		"*Started by:* <@%s>\n"+
		"*Title:* %s\n"+
		"*Description:* %s\n"+
		"*Started at:* %s\n\n"+
		"Please provide updates and coordinate the response in this channel.",
		cmd.Severity, cmd.Severity, cmd.UserID, cmd.Title, descriptionText, inc.StartedAt.Format("2006-01-02 15:04:05"))

	_, _, err = b.api.PostMessage(channel.ID, slack.MsgOptionText(initialMessage, false))
	if err != nil {
		b.logger.Error("Failed to post initial message", "error", err, "channel_id", channel.ID)
	}

	// Post notification in the notifications channel
	var notificationMessage string
	if cmd.Description != "" {
		notificationMessage = fmt.Sprintf("🚨 <@%s> started an incident: *%s*: <#%s>\n*Title:* %s\n*Description:* %s",
			cmd.UserID, cmd.Severity, channel.ID, cmd.Title, cmd.Description)
	} else {
		notificationMessage = fmt.Sprintf("🚨 <@%s> started an incident: *%s*: <#%s>\n*Title:* %s",
			cmd.UserID, cmd.Severity, channel.ID, cmd.Title)
	}

	_, _, err = b.api.PostMessage(b.config.NotificationsChannel, slack.MsgOptionText(notificationMessage, false))
	if err != nil {
		return fmt.Errorf("failed to post notification: %v", err)
	}

	b.logger.Info("Incident created successfully",
		"incident_id", inc.ID,
		"title", cmd.Title,
		"description", cmd.Description,
		"severity", cmd.Severity,
		"user", cmd.Username,
		"channel_id", channel.ID,
		"channel_name", channelName)

	// Store the mapping
	b.mu.Lock()
	b.channelToIncident[channel.ID] = inc.ID
	b.mu.Unlock()

	return nil
}

// HealthCheck handles health check requests (kept for compatibility)
func (b *Bot) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"bot":    "oh-shift",
		"mode":   "socket",
	}); err != nil {
		b.logger.Error("Failed to encode health check response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleEventsAPI handles all Events API events in Socket Mode
func (b *Bot) handleEventsAPI(evt *socketmode.Event, client *socketmode.Client) {
	// Debug log for the full event body - show actual data, not just pointer
	b.logger.Debug("Raw Events API event received",
		"event_type", evt.Type,
		"data_type", fmt.Sprintf("%T", evt.Data),
		"data", evt.Data)

	defer client.Ack(*evt.Request)

	eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
	if !ok {
		b.logger.Debug("Failed to parse Events API event",
			"event_type", evt.Type,
			"data_type", fmt.Sprintf("%T", evt.Data))
		return
	}

	b.logger.Debug("Parsed Events API event",
		"inner_event_type", eventsAPIEvent.Type,
		"team_id", eventsAPIEvent.TeamID,
		"api_app_id", eventsAPIEvent.APIAppID)

	b.logger.Debug("Events API event data", "data", eventsAPIEvent.Data)

	cbEventData, ok := eventsAPIEvent.Data.(*slackevents.EventsAPICallbackEvent)
	if !ok {
		b.logger.Debug("Unable to cast eventData as *slackevents.EventsAPICallbackEvent", "event_type", eventsAPIEvent.Type, "data_type", fmt.Sprintf("%T", eventsAPIEvent.Data))
		return
	}

	b.logger.Debug("Parsed Events API event", "data", cbEventData)

	innerEvent := map[string]any{}
	if err := json.Unmarshal(*cbEventData.InnerEvent, &innerEvent); err != nil {
		b.logger.Debug("Unable to unmarshal innerEvent", "error", err)
		return
	}

	b.logger.Debug("Parsed innerEvent", "data", innerEvent)

	innerEventType, ok := innerEvent["type"].(string)
	if !ok {
		b.logger.Debug("Unable to cast innerEvent type as string", "event_type", eventsAPIEvent.Type, "data_type", fmt.Sprintf("%T", innerEvent["type"]))
		return
	}

	switch innerEventType {
	case "message":
		b.handleMessageEvent(cbEventData)
	case "reaction_added":
		b.handleReactionAddedEvent(cbEventData)
	case "file_shared":
		b.handleFileSharedEvent(cbEventData)
	default:
		b.logger.Debug("Unhandled Events API event data type", "event_type", innerEventType)
		return
	}
}

// handleMessageEvent handles message events from Events API
func (b *Bot) handleMessageEvent(event *slackevents.EventsAPICallbackEvent) {
	msg := slackevents.MessageEvent{}
	if err := json.Unmarshal(*event.InnerEvent, &msg); err != nil {
		b.logger.Debug("Unable to unmarshal MessageEvent", "error", err)
		return
	}

	b.logger.Debug("Received message event",
		"channel", msg.Channel,
		"user", msg.User,
		"message_length", len(msg.Text),
		"timestamp", msg.TimeStamp)

	// Check if this is an incident channel
	if !strings.HasPrefix(msg.Channel, "C") {
		b.logger.Debug("Skipping non-channel message", "channel", msg.Channel)
		return // Not a channel message
	}

	// Get channel info to check if it's an incident channel
	channel, err := b.api.GetConversationInfo(&slack.GetConversationInfoInput{
		ChannelID: msg.Channel,
	})
	if err != nil {
		b.logger.Warn("Failed to get channel info",
			"error", err,
			"channel", msg.Channel)
		return
	}

	b.logger.Debug("Channel info retrieved",
		"channel_id", msg.Channel,
		"channel_name", channel.Name,
		"is_incident_channel", strings.HasPrefix(channel.Name, "_inc-"))

	// Check if this is an incident channel (starts with _inc-)
	if !strings.HasPrefix(channel.Name, "_inc-") {
		b.logger.Debug("Skipping non-incident channel message",
			"channel_name", channel.Name)
		return
	}

	// Find the incident ID for this channel
	incidentID := b.findIncidentIDByChannel(msg.Channel)
	if incidentID == "" {
		b.logger.Warn("No incident ID found for channel",
			"channel_id", msg.Channel,
			"channel_name", channel.Name)
		return
	}

	// Check if we should add this message to the timeline
	shouldAddMessage := b.config.AddAllMessagesToTimeline

	// If not adding all messages, check if this message contains an image
	if !shouldAddMessage {
		// Check if the message contains an image (Slack image URLs)
		if strings.Contains(msg.Text, "files.slack.com") && strings.Contains(msg.Text, "image") {
			shouldAddMessage = true

			b.logger.Debug("Message contains image, will add to timeline",
				"incident_id", incidentID,
				"channel_id", msg.Channel,
				"user", msg.User)
		}
	}

	// Only add to timeline if configured to do so
	if !shouldAddMessage {
		b.logger.Debug("Skipping message (not configured to add all messages and no image detected)",
			"incident_id", incidentID,
			"channel_id", msg.Channel,
			"user", msg.User,
			"message_length", len(msg.Text))

		return
	}

	b.logger.Info("Adding message to incident timeline",
		"incident_id", incidentID,
		"channel_id", msg.Channel,
		"channel_name", channel.Name,
		"user", msg.User,
		"message_length", len(msg.Text))

	// Add message to timeline
	err = b.timelineMgr.AddMessageEntry(incidentID, msg.User, msg.Text, msg.TimeStamp)
	if err != nil {
		b.logger.Error("Failed to add message to timeline",
			"error", err,
			"incident_id", incidentID,
			"channel_id", msg.Channel,
			"user", msg.User)

		return
	}

	b.logger.Info("Message added to timeline successfully",
		"incident_id", incidentID,
		"channel_id", msg.Channel,
		"user", msg.User)
}

// handleReactionAddedEvent handles reaction added events from Events API
func (b *Bot) handleReactionAddedEvent(event *slackevents.EventsAPICallbackEvent) {
	b.logger.Debug("Attempting to parse reaction added event",
		"event", event)

	// Try to parse as the standard slackevents.ReactionAddedEvent first
	reaction := slackevents.ReactionAddedEvent{}
	if err := json.Unmarshal(*event.InnerEvent, &reaction); err != nil {
		b.logger.Debug("Unable to unmarshal ReactionAddedEvent", "error", err)
		return
	}

	b.logger.Debug("Successfully parsed reaction added event",
		"channel", reaction.Item.Channel,
		"user", reaction.User,
		"reaction", reaction.Reaction,
		"timestamp", reaction.Item.Timestamp,
		"item_type", reaction.Item.Type)

	// Check if this is a point_up or point_up_2 reaction
	if reaction.Reaction != "point_up" && reaction.Reaction != "point_up_2" {
		b.logger.Debug("Skipping unhandled reaction",
			"reaction", reaction.Reaction)
		return
	}

	// Check if this is an incident channel
	incidentID := b.findIncidentIDByChannel(reaction.Item.Channel)
	if incidentID == "" {
		b.logger.Warn("No incident ID found for reaction channel",
			"channel_id", reaction.Item.Channel)
		return
	}

	b.logger.Info("Processing point_up reaction for incident",
		"incident_id", incidentID,
		"channel_id", reaction.Item.Channel,
		"user", reaction.User,
		"reaction", reaction.Reaction,
		"message_timestamp", reaction.Item.Timestamp)

	// Get the message that was reacted to
	msg, err := b.api.GetConversationHistory(&slack.GetConversationHistoryParameters{
		ChannelID: reaction.Item.Channel,
		Latest:    reaction.Item.Timestamp,
		Limit:     1,
		Inclusive: true,
	})
	if err != nil {
		b.logger.Error("Failed to get message for reaction",
			"error", err,
			"incident_id", incidentID,
			"channel_id", reaction.Item.Channel,
			"message_timestamp", reaction.Item.Timestamp)

		return
	}

	if len(msg.Messages) == 0 {
		b.logger.Warn("No message found for reaction",
			"incident_id", incidentID,
			"channel_id", reaction.Item.Channel,
			"message_timestamp", reaction.Item.Timestamp)

		return
	}

	messageText := msg.Messages[0].Text
	messageUser := msg.Messages[0].User
	b.logger.Debug("Fetched message for reaction",
		"message_text", messageText,
		"message_user", messageUser,
		"message_ts", msg.Messages[0].Timestamp)

	// Parse Slack timestamp to Go time.Time
	var messageTimestamp time.Time

	if tsFloat, err := strconv.ParseFloat(msg.Messages[0].Timestamp, 64); err == nil {
		sec := int64(tsFloat)
		nsec := int64((tsFloat - float64(sec)) * 1e9)
		messageTimestamp = time.Unix(sec, nsec)
	} else {
		b.logger.Warn("Failed to parse message timestamp, using current time",
			"raw_ts", msg.Messages[0].Timestamp,
			"error", err)

		messageTimestamp = time.Now()
	}

	b.logger.Debug("Parsed message timestamp for reaction",
		"incident_id", incidentID,
		"message_length", len(messageText),
		"message_user", messageUser,
		"parsed_time", messageTimestamp)

	// Add highlighted entry to timeline
	err = b.timelineMgr.AddHighlightedEntry(incidentID, messageUser, messageText, msg.Messages[0].Timestamp, messageTimestamp)
	if err != nil {
		b.logger.Error("Failed to add highlighted entry to timeline",
			"error", err,
			"incident_id", incidentID,
			"channel_id", reaction.Item.Channel,
			"user", messageUser)
	} else {
		b.logger.Info("Highlighted message added to timeline successfully",
			"incident_id", incidentID,
			"channel_id", reaction.Item.Channel,
			"user", messageUser)
	}
}

// handleFileSharedEvent handles file shared events from Events API
func (b *Bot) handleFileSharedEvent(event *slackevents.EventsAPICallbackEvent) {
	file := slackevents.FileSharedEvent{}
	if err := json.Unmarshal(*event.InnerEvent, &file); err != nil {
		b.logger.Debug("Unable to unmarshal FileSharedEvent", "error", err)
		return
	}

	b.logger.Debug("Received file shared event",
		"file_id", file.FileID,
		"user", file.UserID,
		"channel_id", file.ChannelID)

	// Check if this is an incident channel
	incidentID := b.findIncidentIDByChannel(file.ChannelID)
	if incidentID == "" {
		b.logger.Debug("No incident ID found for file channel",
			"channel_id", file.ChannelID,
			"file_id", file.FileID)
		return
	}

	b.logger.Debug("File shared in incident channel",
		"incident_id", incidentID,
		"channel_id", file.ChannelID,
		"file_id", file.FileID)

	// Get file info to check if it's an image
	fileInfo, _, _, err := b.api.GetFileInfo(file.FileID, 0, 0)
	if err != nil {
		b.logger.Error("Failed to get file info",
			"error", err,
			"file_id", file.FileID)
		return
	}

	// Check if it's an image
	if strings.HasPrefix(fileInfo.Mimetype, "image/") {
		caption := fileInfo.Title
		if caption == "" {
			caption = fileInfo.Name
		}

		b.logger.Info("Adding image to incident timeline",
			"incident_id", incidentID,
			"channel_id", file.ChannelID,
			"file_id", file.FileID,
			"file_name", fileInfo.Name,
			"mime_type", fileInfo.Mimetype,
			"caption", caption)

		// Add image to timeline
		err := b.timelineMgr.AddImageEntry(incidentID, file.UserID, fileInfo.URLPrivate, caption, file.FileID)
		if err != nil {
			b.logger.Error("Failed to add image to timeline",
				"error", err,
				"incident_id", incidentID,
				"file_id", file.FileID,
				"user", file.UserID)
		} else {
			b.logger.Info("Image added to timeline successfully",
				"incident_id", incidentID,
				"file_id", file.FileID,
				"user", file.UserID)
		}
	} else {
		b.logger.Debug("Skipping non-image file",
			"incident_id", incidentID,
			"file_id", file.FileID,
			"mime_type", fileInfo.Mimetype)
	}
}

// findIncidentIDByChannel finds the incident ID for a given channel
func (b *Bot) findIncidentIDByChannel(channelID string) string {
	b.mu.RLock()
	incidentID := b.channelToIncident[channelID]
	b.mu.RUnlock()

	if incidentID == "" {
		b.logger.Debug("No incident ID mapping found for channel",
			"channel_id", channelID)
	} else {
		b.logger.Debug("Found incident ID for channel",
			"channel_id", channelID,
			"incident_id", incidentID)
	}

	return incidentID
}
