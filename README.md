# OhShift! - Slack Incident Management Bot

OhShift! is a Slack bot for managing incidents, built in Go. It provides a simple slash command interface to create incident channels and notify team members, and now uses **Slack Socket Mode** (no public HTTP endpoint required).

## Features

- **Slash Command Interface**: Use `/ohshift start <severity> incident <title>` to create incidents
- **Automatic Channel Creation**: Creates incident channels with descriptive names
- **Severity Levels**: Supports SEV1-SEV5 severity levels
- **Notifications**: Posts notifications to a configurable channel
- **Slug Generation**: Automatically converts incident titles to Slack-compatible channel names
- **Help System**: Provides helpful error messages and usage instructions
- **Socket Mode**: Secure, real-time Slack integration without exposing a public HTTP endpoint

## Installation

### Prerequisites

- Go 1.22 or later (tested with Go 1.24)
- Slack workspace with admin permissions
- Slack app with appropriate scopes and **Socket Mode enabled**

### 1. Clone the Repository

```bash
git clone <repository-url>
cd ohshift
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Build the Application

```bash
go build -o ohshift main.go
```

## Configuration

The bot is configured using environment variables:

| Variable                | Description                                 | Default      | Required |
|-------------------------|---------------------------------------------|--------------|----------|
| `SLACK_BOT_TOKEN`       | Slack bot user OAuth token                  | -            | Yes      |
| `SLACK_APP_TOKEN`       | Slack app-level token (for Socket Mode)     | -            | Yes      |
| `SLACK_SIGNING_SECRET`  | Slack app signing secret                    | -            | Yes      |
| `SLASH_COMMAND`         | Slash command to trigger the bot            | `/ohshift`   | No       |
| `NOTIFICATIONS_CHANNEL` | Channel for incident notifications          | `general`    | No       |
| `LOG_LEVEL`             | Logging level (debug, info, warn, error)    | `info`       | No       |

### Example Environment File

Create a `.env` file:

```env
SLACK_BOT_TOKEN=xoxb-your-bot-token-here
SLACK_APP_TOKEN=xapp-your-app-level-token-here
SLACK_SIGNING_SECRET=your-signing-secret-here
SLASH_COMMAND=/ohshift
NOTIFICATIONS_CHANNEL=incidents
LOG_LEVEL=info
```

## Slack App Setup (with Socket Mode)

### 1. Create a Slack App

1. Go to [api.slack.com/apps](https://api.slack.com/apps)
2. Click "Create New App" → "From scratch"
3. Name your app "OhShift!" and select your workspace

### 2. Enable Socket Mode

1. In your app settings, go to **Socket Mode**
2. Toggle **Enable Socket Mode**
3. Generate an **App-Level Token** (starts with `xapp-`)
4. Add this token to your environment as `SLACK_APP_TOKEN`

### 3. Configure Bot Token Scopes

Under "OAuth & Permissions", add these bot token scopes:

- `channels:manage` - Create incident channels
- `chat:write` - Post messages
- `commands` - Handle slash commands

### 4. Install the App

1. Go to "Install App" in the sidebar
2. Click "Install to Workspace"
3. Copy the "Bot User OAuth Token" (starts with `xoxb-`)

### 5. Configure Slash Command

1. Go to "Slash Commands" in the sidebar
2. Click "Create New Command"
3. Configure:
   - Command: `/ohshift`
   - (Request URL is not required for Socket Mode)
   - Short Description: "Manage incidents"
   - Usage Hint: "start <severity> incident <title>"

### 6. Get Signing Secret

1. Go to "Basic Information" in the sidebar
2. Copy the "Signing Secret"

## Usage

### Starting an Incident

Use the slash command format:

```
/ohshift start <severity> incident <incident title>
```

#### Examples:

```
/ohshift start SEV1 incident the website is down
/ohshift start SEV2 incident database connection issues
/ohshift start SEV3 incident slow response times
```

#### Valid Severity Levels:

- `SEV1` - Critical (highest priority)
- `SEV2` - High
- `SEV3` - Medium
- `SEV4` - Low
- `SEV5` - Info (lowest priority)

### What Happens When You Start an Incident

1. **Channel Creation**: A new public channel is created with the name format:
   ```
   _inc-YYYYMMDD-HHMMSS-description
   ```
   Example: `_inc-20241201-143052-website-down`

2. **Initial Message**: A formatted message is posted in the incident channel with:
   - Severity level
   - Who started the incident
   - Description
   - Timestamp

3. **Notification**: A notification is posted in the configured notifications channel:
   ```
   🚨 @username started an incident: SEV1: _inc-20241201-143052-website-down: the website is down
   ```

### Channel Name Generation

The bot automatically converts incident titles to Slack-compatible channel names:

- Converts to lowercase
- Replaces spaces and special characters with hyphens
- Truncates to 64 characters (Slack limit)
- Ensures names don't end with hyphens

## Running the Bot

### Development

```bash
export SLACK_BOT_TOKEN=xoxb-your-token
export SLACK_APP_TOKEN=xapp-your-app-token
export SLACK_SIGNING_SECRET=your-secret
export SLASH_COMMAND=/ohshift
export NOTIFICATIONS_CHANNEL=incidents
export LOG_LEVEL=debug

go run main.go
```

### Production

```bash
./ohshift
```

The bot will connect to Slack via Socket Mode and listen for slash commands in real time. **No public HTTP endpoint is required.**

## Project Structure

```
.
├── config/          # Configuration management
├── incident/        # Incident types and utilities
├── slack/           # Slack API integration (Socket Mode)
├── main.go          # Application entry point
├── go.mod           # Go module file (module path: github.com/fishnix/ohshift)
└── README.md        # This file
```

## Testing

```bash
go test ./...
```

## Building for Different Platforms

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o ohshift-linux main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o ohshift-macos main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o ohshift.exe main.go
```

## Security Considerations

- **Token Security**: Keep your Slack tokens secure and never commit them to version control
- **HTTPS**: Use HTTPS in production for secure communication with Slack (for any outgoing requests)

## Troubleshooting

### Common Issues

1. **"Configuration error"**: Check that all required environment variables are set
2. **"Failed to create channel"**: Ensure the bot has the `channels:manage` scope
3. **"Failed to post notification"**: Verify the notifications channel exists and the bot is a member

### Logs

The bot logs important events to stdout. Check the logs for debugging information.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 