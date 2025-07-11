package incident

import (
	"testing"
	"time"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		want    *Command
		wantErr bool
	}{
		{
			name: "valid SEV1 command",
			text: "start SEV1 incident the website is down",
			want: &Command{
				Action:      "start",
				Severity:    Severity1,
				Title:       "the website is down",
				Description: "",
			},
			wantErr: false,
		},
		{
			name: "valid SEV2 command with long title",
			text: "start SEV2 incident database connection issues causing slow response times",
			want: &Command{
				Action:      "start",
				Severity:    Severity2,
				Title:       "database connection issues causing slow response times",
				Description: "",
			},
			wantErr: false,
		},
		{
			name: "valid SEV1 command with description",
			text: "start SEV1 incident the website is down -- Users cannot access the main application, 500 errors",
			want: &Command{
				Action:      "start",
				Severity:    Severity1,
				Title:       "the website is down",
				Description: "Users cannot access the main application, 500 errors",
			},
			wantErr: false,
		},
		{
			name: "valid SEV2 command with description",
			text: "start SEV2 incident database connection issues -- Connection pool exhausted, affecting all users",
			want: &Command{
				Action:      "start",
				Severity:    Severity2,
				Title:       "database connection issues",
				Description: "Connection pool exhausted, affecting all users",
			},
			wantErr: false,
		},
		{
			name:    "insufficient arguments",
			text:    "start SEV1",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "unknown action",
			text:    "stop SEV1 incident test",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid severity",
			text:    "start SEV6 incident test",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "missing incident keyword",
			text:    "start SEV1 test incident",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty title",
			text:    "start SEV1 incident",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCommand(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Errorf("ParseCommand() returned nil, expected valid command")
				return
			}

			if tt.wantErr && got != nil {
				t.Errorf("ParseCommand() returned command, expected error")
				return
			}

			if got != nil {
				if got.Action != tt.want.Action {
					t.Errorf("ParseCommand() Action = %v, want %v", got.Action, tt.want.Action)
				}

				if got.Severity != tt.want.Severity {
					t.Errorf("ParseCommand() Severity = %v, want %v", got.Severity, tt.want.Severity)
				}

				if got.Title != tt.want.Title {
					t.Errorf("ParseCommand() Title = %v, want %v", got.Title, tt.want.Title)
				}

				if got.Description != tt.want.Description {
					t.Errorf("ParseCommand() Description = %v, want %v", got.Description, tt.want.Description)
				}
			}
		})
	}
}

func TestGenerateChannelName(t *testing.T) {
	now := time.Date(2024, 12, 1, 14, 30, 52, 0, time.UTC)

	tests := []struct {
		name     string
		incident *Incident
		want     string
	}{
		{
			name: "simple incident",
			incident: &Incident{
				Title:     "website down",
				StartedAt: now,
			},
			want: "_inc-20241201-143052-website-down",
		},
		{
			name: "incident with special characters",
			incident: &Incident{
				Title:     "Database Connection Issues!",
				StartedAt: now,
			},
			want: "_inc-20241201-143052-database-connection-issues",
		},
		{
			name: "incident with numbers",
			incident: &Incident{
				Title:     "API v2.1 failing",
				StartedAt: now,
			},
			want: "_inc-20241201-143052-api-v2-1-failing",
		},
		{
			name: "very long incident title",
			incident: &Incident{
				Title:     "This is a very long incident title that should be truncated to fit within Slack's channel name limits and ensure it doesn't exceed the maximum allowed length",
				StartedAt: now,
			},
			want: "_inc-20241201-143052-this-is-a-very-long-incident",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateChannelName(tt.incident)
			if got != tt.want {
				t.Errorf("GenerateChannelName() = %v, want %v", got, tt.want)
			}

			// Verify channel name length is within Slack limits
			if len(got) > 64 {
				t.Errorf("GenerateChannelName() length = %d, want <= 64", len(got))
			}
		})
	}
}

func TestCreateSlug(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple text",
			input: "website down",
			want:  "website-down",
		},
		{
			name:  "text with special characters",
			input: "Database Connection Issues!",
			want:  "database-connection-issues",
		},
		{
			name:  "text with numbers",
			input: "API v2.1 failing",
			want:  "api-v2-1-failing",
		},
		{
			name:  "text with multiple spaces",
			input: "  multiple   spaces  ",
			want:  "multiple-spaces",
		},
		{
			name:  "text with leading/trailing hyphens",
			input: "-test-",
			want:  "test",
		},
		{
			name:  "very long text",
			input: "This is a very long text that should be truncated to prevent overly long slugs",
			want:  "this-is-a-very-long-text-that",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createSlug(tt.input)
			if got != tt.want {
				t.Errorf("createSlug() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidSeverity(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
		want     bool
	}{
		{"SEV1", Severity1, true},
		{"SEV2", Severity2, true},
		{"SEV3", Severity3, true},
		{"SEV4", Severity4, true},
		{"SEV5", Severity5, true},
		{"SEV6", "SEV6", false},
		{"sev1", "sev1", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidSeverity(tt.severity)
			if got != tt.want {
				t.Errorf("isValidSeverity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetHelpMessage(t *testing.T) {
	help := GetHelpMessage()

	// Check that help message contains expected content
	expectedContent := []string{
		"Usage:",
		"/ohshift start",
		"SEV1",
		"incident",
		"Examples:",
	}

	for _, content := range expectedContent {
		if !contains(help, content) {
			t.Errorf("GetHelpMessage() missing expected content: %s", content)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
