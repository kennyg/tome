package schema

import (
	"strings"
	"testing"
)

func TestParseClaudeInstructions(t *testing.T) {
	content := `# Project Instructions

This is a Claude project with specific guidelines.

## Rules

- Follow best practices
- Write clean code`

	inst, err := ParseClaudeInstructions([]byte(content))
	if err != nil {
		t.Fatalf("ParseClaudeInstructions() error = %v", err)
	}

	if inst.Body != content {
		t.Errorf("Body = %q, want %q", inst.Body, content)
	}
	if inst.GetFormat() != FormatClaude {
		t.Errorf("GetFormat() = %v, want %v", inst.GetFormat(), FormatClaude)
	}
	if inst.GetName() != "instructions" {
		t.Errorf("GetName() = %q, want %q", inst.GetName(), "instructions")
	}
}

func TestParseOpenCodeInstructions(t *testing.T) {
	content := "# AGENTS.md content"

	inst, err := ParseOpenCodeInstructions([]byte(content))
	if err != nil {
		t.Fatalf("ParseOpenCodeInstructions() error = %v", err)
	}

	if inst.GetFormat() != FormatOpenCode {
		t.Errorf("GetFormat() = %v, want %v", inst.GetFormat(), FormatOpenCode)
	}
}

func TestClaudeInstructions_Serialize(t *testing.T) {
	inst := &ClaudeInstructions{
		Body: "# Instructions\n\nDo these things.",
	}

	got, err := inst.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	// Should be plain markdown, no frontmatter
	if strings.HasPrefix(string(got), "---") {
		t.Error("ClaudeInstructions should not have frontmatter")
	}
	if string(got) != inst.Body {
		t.Errorf("Serialize() = %q, want %q", string(got), inst.Body)
	}
}

func TestParseCopilotInstructions(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantDesc string
		wantTo   string
		wantBody string
	}{
		{
			name: "full instructions",
			content: `---
description: Guidelines for C# development
applyTo: "**/*.cs"
---
# C# Guidelines

Follow these patterns.`,
			wantDesc: "Guidelines for C# development",
			wantTo:   "**/*.cs",
			wantBody: "# C# Guidelines\n\nFollow these patterns.",
		},
		{
			name: "no applyTo",
			content: `---
description: General guidelines
---
Content here`,
			wantDesc: "General guidelines",
			wantBody: "Content here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst, err := ParseCopilotInstructions([]byte(tt.content))
			if err != nil {
				t.Fatalf("ParseCopilotInstructions() error = %v", err)
			}

			if inst.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", inst.Description, tt.wantDesc)
			}
			if inst.ApplyTo != tt.wantTo {
				t.Errorf("ApplyTo = %q, want %q", inst.ApplyTo, tt.wantTo)
			}
			if inst.Body != tt.wantBody {
				t.Errorf("Body = %q, want %q", inst.Body, tt.wantBody)
			}
			if inst.GetFormat() != FormatCopilot {
				t.Errorf("GetFormat() = %v, want %v", inst.GetFormat(), FormatCopilot)
			}
		})
	}
}

func TestCopilotInstructions_Serialize(t *testing.T) {
	inst := &CopilotInstructions{
		Description: "Test guidelines",
		ApplyTo:     "**/*.go",
		Body:        "# Content",
	}

	got, err := inst.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	content := string(got)
	checks := []string{
		"description: Test guidelines",
		"applyTo:",
		"# Content",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("Serialized content should contain %q", check)
		}
	}
}

