package fetch

import (
	"testing"

	"github.com/kennyg/tome/internal/artifact"
)

func TestBase64Decode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "simple base64",
			input: "SGVsbG8gV29ybGQ=",
			want:  "Hello World",
		},
		{
			name:  "base64 with newlines",
			input: "SGVs\nbG8g\nV29y\nbGQ=",
			want:  "Hello World",
		},
		{
			name:  "base64 with escaped newlines",
			input: "SGVs\\nbG8g\\nV29y\\nbGQ=",
			want:  "Hello World",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:    "invalid base64",
			input:   "not-valid-base64!!!",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := base64Decode(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if string(got) != tt.want {
				t.Errorf("got %q, want %q", string(got), tt.want)
			}
		})
	}
}

func TestAppendPath(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		segment string
		want    string
	}{
		{
			name:    "simple append",
			baseURL: "https://api.github.com/repos/owner/repo/contents",
			segment: "skills",
			want:    "https://api.github.com/repos/owner/repo/contents/skills",
		},
		{
			name:    "with query string",
			baseURL: "https://api.github.com/repos/owner/repo/contents?ref=main",
			segment: "skills",
			want:    "https://api.github.com/repos/owner/repo/contents/skills?ref=main",
		},
		{
			name:    "nested path",
			baseURL: "https://example.com/path",
			segment: "sub/dir",
			want:    "https://example.com/path/sub/dir",
		},
		{
			name:    "empty segment",
			baseURL: "https://example.com/path",
			segment: "",
			want:    "https://example.com/path/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appendPath(tt.baseURL, tt.segment)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsScriptFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"script.py", true},
		{"script.sh", true},
		{"script.js", true},
		{"script.ts", true},
		{"script.rb", true},
		{"SCRIPT.PY", true},  // case insensitive
		{"readme.md", false},
		{"config.yaml", false},
		{"data.json", false},
		{"noextension", false},
		{"path/to/script.py", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := IsScriptFile(tt.path)
			if got != tt.want {
				t.Errorf("IsScriptFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestValidateIncludePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		// Valid paths
		{"valid md", "file.md", false},
		{"valid txt", "file.txt", false},
		{"valid json", "config.json", false},
		{"valid yaml", "config.yaml", false},
		{"valid yml", "config.yml", false},
		{"valid toml", "config.toml", false},
		{"valid tmpl", "template.tmpl", false},
		{"valid py", "script.py", false},
		{"valid sh", "script.sh", false},
		{"valid js", "script.js", false},
		{"valid ts", "script.ts", false},
		{"valid rb", "script.rb", false},
		{"nested valid", "path/to/file.md", false},

		// Invalid paths
		{"absolute path", "/etc/passwd", true},
		{"path traversal", "../secret.md", true},
		{"hidden traversal", "foo/../bar.md", true},
		{"disallowed exe", "binary.exe", true},
		{"disallowed go", "main.go", true},
		{"disallowed html", "page.html", true},
		{"no extension", "noext", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIncludePath(tt.path)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateIncludePath(%q) expected error, got nil", tt.path)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateIncludePath(%q) unexpected error: %v", tt.path, err)
			}
		})
	}
}

