package schema

import (
	"reflect"
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantFM   map[string]interface{}
		wantBody string
		wantErr  bool
	}{
		{
			name: "full frontmatter",
			content: `---
name: test-skill
description: A test skill
version: 1.0.0
---
# Body content

Some text here.`,
			wantFM: map[string]interface{}{
				"name":        "test-skill",
				"description": "A test skill",
				"version":     "1.0.0",
			},
			wantBody: "# Body content\n\nSome text here.",
		},
		{
			name: "no frontmatter",
			content: `# Just a markdown file

No frontmatter here.`,
			wantFM:   map[string]interface{}{},
			wantBody: "# Just a markdown file\n\nNo frontmatter here.",
		},
		{
			name: "empty frontmatter",
			content: `---
---
Body only.`,
			wantFM:   map[string]interface{}{},
			wantBody: "---\n---\nBody only.", // Implementation treats empty frontmatter as no frontmatter
		},
		{
			name: "frontmatter with list",
			content: `---
name: skill
globs:
  - "*.go"
  - "*.md"
---
Content`,
			wantFM: map[string]interface{}{
				"name": "skill",
				"globs": []interface{}{
					"*.go",
					"*.md",
				},
			},
			wantBody: "Content",
		},
		{
			name:     "empty content",
			content:  "",
			wantFM:   map[string]interface{}{},
			wantBody: "",
		},
		{
			name: "unclosed frontmatter",
			content: `---
name: test
No closing delimiter`,
			wantFM:   map[string]interface{}{},
			wantBody: "---\nname: test\nNo closing delimiter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFM, gotBody, err := ParseFrontmatter([]byte(tt.content))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFrontmatter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFM, tt.wantFM) {
				t.Errorf("ParseFrontmatter() FM = %v, want %v", gotFM, tt.wantFM)
			}
			if gotBody != tt.wantBody {
				t.Errorf("ParseFrontmatter() body = %q, want %q", gotBody, tt.wantBody)
			}
		})
	}
}

func TestParseFrontmatterTyped(t *testing.T) {
	type testStruct struct {
		Name        string   `yaml:"name"`
		Description string   `yaml:"description"`
		Tags        []string `yaml:"tags,omitempty"`
	}

	tests := []struct {
		name     string
		content  string
		want     testStruct
		wantBody string
		wantErr  bool
	}{
		{
			name: "basic struct",
			content: `---
name: test
description: A description
---
Body here`,
			want: testStruct{
				Name:        "test",
				Description: "A description",
			},
			wantBody: "Body here",
		},
		{
			name: "struct with slice",
			content: `---
name: test
tags:
  - go
  - cli
---
Content`,
			want: testStruct{
				Name: "test",
				Tags: []string{"go", "cli"},
			},
			wantBody: "Content",
		},
		{
			name:     "no frontmatter",
			content:  "Just body",
			want:     testStruct{},
			wantBody: "Just body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got testStruct
			gotBody, err := ParseFrontmatterTyped([]byte(tt.content), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFrontmatterTyped() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseFrontmatterTyped() = %+v, want %+v", got, tt.want)
			}
			if gotBody != tt.wantBody {
				t.Errorf("ParseFrontmatterTyped() body = %q, want %q", gotBody, tt.wantBody)
			}
		})
	}
}

func TestSerializeFrontmatter(t *testing.T) {
	tests := []struct {
		name    string
		fm      interface{}
		body    string
		wantErr bool
	}{
		{
			name: "basic map",
			fm: map[string]string{
				"name":        "test",
				"description": "A test",
			},
			body: "# Content\n\nSome text.",
		},
		{
			name: "struct",
			fm: struct {
				Name string `yaml:"name"`
			}{Name: "test"},
			body: "Body",
		},
		{
			name: "empty body",
			fm: map[string]string{
				"name": "test",
			},
			body: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SerializeFrontmatter(tt.fm, tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("SerializeFrontmatter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// Verify it starts with ---
			if len(got) < 3 || string(got[:3]) != "---" {
				t.Errorf("SerializeFrontmatter() should start with ---")
			}

			// Verify it contains the body if provided
			if tt.body != "" && !contains(string(got), tt.body) {
				t.Errorf("SerializeFrontmatter() should contain body")
			}
		})
	}
}

func TestGetString(t *testing.T) {
	m := map[string]interface{}{
		"name":   "test",
		"number": 42,
		"nil":    nil,
	}

	tests := []struct {
		key  string
		want string
	}{
		{"name", "test"},
		{"number", ""},
		{"nil", ""},
		{"missing", ""},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := getString(m, tt.key); got != tt.want {
				t.Errorf("getString(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestGetStringSlice(t *testing.T) {
	m := map[string]interface{}{
		"strings":    []string{"a", "b"},
		"interfaces": []interface{}{"x", "y"},
		"mixed":      []interface{}{"z", 42},
		"single":     "not a slice",
	}

	tests := []struct {
		key  string
		want []string
	}{
		{"strings", []string{"a", "b"}},
		{"interfaces", []string{"x", "y"}},
		{"mixed", []string{"z"}}, // 42 is skipped
		{"single", nil},
		{"missing", nil},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := getStringSlice(m, tt.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getStringSlice(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}