func TestCopilotInstructions_Filename(t *testing.T) {
	tests := []struct {
		desc string
		want string
	}{
		{"Guidelines for C# development", "guidelines-for-c#.instructions.md"},
		{"", "instructions.instructions.md"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			inst := &CopilotInstructions{Description: tt.desc}
			if got := inst.Filename(); got != tt.want {
				t.Errorf("Filename() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseCursorRules(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantLegacy bool
		wantDesc   string
		wantBody   string
	}{
		{
			name:       "legacy plain text",
			content:    "You are a helpful coding assistant.\n\nFollow best practices.",
			wantLegacy: true,
			wantBody:   "You are a helpful coding assistant.\n\nFollow best practices.",
		},
		{
			name: "MDC format",
			content: `---
description: Coding rules
globs: "**/*.ts"
alwaysApply: true
---
# Rules

Be consistent.`,
			wantLegacy: false,
			wantDesc:   "Coding rules",
			wantBody:   "# Rules\n\nBe consistent.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, err := ParseCursorRules([]byte(tt.content))
			if err != nil {
				t.Fatalf("ParseCursorRules() error = %v", err)
			}

			if rules.isLegacy != tt.wantLegacy {
				t.Errorf("isLegacy = %v, want %v", rules.isLegacy, tt.wantLegacy)
			}
			if rules.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", rules.Description, tt.wantDesc)
			}
			if rules.Body != tt.wantBody {
				t.Errorf("Body = %q, want %q", rules.Body, tt.wantBody)
			}
			if rules.GetFormat() != FormatCursor {
				t.Errorf("GetFormat() = %v, want %v", rules.GetFormat(), FormatCursor)
			}
		})
	}
}

func TestCursorRules_Serialize(t *testing.T) {
	tests := []struct {
		name      string
		rules     *CursorRules
		wantHasFM bool
	}{
		{
			name: "legacy format",
			rules: &CursorRules{
				Body:     "Plain content",
				isLegacy: true,
			},
			wantHasFM: false,
		},
		{
			name: "MDC format",
			rules: &CursorRules{
				Description: "Test rules",
				Globs:       "**/*.go",
				Body:        "Content",
				isLegacy:    false,
			},
			wantHasFM: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.rules.Serialize()
			if err != nil {
				t.Fatalf("Serialize() error = %v", err)
			}

			hasFM := strings.HasPrefix(string(got), "---")
			if hasFM != tt.wantHasFM {
				t.Errorf("Has frontmatter = %v, want %v", hasFM, tt.wantHasFM)
			}
		})
	}
}

