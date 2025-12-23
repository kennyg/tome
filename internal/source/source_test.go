package source

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Source
		wantErr bool
	}{
		{
			name:  "simple owner/repo",
			input: "kennyg/tome",
			want: &Source{
				Type:     TypeGitHub,
				Host:     "github.com",
				Owner:    "kennyg",
				Repo:     "tome",
				Ref:      "main",
				Original: "kennyg/tome",
			},
		},
		{
			name:  "owner/repo with path",
			input: "kennyg/tome:skills/my-skill",
			want: &Source{
				Type:     TypeGitHub,
				Host:     "github.com",
				Owner:    "kennyg",
				Repo:     "tome",
				Path:     "skills/my-skill",
				Ref:      "main",
				Original: "kennyg/tome:skills/my-skill",
			},
		},
		{
			name:  "owner/repo with ref",
			input: "kennyg/tome@v1.0.0",
			want: &Source{
				Type:     TypeGitHub,
				Host:     "github.com",
				Owner:    "kennyg",
				Repo:     "tome",
				Ref:      "v1.0.0",
				Original: "kennyg/tome@v1.0.0",
			},
		},
		{
			name:  "owner/repo with path and ref",
			input: "kennyg/tome:skills/my-skill@develop",
			want: &Source{
				Type:     TypeGitHub,
				Host:     "github.com",
				Owner:    "kennyg",
				Repo:     "tome",
				Path:     "skills/my-skill",
				Ref:      "develop",
				Original: "kennyg/tome:skills/my-skill@develop",
			},
		},
		{
			name:  "repo with dots in name",
			input: "kennyg/my.repo.name",
			want: &Source{
				Type:     TypeGitHub,
				Host:     "github.com",
				Owner:    "kennyg",
				Repo:     "my.repo.name",
				Ref:      "main",
				Original: "kennyg/my.repo.name",
			},
		},
		{
			name:  "repo with underscores and dashes",
			input: "user_name/repo-name_test",
			want: &Source{
				Type:     TypeGitHub,
				Host:     "github.com",
				Owner:    "user_name",
				Repo:     "repo-name_test",
				Ref:      "main",
				Original: "user_name/repo-name_test",
			},
		},
		{
			name:  "raw githubusercontent URL",
			input: "https://raw.githubusercontent.com/kennyg/tome/main/README.md",
			want: &Source{
				Type:     TypeGitHub,
				Host:     "github.com",
				Owner:    "kennyg",
				Repo:     "tome",
				Path:     "README.md",
				Ref:      "main",
				URL:      "https://raw.githubusercontent.com/kennyg/tome/main/README.md",
				Original: "https://raw.githubusercontent.com/kennyg/tome/main/README.md",
			},
		},
		{
			name:  "github.com blob URL",
			input: "https://github.com/kennyg/tome/blob/main/skills/test/SKILL.md",
			want: &Source{
				Type:     TypeGitHub,
				Host:     "github.com",
				Owner:    "kennyg",
				Repo:     "tome",
				Path:     "skills/test/SKILL.md",
				Ref:      "main",
				URL:      "https://github.com/kennyg/tome/blob/main/skills/test/SKILL.md",
				Original: "https://github.com/kennyg/tome/blob/main/skills/test/SKILL.md",
			},
		},
		{
			name:  "github.com tree URL",
			input: "https://github.com/kennyg/tome/tree/develop/skills",
			want: &Source{
				Type:     TypeGitHub,
				Host:     "github.com",
				Owner:    "kennyg",
				Repo:     "tome",
				Path:     "skills",
				Ref:      "develop",
				URL:      "https://github.com/kennyg/tome/tree/develop/skills",
				Original: "https://github.com/kennyg/tome/tree/develop/skills",
			},
		},
		{
			name:  "generic URL",
			input: "https://example.com/skill.md",
			want: &Source{
				Type:     TypeURL,
				URL:      "https://example.com/skill.md",
				Original: "https://example.com/skill.md",
			},
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Type != tt.want.Type {
				t.Errorf("Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.Host != tt.want.Host {
				t.Errorf("Host = %v, want %v", got.Host, tt.want.Host)
			}
			if got.Owner != tt.want.Owner {
				t.Errorf("Owner = %v, want %v", got.Owner, tt.want.Owner)
			}
			if got.Repo != tt.want.Repo {
				t.Errorf("Repo = %v, want %v", got.Repo, tt.want.Repo)
			}
			if got.Path != tt.want.Path {
				t.Errorf("Path = %v, want %v", got.Path, tt.want.Path)
			}
			if got.Ref != tt.want.Ref {
				t.Errorf("Ref = %v, want %v", got.Ref, tt.want.Ref)
			}
			if got.Original != tt.want.Original {
				t.Errorf("Original = %v, want %v", got.Original, tt.want.Original)
			}
		})
	}
}

func TestParseLocalPath(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"dot path", "./local/skill"},
		{"absolute path", "/tmp/skill.md"},
		{"home path", "~/skills/test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}
			if got.Type != TypeLocal {
				t.Errorf("Type = %v, want %v", got.Type, TypeLocal)
			}
		})
	}
}

