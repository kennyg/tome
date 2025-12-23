package schema

import (
	"strings"
	"testing"
)

func TestParseClaudeCommand(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantName    string
		wantDesc    string
		wantVersion string
		wantBody    string
		wantErr     bool
	}{
		{
			name: "full command",
			content: `---
name: commit
description: Create a git commit
version: 1.0.0
author: kennyg
allowed-tools:
  - Bash
---
# Commit Command

Create a well-formatted git commit.`,
			wantName:    "commit",
			wantDesc:    "Create a git commit",
			wantVersion: "1.0.0",
			wantBody:    "# Commit Command\n\nCreate a well-formatted git commit.",
		},
		{
			name: "minimal command",
			content: `---
name: test
description: Run tests
---
Run the test suite.`,
			wantName: "test",
			wantDesc: "Run tests",
			wantBody: "Run the test suite.",
		},
		{
			name:     "no frontmatter",
			content:  "Just run the command.",
			wantBody: "Just run the command.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ParseClaudeCommand([]byte(tt.content))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseClaudeCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if cmd.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", cmd.Name, tt.wantName)
			}
			if cmd.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", cmd.Description, tt.wantDesc)
			}
			if cmd.Version != tt.wantVersion {
				t.Errorf("Version = %q, want %q", cmd.Version, tt.wantVersion)
			}
			if cmd.Body != tt.wantBody {
				t.Errorf("Body = %q, want %q", cmd.Body, tt.wantBody)
			}
			if cmd.GetFormat() != FormatClaude {
				t.Errorf("GetFormat() = %v, want %v", cmd.GetFormat(), FormatClaude)
			}
		})
	}
}

func TestParseOpenCodeCommand(t *testing.T) {
	content := `---
name: build
description: Build the project
---
Run the build command.`

	cmd, err := ParseOpenCodeCommand([]byte(content))
	if err != nil {
		t.Fatalf("ParseOpenCodeCommand() error = %v", err)
	}

	if cmd.GetFormat() != FormatOpenCode {
		t.Errorf("GetFormat() = %v, want %v", cmd.GetFormat(), FormatOpenCode)
	}
	if cmd.Name != "build" {
		t.Errorf("Name = %q, want %q", cmd.Name, "build")
	}
}

func TestClaudeCommand_Serialize(t *testing.T) {
	cmd := &ClaudeCommand{
		Name:         "deploy",
		Description:  "Deploy the application",
		Version:      "2.0.0",
		Author:       "tester",
		AllowedTools: []string{"Bash", "Read"},
		Body:         "# Deploy\n\nDeploy to production.",
	}

	got, err := cmd.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	content := string(got)

	checks := []string{
		"name: deploy",
		"description: Deploy the application",
		"version: 2.0.0",
		"author: tester",
		"allowed-tools:",
		"Bash",
		"# Deploy",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("Serialized content should contain %q", check)
		}
	}
}

