# OhShift! - Slack Incident Management Bot

OhShift! is a Slack bot for managing incidents, built in Go. It provides a simple slash command interface to create incident channels and notify team members, and uses **Slack Socket Mode** (no public HTTP endpoint required).

## Features

- **Slash Command Interface**: Use `/shift start <severity> incident <title>` to create incidents
- **Automatic Channel Creation**: Creates incident channels with descriptive names
- **Severity Levels**: Supports SEV0-SEV3 severity levels
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
| `DB_URI`                | PostgreSQL database connection string       | -            | Yes      |
| `SLASH_COMMAND`         | Slash command to trigger the bot            | `/shift`   | No       |
| `NOTIFICATIONS_CHANNEL` | Channel for incident notifications          | `general`    | No       |
| `LOG_LEVEL`             | Logging level (debug, info, warn, error)    | `info`       | No       |

### Example Environment File

Create a `.env` file:

```env
SLACK_BOT_TOKEN=xoxb-your-bot-token
SLACK_APP_TOKEN=xapp-your-app-level-token
SLACK_SIGNING_SECRET=your-signing-secret
DB_URI=postgres://username:password@localhost:5432/ohshift?sslmode=disable
SLASH_COMMAND=/shift
NOTIFICATIONS_CHANNEL=incidents
LOG_LEVEL=info
```

## Slack App Setup (with Socket Mode)

### 1. Create a Slack App

1. Go to [api.slack.com/apps](https://api.slack.com/apps)
2. Click "Create New App" → "From manifest"
3. Select your workspace.
4. Modify the example manifest (`slack.example.manifest.json`) and past the contents
5. Review the summary and click create

### 2. Generate an app token

1. In your app settings, go to **Basic Information**
2. Scroll to **App-Level Tokens** , click **Generate Token and Scopes**
3. Set the name and add the `connections:write` scope.
3. Click **Generate** and save the token (starts with `xapp-`)
4. Add this token to your environment as `SLACK_APP_TOKEN`

### 3. Install the App

1. Go to "Install App" in the sidebar
2. Click "Install to Workspace"
3. Copy the "Bot User OAuth Token" (starts with `xoxb-`)

### 4. Configure Slash Command

1. Go to "Slash Commands" in the sidebar
2. Click "Create New Command"
3. Configure:
   - Command: `/shift`
   - (Request URL is not required for Socket Mode)
   - Short Description: "Manage incidents"
   - Usage Hint: "start <severity> incident <title>"

### 5. Get Signing Secret

1. Go to "Basic Information" in the sidebar
2. Copy the "Signing Secret"

## Usage

### Starting an Incident

Use the slash command format:

```
/shift start <severity> incident <incident title>
```

#### Examples:

```
/shift start SEV0 incident the website is down
/shift start SEV1 incident database connection issues
/shift start SEV2 incident slow response times
```

#### Valid Severity Levels:

- `SEV0` - Major Customer Impact (highest priority)
- `SEV1` - High Customer Impact
- `SEV2` - Low/No Customer Impact
- `SEV3` - Maintenance (lowest priority)

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
   🚨 @username started an incident: SEV0: _inc-20241201-143052-website-down: the website is down
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
export DB_URI=postgres://username:password@localhost:5432/ohshift?sslmode=disable
export SLASH_COMMAND=/shift
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

## Development

### Database Models

The application uses [SQLBoiler](https://github.com/volatiletech/sqlboiler) to generate Go models from the database schema. To regenerate the models after schema changes:

```bash
make gen-models
```

This command will:
1. Create a temporary database with the current schema
2. Generate Go models using SQLBoiler
3. Generate an Entity Relationship Diagram (ERD) in Mermaid format

The automatically generated ERD can be found at: [docs/gen_models_erd.md](docs/gen_models_erd.md)

### Available Make Commands

```bash
make help          # Show all available commands
make build         # Build the application
make test          # Run tests
make lint          # Run golangci-lint
make gen-models    # Generate database models and ERD
make clean         # Clean build artifacts
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