func TestSource_GitHubRawURL(t *testing.T) {
	tests := []struct {
		name   string
		source *Source
		path   string
		want   string
	}{
		{
			name: "public github no path",
			source: &Source{
				Type:  TypeGitHub,
				Host:  "github.com",
				Owner: "kennyg",
				Repo:  "tome",
				Ref:   "main",
			},
			path: "SKILL.md",
			want: "https://raw.githubusercontent.com/kennyg/tome/main/SKILL.md",
		},
		{
			name: "public github with source path",
			source: &Source{
				Type:  TypeGitHub,
				Host:  "github.com",
				Owner: "kennyg",
				Repo:  "tome",
				Path:  "skills/test",
				Ref:   "main",
			},
			path: "SKILL.md",
			want: "https://raw.githubusercontent.com/kennyg/tome/main/skills/test/SKILL.md",
		},
		{
			name: "public github use source path only",
			source: &Source{
				Type:  TypeGitHub,
				Host:  "github.com",
				Owner: "kennyg",
				Repo:  "tome",
				Path:  "skills/test/SKILL.md",
				Ref:   "v1.0.0",
			},
			path: "",
			want: "https://raw.githubusercontent.com/kennyg/tome/v1.0.0/skills/test/SKILL.md",
		},
		{
			name: "github enterprise",
			source: &Source{
				Type:  TypeGitHub,
				Host:  "github.company.com",
				Owner: "team",
				Repo:  "skills",
				Ref:   "main",
			},
			path: "SKILL.md",
			want: "https://github.company.com/team/skills/raw/main/SKILL.md",
		},
		{
			name: "non-github source returns empty",
			source: &Source{
				Type: TypeURL,
				URL:  "https://example.com/skill.md",
			},
			path: "anything",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.source.GitHubRawURL(tt.path)
			if got != tt.want {
				t.Errorf("GitHubRawURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSource_GitHubAPIURL(t *testing.T) {
	tests := []struct {
		name   string
		source *Source
		want   string
	}{
		{
			name: "public github root",
			source: &Source{
				Type:  TypeGitHub,
				Host:  "github.com",
				Owner: "kennyg",
				Repo:  "tome",
				Ref:   "main",
			},
			want: "https://api.github.com/repos/kennyg/tome/contents?ref=main",
		},
		{
			name: "public github with path",
			source: &Source{
				Type:  TypeGitHub,
				Host:  "github.com",
				Owner: "kennyg",
				Repo:  "tome",
				Path:  "skills",
				Ref:   "develop",
			},
			want: "https://api.github.com/repos/kennyg/tome/contents/skills?ref=develop",
		},
		{
			name: "github enterprise",
			source: &Source{
				Type:  TypeGitHub,
				Host:  "github.company.com",
				Owner: "team",
				Repo:  "skills",
				Ref:   "main",
			},
			want: "https://github.company.com/api/v3/repos/team/skills/contents?ref=main",
		},
		{
			name: "non-github source returns empty",
			source: &Source{
				Type: TypeLocal,
				Path: "/tmp/skill",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.source.GitHubAPIURL()
			if got != tt.want {
				t.Errorf("GitHubAPIURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSource_String(t *testing.T) {
	tests := []struct {
		name   string
		source *Source
		want   string
	}{
		{
			name: "github simple",
			source: &Source{
				Type:  TypeGitHub,
				Owner: "kennyg",
				Repo:  "tome",
				Ref:   "main",
			},
			want: "kennyg/tome",
		},
		{
			name: "github with path",
			source: &Source{
				Type:  TypeGitHub,
				Owner: "kennyg",
				Repo:  "tome",
				Path:  "skills/test",
				Ref:   "main",
			},
			want: "kennyg/tome:skills/test",
		},
		{
			name: "github with non-main ref",
			source: &Source{
				Type:  TypeGitHub,
				Owner: "kennyg",
				Repo:  "tome",
				Ref:   "develop",
			},
			want: "kennyg/tome@develop",
		},
		{
			name: "github with path and ref",
			source: &Source{
				Type:  TypeGitHub,
				Owner: "kennyg",
				Repo:  "tome",
				Path:  "skills",
				Ref:   "v1.0.0",
			},
			want: "kennyg/tome:skills@v1.0.0",
		},
		{
			name: "local path",
			source: &Source{
				Type: TypeLocal,
				Path: "/tmp/skill.md",
			},
			want: "/tmp/skill.md",
		},
		{
			name: "url",
			source: &Source{
				Type: TypeURL,
				URL:  "https://example.com/skill.md",
			},
			want: "https://example.com/skill.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.source.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSource_IsEnterprise(t *testing.T) {
	tests := []struct {
		name   string
		source *Source
		want   bool
	}{
		{
			name:   "public github.com",
			source: &Source{Host: "github.com"},
			want:   false,
		},
		{
			name:   "empty host (defaults to github.com)",
			source: &Source{Host: ""},
			want:   false,
		},
		{
			name:   "github enterprise",
			source: &Source{Host: "github.company.com"},
			want:   true,
		},
		{
			name:   "ghe host",
			source: &Source{Host: "ghe.company.com"},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.source.IsEnterprise()
			if got != tt.want {
				t.Errorf("IsEnterprise() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsGitHubHost(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{"github.com", true},
		{"raw.githubusercontent.com", true},
		{"github.company.com", true},
		{"git.company.com", true},
		{"ghe.company.com", true},
		{"raw.github.company.com", true},
		{"example.com", false},
		{"gitlab.com", false},
		{"bitbucket.org", false},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			got := isGitHubHost(tt.host)
			if got != tt.want {
				t.Errorf("isGitHubHost(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}