func TestClaudeCommand_Filename(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"commit", "commit.md"},
		{"Git Commit", "git-commit.md"},
		{"", "command.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &ClaudeCommand{Name: tt.name}
			if got := cmd.Filename(); got != tt.want {
				t.Errorf("Filename() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClaudeCommand_Interface(t *testing.T) {
	cmd := &ClaudeCommand{
		Name:        "test-cmd",
		Description: "Testing",
		Body:        "Content",
	}

	// Verify it implements Skill interface
	var _ Skill = cmd

	if cmd.GetName() != "test-cmd" {
		t.Errorf("GetName() = %q", cmd.GetName())
	}
	if cmd.GetDescription() != "Testing" {
		t.Errorf("GetDescription() = %q", cmd.GetDescription())
	}
	if cmd.GetBody() != "Content" {
		t.Errorf("GetBody() = %q", cmd.GetBody())
	}
}

func TestClaudeCommand_Metadata(t *testing.T) {
	cmd := &ClaudeCommand{
		Name:        "meta-test",
		Description: "Metadata test",
		Version:     "1.0.0",
		Author:      "author",
		Body:        "Body",
	}

	// Test ToMetadata
	meta := cmd.ToMetadata()
	if meta.Name != "meta-test" {
		t.Errorf("ToMetadata().Name = %q", meta.Name)
	}
	if meta.Version != "1.0.0" {
		t.Errorf("ToMetadata().Version = %q", meta.Version)
	}

	// Test FromMetadata
	newCmd := &ClaudeCommand{}
	newCmd.FromMetadata(SkillMetadata{
		Name:        "from-meta",
		Description: "From metadata",
		Version:     "2.0.0",
		Author:      "new-author",
		Body:        "New body",
	})

	if newCmd.Name != "from-meta" {
		t.Errorf("FromMetadata Name = %q", newCmd.Name)
	}
	if newCmd.Body != "New body" {
		t.Errorf("FromMetadata Body = %q", newCmd.Body)
	}
}

func TestIsCommandFile(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{".claude/commands/commit.md", true},
		{".opencode/command/build.md", true},
		{"prompts/create.prompt.md", true},
		{".claude/skills/pdf/SKILL.md", false},
		{"agents/test.agent.md", false},
		{"random.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			if got := IsCommandFile(tt.filename); got != tt.want {
				t.Errorf("IsCommandFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestConvertCommand(t *testing.T) {
	tests := []struct {
		name       string
		cmd        Skill
		target     Format
		wantChecks []string
	}{
		{
			name: "claude to copilot",
			cmd: &ClaudeCommand{
				Name:        "commit",
				Description: "Create commit",
				Body:        "Create a git commit.",
			},
			target: FormatCopilot,
			wantChecks: []string{
				"agent: commit",
				"description: Create commit",
				"Create a git commit.",
			},
		},
		{
			name: "copilot to claude",
			cmd: &CopilotPrompt{
				Agent:       "review",
				Description: "Review code",
				Body:        "Review the code changes.",
			},
			target: FormatClaude,
			wantChecks: []string{
				"name: review",
				"description: Review code",
				"Review the code changes.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertCommand(tt.cmd, tt.target)
			if err != nil {
				t.Fatalf("ConvertCommand() error = %v", err)
			}

			content := string(got)
			for _, check := range tt.wantChecks {
				if !strings.Contains(content, check) {
					t.Errorf("ConvertCommand() output should contain %q\nGot: %s", check, content)
				}
			}
		})
	}
}

func TestConvertToClaudeCommand(t *testing.T) {
	prompt := &CopilotPrompt{
		Agent:       "test-prompt",
		Description: "A test prompt",
		Body:        "Content",
	}

	cmd := ConvertToClaudeCommand(prompt)
	if cmd.Name != "test-prompt" {
		t.Errorf("Name = %q", cmd.Name)
	}
	if cmd.Description != "A test prompt" {
		t.Errorf("Description = %q", cmd.Description)
	}

	// Test passthrough
	original := &ClaudeCommand{Name: "original", Author: "author"}
	same := ConvertToClaudeCommand(original)
	if same != original {
		t.Error("ConvertToClaudeCommand should return same instance for ClaudeCommand")
	}
}

func TestConvertToCopilotPrompt(t *testing.T) {
	cmd := &ClaudeCommand{
		Name:        "test-cmd",
		Description: "A test command",
		Body:        "Content",
	}

	prompt := ConvertToCopilotPrompt(cmd)
	if prompt.Agent != "test-cmd" {
		t.Errorf("Agent = %q", prompt.Agent)
	}
	if prompt.Description != "A test command" {
		t.Errorf("Description = %q", prompt.Description)
	}

	// Test passthrough
	original := &CopilotPrompt{Agent: "original"}
	same := ConvertToCopilotPrompt(original)
	if same != original {
		t.Error("ConvertToCopilotPrompt should return same instance for CopilotPrompt")
	}
}

func TestParseCommand(t *testing.T) {
	content := `---
name: test
description: A test
---
Body`

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
			_, err := ParseCommand([]byte(content), tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommandOutputFilename(t *testing.T) {
	cmd := &ClaudeCommand{
		Name:        "TestCmd",
		Description: "A test",
	}

	tests := []struct {
		format Format
		want   string
	}{
		{FormatClaude, "test-cmd.md"},
		{FormatOpenCode, "test-cmd.md"},
		{FormatCopilot, "test-cmd.prompt.md"},
		{FormatCursor, "test-cmd.md"},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			if got := CommandOutputFilename(cmd, tt.format); got != tt.want {
				t.Errorf("CommandOutputFilename() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCommandOutputDirectory(t *testing.T) {
	cmd := &ClaudeCommand{Name: "test"}

	tests := []struct {
		format Format
		want   string
	}{
		{FormatClaude, "commands"},
		{FormatOpenCode, ".opencode/command"},
		{FormatCopilot, "prompts"},
		{FormatCursor, ".cursor/rules"},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			if got := CommandOutputDirectory(cmd, tt.format); got != tt.want {
				t.Errorf("CommandOutputDirectory() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConvertCommandWithInfo(t *testing.T) {
	cmd := &ClaudeCommand{
		Name:         "test",
		Description:  "A test",
		Version:      "1.0.0",
		Author:       "author",
		AllowedTools: []string{"Bash"},
		Body:         "Content",
	}

	// Convert to Copilot - should warn about lost fields
	result, err := ConvertCommandWithInfo(cmd, FormatCopilot)
	if err != nil {
		t.Fatalf("ConvertCommandWithInfo() error = %v", err)
	}

	if result.SourceFormat != FormatClaude {
		t.Errorf("SourceFormat = %v", result.SourceFormat)
	}
	if result.TargetFormat != FormatCopilot {
		t.Errorf("TargetFormat = %v", result.TargetFormat)
	}
	// Should have warnings for allowed-tools, version, and author
	if len(result.Warnings) != 3 {
		t.Errorf("Expected 3 warnings, got %d: %v", len(result.Warnings), result.Warnings)
	}

	// Convert to Cursor - should warn about no command concept
	result2, err := ConvertCommandWithInfo(cmd, FormatCursor)
	if err != nil {
		t.Fatalf("ConvertCommandWithInfo() error = %v", err)
	}
	// Should have warning about Cursor not having commands + allowed-tools
	if len(result2.Warnings) < 1 {
		t.Errorf("Expected at least 1 warning for Cursor, got %d", len(result2.Warnings))
	}
}

func TestDetectArtifactType(t *testing.T) {
	tests := []struct {
		filename string
		want     ArtifactType
	}{
		// Skills
		{"test.agent.md", ArtifactSkill},
		{".claude/skills/pdf/SKILL.md", ArtifactSkill},
		{".opencode/skill/test/SKILL.md", ArtifactSkill},
		{".cursor/rules/coding.md", ArtifactSkill},

		// Commands
		{"test.prompt.md", ArtifactCommand},
		{".claude/commands/commit.md", ArtifactCommand},
		{".opencode/command/build.md", ArtifactCommand},

		// Default to skill
		{"random.md", ArtifactSkill},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			if got := DetectArtifactType(tt.filename); got != tt.want {
				t.Errorf("DetectArtifactType(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestCopilotPrompt_Metadata(t *testing.T) {
	prompt := &CopilotPrompt{
		Agent:       "test-agent",
		Description: "Test description",
		Body:        "Body content",
	}

	// Test ToMetadata
	meta := prompt.ToMetadata()
	if meta.Name != "test-agent" {
		t.Errorf("ToMetadata().Name = %q, want %q", meta.Name, "test-agent")
	}
	if meta.Description != "Test description" {
		t.Errorf("ToMetadata().Description = %q", meta.Description)
	}

	// Test FromMetadata
	newPrompt := &CopilotPrompt{}
	newPrompt.FromMetadata(SkillMetadata{
		Name:        "from-meta",
		Description: "From metadata",
		Body:        "New body",
	})

	if newPrompt.Agent != "from-meta" {
		t.Errorf("FromMetadata Agent = %q", newPrompt.Agent)
	}
	if newPrompt.Body != "New body" {
		t.Errorf("FromMetadata Body = %q", newPrompt.Body)
	}
}

func TestCopilotPrompt_Filename(t *testing.T) {
	tests := []struct {
		agent string
		want  string
	}{
		{"commit", "commit.prompt.md"},
		{"Code Review", "code-review.prompt.md"},
		{"", "command.prompt.md"},
	}

	for _, tt := range tests {
		t.Run(tt.agent, func(t *testing.T) {
			prompt := &CopilotPrompt{Agent: tt.agent}
			if got := prompt.Filename(); got != tt.want {
				t.Errorf("Filename() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Round-trip tests
func TestRoundTrip_ClaudeCommandToCopilotPromptToClaudeCommand(t *testing.T) {
	original := &ClaudeCommand{
		Name:        "commit",
		Description: "Create a git commit",
		Body:        "Create a well-formatted commit.",
	}

	// Claude -> Copilot
	copilotBytes, err := ConvertCommand(original, FormatCopilot)
	if err != nil {
		t.Fatalf("Convert to Copilot: %v", err)
	}

	copilot, err := ParseCopilotPrompt(copilotBytes)
	if err != nil {
		t.Fatalf("Parse Copilot: %v", err)
	}

	// Copilot -> Claude
	claudeBytes, err := ConvertCommand(copilot, FormatClaude)
	if err != nil {
		t.Fatalf("Convert to Claude: %v", err)
	}

	result, err := ParseClaudeCommand(claudeBytes)
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
	if strings.TrimSpace(result.Body) != strings.TrimSpace(original.Body) {
		t.Errorf("Body = %q, want %q", result.Body, original.Body)
	}
}
