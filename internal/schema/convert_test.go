package schema

import (
	"strings"
	"testing"
)

func TestConvert(t *testing.T) {
	tests := []struct {
		name       string
		skill      Skill
		target     Format
		wantChecks []string
		wantErr    bool
	}{
		{
			name: "claude to copilot",
			skill: &ClaudeSkill{
				Name:        "test-skill",
				Description: "A test skill",
				Version:     "1.0.0",
				Body:        "# Content",
			},
			target: FormatCopilot,
			wantChecks: []string{
				"name: test-skill",
				"description: A test skill",
				"version: 1.0.0",
				"# Content",
			},
		},
		{
			name: "copilot to claude",
			skill: &CopilotAgent{
				Name:        "C# Expert",
				Description: "An agent",
				Version:     "2025-01-01",
				Body:        "You are an expert.",
			},
			target: FormatClaude,
			wantChecks: []string{
				"name: C# Expert",
				"description: An agent",
				"2025-01-01", // YAML may quote the version
				"You are an expert.",
			},
		},
		{
			name: "claude to cursor",
			skill: &ClaudeSkill{
				Name:        "test",
				Description: "A test",
				Body:        "Body content",
			},
			target: FormatCursor,
			wantChecks: []string{
				"name: test",
				"description: A test",
				"Body content",
			},
		},
		{
			name: "cursor to copilot",
			skill: &CursorSkill{
				Name:        "rules",
				Description: "Coding rules",
				Body:        "Be consistent.",
			},
			target: FormatCopilot,
			wantChecks: []string{
				"name: rules",
				"description: Coding rules",
				"Be consistent.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Convert(tt.skill, tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			content := string(got)
			for _, check := range tt.wantChecks {
				if !strings.Contains(content, check) {
					t.Errorf("Convert() output should contain %q\nGot: %s", check, content)
				}
			}
		})
	}
}

func TestConvert_InvalidFormat(t *testing.T) {
	skill := &ClaudeSkill{Name: "test", Description: "test", Body: "body"}
	_, err := Convert(skill, Format("invalid"))
	if err == nil {
		t.Error("Convert() should return error for invalid format")
	}
}

