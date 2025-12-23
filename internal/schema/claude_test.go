package schema

import (
	"strings"
	"testing"
)

func TestParseClaudeSkill(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantName    string
		wantDesc    string
		wantVersion string
		wantAuthor  string
		wantGlobs   []string
		wantBody    string
		wantErr     bool
	}{
		{
			name: "full skill",
			content: `---
name: pdf-helper
description: Helps with PDF manipulation
version: 1.0.0
author: kennyg
globs:
  - "*.pdf"
includes:
  - helper.py
---
# PDF Helper

Use this skill for PDF tasks.`,
			wantName:    "pdf-helper",
			wantDesc:    "Helps with PDF manipulation",
			wantVersion: "1.0.0",
			wantAuthor:  "kennyg",
			wantGlobs:   []string{"*.pdf"},
			wantBody:    "# PDF Helper\n\nUse this skill for PDF tasks.",
		},
		{
			name: "minimal skill",
			content: `---
name: test
description: A test
---
Body`,
			wantName: "test",
			wantDesc: "A test",
			wantBody: "Body",
		},
		{
			name: "skill with allowed-tools",
			content: `---
name: builder
description: Build helper
allowed-tools:
  - Bash
  - Read
---
Content`,
			wantName: "builder",
			wantDesc: "Build helper",
			wantBody: "Content",
		},
		{
			name:     "no frontmatter",
			content:  "Just markdown",
			wantBody: "Just markdown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skill, err := ParseClaudeSkill([]byte(tt.content))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseClaudeSkill() error = %v, wantErr %v", err, tt.wantErr)
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
			if skill.Version != tt.wantVersion {
				t.Errorf("Version = %q, want %q", skill.Version, tt.wantVersion)
			}
			if skill.Author != tt.wantAuthor {
				t.Errorf("Author = %q, want %q", skill.Author, tt.wantAuthor)
			}
			if skill.Body != tt.wantBody {
				t.Errorf("Body = %q, want %q", skill.Body, tt.wantBody)
			}
			if skill.GetFormat() != FormatClaude {
				t.Errorf("GetFormat() = %v, want %v", skill.GetFormat(), FormatClaude)
			}
		})
	}
}

func TestParseOpenCodeSkill(t *testing.T) {
	content := `---
name: test-skill
description: OpenCode test
---
Body`

	skill, err := ParseOpenCodeSkill([]byte(content))
	if err != nil {
		t.Fatalf("ParseOpenCodeSkill() error = %v", err)
	}

	if skill.GetFormat() != FormatOpenCode {
		t.Errorf("GetFormat() = %v, want %v", skill.GetFormat(), FormatOpenCode)
	}
	if skill.Name != "test-skill" {
		t.Errorf("Name = %q, want %q", skill.Name, "test-skill")
	}
}

func TestClaudeSkill_Serialize(t *testing.T) {
	skill := &ClaudeSkill{
		Name:        "test-skill",
		Description: "A test skill",
		Version:     "1.0.0",
		Author:      "tester",
		Globs:       []string{"*.go"},
		Body:        "# Content\n\nBody text.",
	}

	got, err := skill.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	content := string(got)

	// Check frontmatter markers
	if !strings.HasPrefix(content, "---\n") {
		t.Error("Should start with ---")
	}
	if !strings.Contains(content, "\n---\n") {
		t.Error("Should have closing ---")
	}

	// Check fields are present
	checks := []string{
		"name: test-skill",
		"description: A test skill",
		"version: 1.0.0",
		"author: tester",
		"globs:",
		"*.go", // YAML serialization may vary
		"# Content",
		"Body text.",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("Serialized content should contain %q", check)
		}
	}
}

func TestClaudeSkill_Interface(t *testing.T) {
	skill := &ClaudeSkill{
		Name:        "interface-test",
		Description: "Testing interface",
		Body:        "Body content",
	}

	// Test interface methods
	if skill.GetName() != "interface-test" {
		t.Errorf("GetName() = %q, want %q", skill.GetName(), "interface-test")
	}
	if skill.GetDescription() != "Testing interface" {
		t.Errorf("GetDescription() = %q", skill.GetDescription())
	}
	if skill.GetBody() != "Body content" {
		t.Errorf("GetBody() = %q", skill.GetBody())
	}
	if skill.GetFormat() != FormatClaude {
		t.Errorf("GetFormat() = %v", skill.GetFormat())
	}

	// Test format setter
	skill.SetFormat(FormatOpenCode)
	if skill.GetFormat() != FormatOpenCode {
		t.Errorf("After SetFormat(), GetFormat() = %v", skill.GetFormat())
	}
}

func TestClaudeSkill_Metadata(t *testing.T) {
	skill := &ClaudeSkill{
		Name:        "meta-test",
		Description: "Metadata test",
		Version:     "2.0.0",
		Author:      "author",
		Body:        "Body",
	}

	// Test ToMetadata
	meta := skill.ToMetadata()
	if meta.Name != "meta-test" {
		t.Errorf("ToMetadata().Name = %q", meta.Name)
	}
	if meta.Version != "2.0.0" {
		t.Errorf("ToMetadata().Version = %q", meta.Version)
	}

	// Test FromMetadata
	newSkill := &ClaudeSkill{}
	newSkill.FromMetadata(SkillMetadata{
		Name:        "from-meta",
		Description: "From metadata",
		Version:     "3.0.0",
		Author:      "new-author",
		Body:        "New body",
	})

	if newSkill.Name != "from-meta" {
		t.Errorf("FromMetadata Name = %q", newSkill.Name)
	}
	if newSkill.Body != "New body" {
		t.Errorf("FromMetadata Body = %q", newSkill.Body)
	}
}