func TestDetectArtifactType(t *testing.T) {
	tests := []struct {
		filename string
		want     artifact.Type
	}{
		{"SKILL.md", artifact.TypeSkill},
		{"skill.md", artifact.TypeSkill},
		{"Skill.MD", artifact.TypeSkill},
		{"path/to/SKILL.md", artifact.TypeSkill},
		{"commit.md", artifact.TypeCommand},
		{"review-pr.md", artifact.TypeCommand},
		{"path/to/command.md", artifact.TypeCommand},
		{"readme.txt", ""},
		{"script.py", ""},
		{"noextension", ""},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := DetectArtifactType(tt.filename)
			if got != tt.want {
				t.Errorf("DetectArtifactType(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestIsArtifactFile(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		// Artifacts
		{"SKILL.md", true},
		{"skill.md", true},
		{"commit.md", true},
		{"review-pr.md", true},
		{"custom-command.md", true},

		// Non-artifacts (excluded files)
		{"README.md", false},
		{"readme.md", false},
		{"LICENSE.md", false},
		{"CHANGELOG.md", false},
		{"CONTRIBUTING.md", false},
		{"AGENTS.md", false},
		{"CLAUDE.md", false},

		// Non-markdown files
		{"script.py", false},
		{"config.yaml", false},
		{"noextension", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := IsArtifactFile(tt.filename)
			if got != tt.want {
				t.Errorf("IsArtifactFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestCommandNameFromFile(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"commit.md", "commit"},
		{"review-pr.md", "review-pr"},
		{"path/to/command.md", "command"},
		{"My Command.md", "My-Command"},
		{"special@chars!.md", "special-chars"},
		{".md", "unnamed"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := CommandNameFromFile(tt.filename)
			if got != tt.want {
				t.Errorf("CommandNameFromFile(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"simple", "simple"},
		{"with-dash", "with-dash"},
		{"with_underscore", "with_underscore"},
		{"UPPERCASE", "UPPERCASE"},
		{"MixedCase123", "MixedCase123"},
		{"spaces here", "spaces-here"},
		{"special@#$chars", "special-chars"},
		{"multiple---dashes", "multiple-dashes"},
		{"-leading-dash", "leading-dash"},
		{"trailing-dash-", "trailing-dash"},
		{"-both-sides-", "both-sides"},
		{"", "unnamed"},
		{"@#$%", "unnamed"},
		{"hello world!", "hello-world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeFilename(tt.name)
			if got != tt.want {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestParseFrontmatter(t *testing.T) {
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
			name: "full frontmatter",
			content: `---
name: test-skill
description: A test skill
version: 1.0.0
---
# Body content`,
			wantName:    "test-skill",
			wantDesc:    "A test skill",
			wantVersion: "1.0.0",
			wantBody:    "# Body content",
		},
		{
			name: "frontmatter with includes",
			content: `---
name: skill-with-includes
includes:
  - helper.py
  - config.yaml
---
Body here`,
			wantName: "skill-with-includes",
			wantBody: "Body here",
		},
		{
			name:     "no frontmatter",
			content:  "# Just a heading\n\nSome content",
			wantBody: "# Just a heading\n\nSome content",
		},
		{
			name:     "unclosed frontmatter",
			content:  "---\nname: test\nNo closing delimiter",
			wantBody: "---\nname: test\nNo closing delimiter",
		},
		{
			name:    "empty content",
			content: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body, err := parseFrontmatter([]byte(tt.content))
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if fm.Name != tt.wantName {
				t.Errorf("name = %q, want %q", fm.Name, tt.wantName)
			}
			if fm.Description != tt.wantDesc {
				t.Errorf("description = %q, want %q", fm.Description, tt.wantDesc)
			}
			if fm.Version != tt.wantVersion {
				t.Errorf("version = %q, want %q", fm.Version, tt.wantVersion)
			}
			if body != tt.wantBody {
				t.Errorf("body = %q, want %q", body, tt.wantBody)
			}
		})
	}
}

func TestExtractNameFromContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "H1 heading",
			content: "# My Skill Name\n\nSome content",
			want:    "My Skill Name",
		},
		{
			name:    "H1 with leading whitespace",
			content: "  # Trimmed Heading\n\nContent",
			want:    "Trimmed Heading",
		},
		{
			name:    "multiple headings",
			content: "# First Heading\n## Second\n# Third",
			want:    "First Heading",
		},
		{
			name:    "no heading",
			content: "Just some content\nwithout any heading",
			want:    "unnamed-skill",
		},
		{
			name:    "empty content",
			content: "",
			want:    "unnamed-skill",
		},
		{
			name:    "H2 only",
			content: "## Not H1\n\nContent",
			want:    "unnamed-skill",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractNameFromContent(tt.content)
			if got != tt.want {
				t.Errorf("extractNameFromContent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractDescriptionFromContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "paragraph after heading",
			content: "# Heading\n\nThis is the description paragraph.",
			want:    "This is the description paragraph.",
		},
		{
			name:    "skips bullets",
			content: "- bullet point\n* another bullet\n\nActual description here",
			want:    "Actual description here",
		},
		{
			name:    "skips numbered lists",
			content: "1. First item\n2. Second item\n\nDescription text",
			want:    "Description text",
		},
		{
			name:    "skips blockquotes",
			content: "> Quote here\n\nDescription after quote",
			want:    "Description after quote",
		},
		{
			name:    "skips code block delimiters",
			content: "```python\n\nDescription after code",
			want:    "Description after code",
		},
		{
			name:    "skips horizontal rules",
			content: "---\n\nDescription after rule",
			want:    "Description after rule",
		},
		{
			name:    "truncates long descriptions",
			content: "# Heading\n\n" + string(make([]byte, 250)),
			want:    string(make([]byte, 200)) + "...",
		},
		{
			name:    "empty content",
			content: "",
			want:    "",
		},
		{
			name:    "only headings and bullets",
			content: "# Heading\n- bullet\n* bullet\n## Subheading",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDescriptionFromContent(tt.content)
			if got != tt.want {
				t.Errorf("extractDescriptionFromContent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseSkill(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		sourceURL string
		wantName  string
		wantDesc  string
		wantErr   bool
	}{
		{
			name: "skill with frontmatter",
			content: `---
name: my-skill
description: Does something cool
version: 1.0.0
---
# My Skill

This skill helps with things.`,
			sourceURL: "https://github.com/owner/repo",
			wantName:  "my-skill",
			wantDesc:  "Does something cool",
		},
		{
			name: "skill without frontmatter",
			content: `# Auto Named Skill

This skill has auto-extracted name and description.`,
			sourceURL: "https://github.com/owner/repo",
			wantName:  "Auto Named Skill",
			wantDesc:  "This skill has auto-extracted name and description.",
		},
		{
			name: "skill with invalid include",
			content: `---
name: bad-skill
includes:
  - ../../../etc/passwd
---
Content`,
			sourceURL: "https://github.com/owner/repo",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			art, err := ParseSkill([]byte(tt.content), tt.sourceURL)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if art.Name != tt.wantName {
				t.Errorf("name = %q, want %q", art.Name, tt.wantName)
			}
			if art.Description != tt.wantDesc {
				t.Errorf("description = %q, want %q", art.Description, tt.wantDesc)
			}
			if art.Type != artifact.TypeSkill {
				t.Errorf("type = %q, want %q", art.Type, artifact.TypeSkill)
			}
			if art.SourceURL != tt.sourceURL {
				t.Errorf("sourceURL = %q, want %q", art.SourceURL, tt.sourceURL)
			}
		})
	}
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		filename  string
		sourceURL string
		wantName  string
		wantDesc  string
	}{
		{
			name: "command with frontmatter",
			content: `---
name: custom-name
description: Custom description
---
# Command Content`,
			filename:  "commit.md",
			sourceURL: "https://github.com/owner/repo",
			wantName:  "custom-name",
			wantDesc:  "Custom description",
		},
		{
			name: "command without frontmatter",
			content: `# Review PR

This command helps review PRs.`,
			filename:  "review-pr.md",
			sourceURL: "https://github.com/owner/repo",
			wantName:  "review-pr",
			wantDesc:  "This command helps review PRs.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			art, err := ParseCommand([]byte(tt.content), tt.filename, tt.sourceURL)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if art.Name != tt.wantName {
				t.Errorf("name = %q, want %q", art.Name, tt.wantName)
			}
			if art.Description != tt.wantDesc {
				t.Errorf("description = %q, want %q", art.Description, tt.wantDesc)
			}
			if art.Type != artifact.TypeCommand {
				t.Errorf("type = %q, want %q", art.Type, artifact.TypeCommand)
			}
			if art.Filename != tt.filename {
				t.Errorf("filename = %q, want %q", art.Filename, tt.filename)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
	if client.http == nil {
		t.Error("client.http is nil")
	}
	if client.gh == nil {
		t.Error("client.gh is nil")
	}
}

func TestGitHubContent(t *testing.T) {
	content := GitHubContent{
		Name:        "SKILL.md",
		Path:        "skills/my-skill/SKILL.md",
		Type:        "file",
		DownloadURL: "https://raw.githubusercontent.com/owner/repo/main/skills/my-skill/SKILL.md",
		SkillDir:    "skills/my-skill",
	}

	if content.Name != "SKILL.md" {
		t.Errorf("Name = %q, want SKILL.md", content.Name)
	}
	if content.Type != "file" {
		t.Errorf("Type = %q, want file", content.Type)
	}
}

func TestFrontmatter(t *testing.T) {
	fm := Frontmatter{
		Name:         "test-skill",
		Description:  "A test",
		Version:      "1.0.0",
		Author:       "tester",
		License:      "MIT",
		Globs:        []string{"*.go", "*.py"},
		Includes:     []string{"helper.py"},
		AllowedTools: []string{"Bash", "Read"},
	}

	if fm.Name != "test-skill" {
		t.Errorf("Name = %q", fm.Name)
	}
	if len(fm.Globs) != 2 {
		t.Errorf("Globs len = %d, want 2", len(fm.Globs))
	}
	if len(fm.AllowedTools) != 2 {
		t.Errorf("AllowedTools len = %d, want 2", len(fm.AllowedTools))
	}
}

func TestIncludedFile(t *testing.T) {
	file := IncludedFile{
		Path:    "helper.py",
		Content: []byte("print('hello')"),
	}

	if file.Path != "helper.py" {
		t.Errorf("Path = %q", file.Path)
	}
	if string(file.Content) != "print('hello')" {
		t.Errorf("Content = %q", string(file.Content))
	}
}

func TestConstants(t *testing.T) {
	if MaxIncludeFileSize != 100*1024 {
		t.Errorf("MaxIncludeFileSize = %d, want %d", MaxIncludeFileSize, 100*1024)
	}
	if MaxTotalIncludeSize != 1024*1024 {
		t.Errorf("MaxTotalIncludeSize = %d, want %d", MaxTotalIncludeSize, 1024*1024)
	}
}