func TestConvertToClaudeSkill(t *testing.T) {
	tests := []struct {
		name  string
		skill Skill
		check func(*testing.T, *ClaudeSkill)
	}{
		{
			name: "from copilot",
			skill: &CopilotAgent{
				Name:        "Copilot Agent",
				Description: "An agent",
				Version:     "1.0.0",
				Body:        "Content",
			},
			check: func(t *testing.T, cs *ClaudeSkill) {
				if cs.Name != "Copilot Agent" {
					t.Errorf("Name = %q", cs.Name)
				}
				if cs.Version != "1.0.0" {
					t.Errorf("Version = %q", cs.Version)
				}
			},
		},
		{
			name: "from cursor",
			skill: &CursorSkill{
				Name:        "Cursor Skill",
				Description: "A skill",
				Body:        "Content",
			},
			check: func(t *testing.T, cs *ClaudeSkill) {
				if cs.Name != "Cursor Skill" {
					t.Errorf("Name = %q", cs.Name)
				}
			},
		},
		{
			name: "from claude (passthrough)",
			skill: &ClaudeSkill{
				Name:   "Original",
				Author: "tester",
			},
			check: func(t *testing.T, cs *ClaudeSkill) {
				if cs.Author != "tester" {
					t.Errorf("Author should be preserved: %q", cs.Author)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertToClaudeSkill(tt.skill)
			tt.check(t, result)
		})
	}
}

func TestConvertToCopilotAgent(t *testing.T) {
	tests := []struct {
		name  string
		skill Skill
		check func(*testing.T, *CopilotAgent)
	}{
		{
			name: "from claude",
			skill: &ClaudeSkill{
				Name:        "Claude Skill",
				Description: "A skill",
				Version:     "2.0.0",
				Body:        "Content",
			},
			check: func(t *testing.T, ca *CopilotAgent) {
				if ca.Name != "Claude Skill" {
					t.Errorf("Name = %q", ca.Name)
				}
				if ca.Version != "2.0.0" {
					t.Errorf("Version = %q", ca.Version)
				}
			},
		},
		{
			name: "from copilot (passthrough)",
			skill: &CopilotAgent{
				Name:    "Original",
				Version: "original-version",
			},
			check: func(t *testing.T, ca *CopilotAgent) {
				if ca.Version != "original-version" {
					t.Errorf("Version should be preserved: %q", ca.Version)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertToCopilotAgent(tt.skill)
			tt.check(t, result)
		})
	}
}

func TestConvertToCursorSkill(t *testing.T) {
	skill := &ClaudeSkill{
		Name:        "Test",
		Description: "A test",
		Body:        "Content",
	}

	result := ConvertToCursorSkill(skill)
	if result.Name != "Test" {
		t.Errorf("Name = %q", result.Name)
	}
	if result.Body != "Content" {
		t.Errorf("Body = %q", result.Body)
	}

	// Test passthrough
	original := &CursorSkill{Name: "Original"}
	same := ConvertToCursorSkill(original)
	if same != original {
		t.Error("ConvertToCursorSkill should return same instance for CursorSkill")
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		content string
		format  Format
		wantErr bool
	}{
		{
			name: "claude format",
			content: `---
name: test
description: A test
---
Body`,
			format: FormatClaude,
		},
		{
			name: "opencode format",
			content: `---
name: test
description: A test
---
Body`,
			format: FormatOpenCode,
		},
		{
			name: "copilot format",
			content: `---
name: test
description: A test
---
Body`,
			format: FormatCopilot,
		},
		{
			name: "cursor format",
			content: `---
name: test
---
Body`,
			format: FormatCursor,
		},
		{
			name:    "invalid format",
			content: "test",
			format:  Format("invalid"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.content), tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseAuto(t *testing.T) {
	content := `---
name: test
description: A test
---
Body`

	tests := []struct {
		filename   string
		wantFormat Format
	}{
		{"test.agent.md", FormatCopilot},
		{".claude/skills/test/SKILL.md", FormatClaude},
		{".cursor/rules/coding.md", FormatCursor},
		{".opencode/skill/test/SKILL.md", FormatOpenCode},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			skill, err := ParseAuto([]byte(content), tt.filename)
			if err != nil {
				t.Fatalf("ParseAuto() error = %v", err)
			}
			if skill.GetFormat() != tt.wantFormat {
				t.Errorf("GetFormat() = %v, want %v", skill.GetFormat(), tt.wantFormat)
			}
		})
	}
}

func TestConvertWithInfo(t *testing.T) {
	skill := &ClaudeSkill{
		Name:         "test",
		Description:  "A test",
		Globs:        []string{"*.go"},
		Includes:     []string{"helper.go"},
		AllowedTools: []string{"Bash"},
		Body:         "Content",
	}

	// Convert to Copilot - should warn about lost fields
	result, err := ConvertWithInfo(skill, FormatCopilot)
	if err != nil {
		t.Fatalf("ConvertWithInfo() error = %v", err)
	}

	if result.SourceFormat != FormatClaude {
		t.Errorf("SourceFormat = %v", result.SourceFormat)
	}
	if result.TargetFormat != FormatCopilot {
		t.Errorf("TargetFormat = %v", result.TargetFormat)
	}
	if len(result.Warnings) != 3 {
		t.Errorf("Expected 3 warnings, got %d: %v", len(result.Warnings), result.Warnings)
	}

	// Convert to Cursor - should warn about allowed-tools only
	result2, err := ConvertWithInfo(skill, FormatCursor)
	if err != nil {
		t.Fatalf("ConvertWithInfo() error = %v", err)
	}
	if len(result2.Warnings) != 1 {
		t.Errorf("Expected 1 warning for Cursor, got %d: %v", len(result2.Warnings), result2.Warnings)
	}
}

func TestOutputFilename(t *testing.T) {
	skill := &ClaudeSkill{
		Name:        "TestSkill", // No space to avoid double-dash
		Description: "A test",
	}

	tests := []struct {
		format Format
		want   string
	}{
		{FormatClaude, "SKILL.md"},
		{FormatOpenCode, "SKILL.md"},
		{FormatCopilot, "test-skill.agent.md"},
		{FormatCursor, "test-skill.md"},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			if got := OutputFilename(skill, tt.format); got != tt.want {
				t.Errorf("OutputFilename() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOutputFilename_NoName(t *testing.T) {
	skill := &ClaudeSkill{Description: "test"}

	if got := OutputFilename(skill, FormatCopilot); got != "skill.agent.md" {
		t.Errorf("OutputFilename() for empty name = %q, want %q", got, "skill.agent.md")
	}
}

func TestOutputDirectory(t *testing.T) {
	skill := &ClaudeSkill{
		Name:        "TestSkill", // No space to avoid double-dash
		Description: "A test",
	}

	tests := []struct {
		format Format
		want   string
	}{
		{FormatClaude, "skills/test-skill"},
		{FormatOpenCode, ".opencode/skill/test-skill"},
		{FormatCopilot, "agents"},
		{FormatCursor, ".cursor/rules"},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			if got := OutputDirectory(skill, tt.format); got != tt.want {
				t.Errorf("OutputDirectory() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"TestSkill", "test-skill"},
		{"test skill", "test-skill"},
		{"test_skill", "test-skill"},
		{"already-kebab", "already-kebab"},
		{"CSharpExpert", "c-sharp-expert"},
		{"PDF", "p-d-f"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := toKebabCase(tt.input); got != tt.want {
				t.Errorf("toKebabCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// Round-trip tests to ensure data integrity through conversions
func TestRoundTrip_ClaudeToCopilotToClaude(t *testing.T) {
	original := &ClaudeSkill{
		Name:        "test-skill",
		Description: "A test skill",
		Version:     "1.0.0",
		Body:        "# Content\n\nBody text here.",
	}

	// Claude -> Copilot
	copilotBytes, err := Convert(original, FormatCopilot)
	if err != nil {
		t.Fatalf("Convert to Copilot: %v", err)
	}

	copilot, err := ParseCopilotAgent(copilotBytes)
	if err != nil {
		t.Fatalf("Parse Copilot: %v", err)
	}

	// Copilot -> Claude
	claudeBytes, err := Convert(copilot, FormatClaude)
	if err != nil {
		t.Fatalf("Convert to Claude: %v", err)
	}

	result, err := ParseClaudeSkill(claudeBytes)
	if err != nil {
		t.Fatalf("Parse Claude: %v", err)
	}

	// Verify preserved fields
	if result.Name != original.Name {
		t.Errorf("Name = %q, want %q", result.Name, original.Name)
	}
	if result.Description != original.Description {
		t.Errorf("Description = %q, want %q", result.Description, original.Description)
	}
	if result.Version != original.Version {
		t.Errorf("Version = %q, want %q", result.Version, original.Version)
	}
	// Body may have leading newlines due to serialization
	if strings.TrimSpace(result.Body) != strings.TrimSpace(original.Body) {
		t.Errorf("Body = %q, want %q", result.Body, original.Body)
	}
}

func TestRoundTrip_CopilotToClaudeToCopilot(t *testing.T) {
	original := &CopilotAgent{
		Name:        "C# Expert",
		Description: "An expert agent",
		Version:     "2025-01-01",
		Body:        "You are an expert.",
	}

	// Copilot -> Claude
	claudeBytes, err := Convert(original, FormatClaude)
	if err != nil {
		t.Fatalf("Convert to Claude: %v", err)
	}

	claude, err := ParseClaudeSkill(claudeBytes)
	if err != nil {
		t.Fatalf("Parse Claude: %v", err)
	}

	// Claude -> Copilot
	copilotBytes, err := Convert(claude, FormatCopilot)
	if err != nil {
		t.Fatalf("Convert to Copilot: %v", err)
	}

	result, err := ParseCopilotAgent(copilotBytes)
	if err != nil {
		t.Fatalf("Parse Copilot: %v", err)
	}

	// Verify preserved fields
	if result.Name != original.Name {
		t.Errorf("Name = %q, want %q", result.Name, original.Name)
	}
	if result.Description != original.Description {
		t.Errorf("Description = %q, want %q", result.Description, original.Description)
	}
	if result.Version != original.Version {
		t.Errorf("Version = %q, want %q", result.Version, original.Version)
	}
	// Body may have leading newline due to serialization
	if strings.TrimSpace(result.Body) != strings.TrimSpace(original.Body) {
		t.Errorf("Body = %q, want %q", result.Body, original.Body)
	}
}
