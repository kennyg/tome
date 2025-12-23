package schema

import (
	"testing"
)

func TestFormat_IsValid(t *testing.T) {
	tests := []struct {
		format Format
		want   bool
	}{
		{FormatClaude, true},
		{FormatOpenCode, true},
		{FormatCopilot, true},
		{FormatCursor, true},
		{Format("unknown"), false},
		{Format(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			if got := tt.format.IsValid(); got != tt.want {
				t.Errorf("Format(%q).IsValid() = %v, want %v", tt.format, got, tt.want)
			}
		})
	}
}

func TestFormat_String(t *testing.T) {
	tests := []struct {
		format Format
		want   string
	}{
		{FormatClaude, "claude"},
		{FormatOpenCode, "opencode"},
		{FormatCopilot, "copilot"},
		{FormatCursor, "cursor"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.format.String(); got != tt.want {
				t.Errorf("Format.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllFormats(t *testing.T) {
	formats := AllFormats()
	if len(formats) != 4 {
		t.Errorf("AllFormats() returned %d formats, want 4", len(formats))
	}

	// Verify all formats are valid
	for _, f := range formats {
		if !f.IsValid() {
			t.Errorf("AllFormats() contains invalid format: %s", f)
		}
	}
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     Format
	}{
		// Copilot patterns
		{"copilot agent", "CSharpExpert.agent.md", FormatCopilot},
		{"copilot agent path", "agents/CSharpExpert.agent.md", FormatCopilot},
		{"copilot prompt", "create-readme.prompt.md", FormatCopilot},
		{"copilot prompt path", "prompts/create-readme.prompt.md", FormatCopilot},

		// Cursor patterns
		{"cursor rules", ".cursor/rules/coding.md", FormatCursor},
		{"cursor path", "project/.cursor/settings.md", FormatCursor},

		// OpenCode patterns
		{"opencode skill", ".opencode/skill/test/SKILL.md", FormatOpenCode},
		{"opencode path", "project/.opencode/command/test.md", FormatOpenCode},

		// Claude patterns
		{"claude skill", ".claude/skills/pdf/SKILL.md", FormatClaude},
		{"claude path", "project/.claude/commands/test.md", FormatClaude},
		{"skill.md default", "skills/test/SKILL.md", FormatClaude},
		{"skill.md lowercase", "skills/test/skill.md", FormatClaude},

		// Default to Claude
		{"unknown", "random.md", FormatClaude},
		{"no extension", "README", FormatClaude},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectFormat(tt.filename, nil)
			if got != tt.want {
				t.Errorf("DetectFormat(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}
