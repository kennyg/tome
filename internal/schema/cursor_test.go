package schema

import (
	"strings"
	"testing"
)

func TestParseCursorSkill(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantName string
		wantDesc string
		wantBody string
		wantErr  bool
	}{
		{
			name: "with frontmatter",
			content: `---
name: coding-rules
description: Rules for coding
---
# Coding Rules

Always use meaningful names.`,
			wantName: "coding-rules",
			wantDesc: "Rules for coding",
			wantBody: "# Coding Rules\n\nAlways use meaningful names.",
		},
		{
			name: "no frontmatter",
			content: `# Simple Rules

Just body content here.`,
			wantName: "",
			wantDesc: "",
			wantBody: "# Simple Rules\n\nJust body content here.",
		},
		{
			name: "minimal frontmatter",
			content: `---
name: test
---
Body`,
			wantName: "test",
			wantBody: "Body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skill, err := ParseCursorSkill([]byte(tt.content))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCursorSkill() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if skill.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", skill.Name, tt.wantName)
			}
			if skill.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", skill.Description, tt.wantDesc)
			}
			if skill.Body != tt.wantBody {
				t.Errorf("Body = %q, want %q", skill.Body, tt.wantBody)
			}
			if skill.GetFormat() != FormatCursor {
				t.Errorf("GetFormat() = %v, want %v", skill.GetFormat(), FormatCursor)
			}
		})
	}
}

func TestCursorSkill_Serialize(t *testing.T) {
	tests := []struct {
		name       string
		skill      *CursorSkill
		wantFM     bool // should have frontmatter
		wantChecks []string
	}{
		{
			name: "with metadata",
			skill: &CursorSkill{
				Name:        "test-skill",
				Description: "A test",
				Body:        "# Content",
			},
			wantFM: true,
			wantChecks: []string{
				"name: test-skill",
				"description: A test",
				"# Content",
			},
		},
		{
			name: "body only",
			skill: &CursorSkill{
				Body: "# Just body content",
			},
			wantFM: false,
			wantChecks: []string{
				"# Just body content",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.skill.Serialize()
			if err != nil {
				t.Fatalf("Serialize() error = %v", err)
			}

			content := string(got)

			// Check frontmatter presence
			hasFM := strings.HasPrefix(content, "---")
			if hasFM != tt.wantFM {
				t.Errorf("Frontmatter presence = %v, want %v", hasFM, tt.wantFM)
			}

			// Check expected content
			for _, check := range tt.wantChecks {
				if !strings.Contains(content, check) {
					t.Errorf("Serialized content should contain %q", check)
				}
			}
		})
	}
}

func TestCursorSkill_Filename(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Coding Rules", "coding-rules.md"},
		{"test", "test.md"},
		{"", "rules.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skill := &CursorSkill{Name: tt.name}
			if got := skill.Filename(); got != tt.want {
				t.Errorf("Filename() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsCursorFile(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{".cursor/rules/coding.md", true},
		{"project/.cursor/settings.md", true},
		{"cursor/config.md", true},
		{".claude/skills/test.md", false},
		{"random.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			if got := IsCursorFile(tt.filename); got != tt.want {
				t.Errorf("IsCursorFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestCursorSkill_Interface(t *testing.T) {
	skill := &CursorSkill{
		Name:        "interface-test",
		Description: "Testing",
		Body:        "Content",
	}

	// Verify it implements Skill
	var _ Skill = skill

	if skill.GetName() != "interface-test" {
		t.Errorf("GetName() = %q", skill.GetName())
	}
	if skill.GetDescription() != "Testing" {
		t.Errorf("GetDescription() = %q", skill.GetDescription())
	}
	if skill.GetBody() != "Content" {
		t.Errorf("GetBody() = %q", skill.GetBody())
	}
	if skill.GetFormat() != FormatCursor {
		t.Errorf("GetFormat() = %v", skill.GetFormat())
	}
}

func TestCursorSkill_Metadata(t *testing.T) {
	skill := &CursorSkill{
		Name:        "meta-test",
		Description: "Metadata test",
		Body:        "Body",
	}

	// Test ToMetadata
	meta := skill.ToMetadata()
	if meta.Name != "meta-test" {
		t.Errorf("ToMetadata().Name = %q", meta.Name)
	}

	// Test FromMetadata
	newSkill := &CursorSkill{}
	newSkill.FromMetadata(SkillMetadata{
		Name:        "from-meta",
		Description: "From metadata",
		Body:        "New body",
	})

	if newSkill.Name != "from-meta" {
		t.Errorf("FromMetadata Name = %q", newSkill.Name)
	}
	if newSkill.Body != "New body" {
		t.Errorf("FromMetadata Body = %q", newSkill.Body)
	}
}