func TestIsInstructionsFile(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{"CLAUDE.md", true},
		{"claude.md", true},
		{"AGENTS.md", true},
		{"agents.md", true},
		{"csharp.instructions.md", true},
		{"instructions/general.instructions.md", true},
		{".cursorrules", true},
		{".cursor/rules/coding.mdc", true},
		// Not instructions
		{"SKILL.md", false},
		{"test.agent.md", false},
		{"test.prompt.md", false},
		{".cursor/rules/coding.md", false}, // .md not .mdc
		{"random.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			if got := IsInstructionsFile(tt.filename); got != tt.want {
				t.Errorf("IsInstructionsFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestConvertInstructions(t *testing.T) {
	tests := []struct {
		name       string
		inst       Skill
		target     Format
		wantChecks []string
	}{
		{
			name: "claude to copilot",
			inst: &ClaudeInstructions{
				Body: "# Project Guidelines\n\nFollow these rules.",
			},
			target: FormatCopilot,
			wantChecks: []string{
				"description:",
				"# Project Guidelines",
			},
		},
		{
			name: "copilot to claude",
			inst: &CopilotInstructions{
				Description: "C# Guidelines",
				ApplyTo:     "**/*.cs",
				Body:        "# Rules\n\nBe consistent.",
			},
			target: FormatClaude,
			wantChecks: []string{
				"# Rules",
				"Be consistent.",
			},
		},
		{
			name: "cursor to copilot",
			inst: &CursorRules{
				Description: "Coding rules",
				Body:        "Follow best practices.",
			},
			target: FormatCopilot,
			wantChecks: []string{
				"description: Coding rules",
				"Follow best practices.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertInstructions(tt.inst, tt.target)
			if err != nil {
				t.Fatalf("ConvertInstructions() error = %v", err)
			}

			content := string(got)
			for _, check := range tt.wantChecks {
				if !strings.Contains(content, check) {
					t.Errorf("ConvertInstructions() output should contain %q\nGot: %s", check, content)
				}
			}
		})
	}
}

func TestConvertToClaudeInstructions(t *testing.T) {
	copilot := &CopilotInstructions{
		Description: "Test",
		Body:        "Content",
	}

	inst := ConvertToClaudeInstructions(copilot)
	if inst.Body != "Content" {
		t.Errorf("Body = %q", inst.Body)
	}

	// Test passthrough
	original := &ClaudeInstructions{Body: "Original"}
	same := ConvertToClaudeInstructions(original)
	if same != original {
		t.Error("ConvertToClaudeInstructions should return same instance for ClaudeInstructions")
	}
}

func TestConvertToCopilotInstructions(t *testing.T) {
	claude := &ClaudeInstructions{
		Body: "# My Guidelines\n\nFollow these.",
	}

	inst := ConvertToCopilotInstructions(claude)
	if inst.Body != claude.Body {
		t.Errorf("Body = %q", inst.Body)
	}
	// Should derive description from first line
	if inst.Description != "My Guidelines" {
		t.Errorf("Description = %q, want %q", inst.Description, "My Guidelines")
	}

	// Test passthrough
	original := &CopilotInstructions{Description: "Original"}
	same := ConvertToCopilotInstructions(original)
	if same != original {
		t.Error("ConvertToCopilotInstructions should return same instance for CopilotInstructions")
	}
}

func TestConvertToCursorRules(t *testing.T) {
	copilot := &CopilotInstructions{
		Description: "Test rules",
		Body:        "Content",
	}

	rules := ConvertToCursorRules(copilot)
	if rules.Description != "Test rules" {
		t.Errorf("Description = %q", rules.Description)
	}
	if rules.Body != "Content" {
		t.Errorf("Body = %q", rules.Body)
	}

	// Test passthrough
	original := &CursorRules{Description: "Original"}
	same := ConvertToCursorRules(original)
	if same != original {
		t.Error("ConvertToCursorRules should return same instance for CursorRules")
	}
}

func TestParseInstructions(t *testing.T) {
	content := `# Instructions

Content here.`

	tests := []struct {
		format  Format
		wantErr bool
	}{
		{FormatClaude, false},
		{FormatOpenCode, false},
		{FormatCopilot, false},
		{FormatCursor, false},
		{Format("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			_, err := ParseInstructions([]byte(content), tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInstructions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInstructionsOutputFilename(t *testing.T) {
	inst := &ClaudeInstructions{Body: "test"}

	tests := []struct {
		format Format
		want   string
	}{
		{FormatClaude, "CLAUDE.md"},
		{FormatOpenCode, "AGENTS.md"},
		{FormatCopilot, "project.instructions.md"},
		{FormatCursor, ".cursorrules"},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			if got := InstructionsOutputFilename(inst, tt.format); got != tt.want {
				t.Errorf("InstructionsOutputFilename() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInstructionsOutputDirectory(t *testing.T) {
	inst := &ClaudeInstructions{Body: "test"}

	tests := []struct {
		format Format
		want   string
	}{
		{FormatClaude, ""},
		{FormatOpenCode, ""},
		{FormatCopilot, "instructions"},
		{FormatCursor, ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			if got := InstructionsOutputDirectory(inst, tt.format); got != tt.want {
				t.Errorf("InstructionsOutputDirectory() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConvertInstructionsWithInfo(t *testing.T) {
	inst := &CopilotInstructions{
		Description: "Test",
		ApplyTo:     "**/*.cs",
		Body:        "Content",
	}

	// Convert to Claude - should warn about applyTo
	result, err := ConvertInstructionsWithInfo(inst, FormatClaude)
	if err != nil {
		t.Fatalf("ConvertInstructionsWithInfo() error = %v", err)
	}

	if result.SourceFormat != FormatCopilot {
		t.Errorf("SourceFormat = %v", result.SourceFormat)
	}
	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d: %v", len(result.Warnings), result.Warnings)
	}
}

func TestDetectArtifactType_Instructions(t *testing.T) {
	tests := []struct {
		filename string
		want     ArtifactType
	}{
		{"CLAUDE.md", ArtifactInstructions},
		{"claude.md", ArtifactInstructions},
		{"AGENTS.md", ArtifactInstructions},
		{"csharp.instructions.md", ArtifactInstructions},
		{".cursorrules", ArtifactInstructions},
		{".cursor/rules/coding.mdc", ArtifactInstructions},
		// Not instructions
		{"SKILL.md", ArtifactSkill},
		{"test.agent.md", ArtifactSkill},
		{"test.prompt.md", ArtifactCommand},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			if got := DetectArtifactType(tt.filename); got != tt.want {
				t.Errorf("DetectArtifactType(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

// Round-trip tests
func TestRoundTrip_ClaudeInstructionsToCopilotToClaude(t *testing.T) {
	original := &ClaudeInstructions{
		Body: "# Project Guidelines\n\nFollow best practices.",
	}

	// Claude -> Copilot
	copilotBytes, err := ConvertInstructions(original, FormatCopilot)
	if err != nil {
		t.Fatalf("Convert to Copilot: %v", err)
	}

	copilot, err := ParseCopilotInstructions(copilotBytes)
	if err != nil {
		t.Fatalf("Parse Copilot: %v", err)
	}

	// Copilot -> Claude
	claudeBytes, err := ConvertInstructions(copilot, FormatClaude)
	if err != nil {
		t.Fatalf("Convert to Claude: %v", err)
	}

	result, err := ParseClaudeInstructions(claudeBytes)
	if err != nil {
		t.Fatalf("Parse Claude: %v", err)
	}

	// Body should be preserved
	if strings.TrimSpace(result.Body) != strings.TrimSpace(original.Body) {
		t.Errorf("Body = %q, want %q", result.Body, original.Body)
	}
}
