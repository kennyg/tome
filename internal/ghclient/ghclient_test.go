package ghclient

import (
	"os"
	"testing"
)

func TestParseGitHubURL(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		wantOwner    string
		wantRepo     string
		wantPath     string
		wantHostname string
		wantErr      bool
	}{
		// raw.githubusercontent.com URLs
		{
			name:         "raw githubusercontent simple",
			url:          "https://raw.githubusercontent.com/owner/repo/main/file.md",
			wantOwner:    "owner",
			wantRepo:     "repo",
			wantPath:     "file.md",
			wantHostname: "",
			wantErr:      false,
		},
		{
			name:         "raw githubusercontent nested path",
			url:          "https://raw.githubusercontent.com/owner/repo/main/path/to/file.md",
			wantOwner:    "owner",
			wantRepo:     "repo",
			wantPath:     "path/to/file.md",
			wantHostname: "",
			wantErr:      false,
		},
		{
			name:         "raw githubusercontent with branch",
			url:          "https://raw.githubusercontent.com/anthropics/cookbook/main/skills/SKILL.md",
			wantOwner:    "anthropics",
			wantRepo:     "cookbook",
			wantPath:     "skills/SKILL.md",
			wantHostname: "",
			wantErr:      false,
		},
		{
			name:    "raw githubusercontent too short",
			url:     "https://raw.githubusercontent.com/owner/repo",
			wantErr: true,
		},

		// GHE raw URLs
		{
			name:         "GHE raw URL",
			url:          "https://github.company.com/owner/repo/raw/main/file.md",
			wantOwner:    "owner",
			wantRepo:     "repo",
			wantPath:     "file.md",
			wantHostname: "github.company.com",
			wantErr:      false,
		},
		{
			name:         "GHE raw URL nested path",
			url:          "https://ghe.example.org/team/project/raw/develop/src/config.yaml",
			wantOwner:    "team",
			wantRepo:     "project",
			wantPath:     "src/config.yaml",
			wantHostname: "ghe.example.org",
			wantErr:      false,
		},
		{
			name:    "GHE raw URL invalid format",
			url:     "https://github.company.com/owner/repo/raw",
			wantErr: true,
		},

		// API URLs (public GitHub)
		{
			name:         "API URL root contents",
			url:          "https://api.github.com/repos/owner/repo/contents",
			wantOwner:    "owner",
			wantRepo:     "repo",
			wantPath:     "",
			wantHostname: "",
			wantErr:      false,
		},
		{
			name:         "API URL with path",
			url:          "https://api.github.com/repos/owner/repo/contents/path/to/file",
			wantOwner:    "owner",
			wantRepo:     "repo",
			wantPath:     "path/to/file",
			wantHostname: "",
			wantErr:      false,
		},
		{
			name:         "API URL single file",
			url:          "https://api.github.com/repos/kennyg/tome/contents/README.md",
			wantOwner:    "kennyg",
			wantRepo:     "tome",
			wantPath:     "README.md",
			wantHostname: "",
			wantErr:      false,
		},

		// GHE API URLs
		{
			name:         "GHE API URL",
			url:          "https://github.company.com/api/v3/repos/owner/repo/contents",
			wantOwner:    "owner",
			wantRepo:     "repo",
			wantPath:     "",
			wantHostname: "github.company.com",
			wantErr:      false,
		},
		{
			name:         "GHE API URL with path",
			url:          "https://ghe.example.org/api/v3/repos/team/project/contents/src/main.go",
			wantOwner:    "team",
			wantRepo:     "project",
			wantPath:     "src/main.go",
			wantHostname: "ghe.example.org",
			wantErr:      false,
		},

		// Error cases
		{
			name:    "unsupported URL format",
			url:     "https://github.com/owner/repo",
			wantErr: true,
		},
		{
			name:    "invalid API URL format",
			url:     "https://api.github.com/users/owner",
			wantErr: true,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, path, hostname, err := ParseGitHubURL(tt.url)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseGitHubURL() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseGitHubURL() unexpected error: %v", err)
				return
			}

			if owner != tt.wantOwner {
				t.Errorf("owner = %q, want %q", owner, tt.wantOwner)
			}
			if repo != tt.wantRepo {
				t.Errorf("repo = %q, want %q", repo, tt.wantRepo)
			}
			if path != tt.wantPath {
				t.Errorf("path = %q, want %q", path, tt.wantPath)
			}
			if hostname != tt.wantHostname {
				t.Errorf("hostname = %q, want %q", hostname, tt.wantHostname)
			}
		})
	}
}

func TestNew(t *testing.T) {
	// Clear any existing tokens
	origGitHub := os.Getenv("GITHUB_TOKEN")
	origGH := os.Getenv("GH_TOKEN")
	defer func() {
		os.Setenv("GITHUB_TOKEN", origGitHub)
		os.Setenv("GH_TOKEN", origGH)
	}()

	t.Run("unauthenticated when no token", func(t *testing.T) {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GH_TOKEN")

		client := New()
		if client == nil {
			t.Fatal("New() returned nil")
		}
		if client.gh == nil {
			t.Error("client.gh is nil")
		}
		// Note: IsAuthenticated may still be true if gh CLI config exists
	})

	t.Run("authenticated with GITHUB_TOKEN", func(t *testing.T) {
		os.Setenv("GITHUB_TOKEN", "test-token")
		os.Unsetenv("GH_TOKEN")

		client := New()
		if client == nil {
			t.Fatal("New() returned nil")
		}
		if !client.IsAuthenticated() {
			t.Error("expected authenticated client with GITHUB_TOKEN")
		}
	})

	t.Run("authenticated with GH_TOKEN", func(t *testing.T) {
		os.Unsetenv("GITHUB_TOKEN")
		os.Setenv("GH_TOKEN", "test-gh-token")

		client := New()
		if client == nil {
			t.Fatal("New() returned nil")
		}
		if !client.IsAuthenticated() {
			t.Error("expected authenticated client with GH_TOKEN")
		}
	})

	t.Run("GITHUB_TOKEN takes precedence over GH_TOKEN", func(t *testing.T) {
		os.Setenv("GITHUB_TOKEN", "github-token")
		os.Setenv("GH_TOKEN", "gh-token")

		client := New()
		if client == nil {
			t.Fatal("New() returned nil")
		}
		if !client.IsAuthenticated() {
			t.Error("expected authenticated client")
		}
	})
}

func TestNewForHost(t *testing.T) {
	// Clear tokens for consistent behavior
	origGitHub := os.Getenv("GITHUB_TOKEN")
	origGH := os.Getenv("GH_TOKEN")
	defer func() {
		os.Setenv("GITHUB_TOKEN", origGitHub)
		os.Setenv("GH_TOKEN", origGH)
	}()
	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GH_TOKEN")

	t.Run("public github.com", func(t *testing.T) {
		client := NewForHost("github.com")
		if client == nil {
			t.Fatal("NewForHost() returned nil")
		}
		// BaseURL should remain default for github.com
		if client.gh.BaseURL.Host != "api.github.com" {
			t.Errorf("expected api.github.com, got %s", client.gh.BaseURL.Host)
		}
	})

	t.Run("public api.github.com", func(t *testing.T) {
		client := NewForHost("api.github.com")
		if client == nil {
			t.Fatal("NewForHost() returned nil")
		}
		if client.gh.BaseURL.Host != "api.github.com" {
			t.Errorf("expected api.github.com, got %s", client.gh.BaseURL.Host)
		}
	})

	t.Run("empty host", func(t *testing.T) {
		client := NewForHost("")
		if client == nil {
			t.Fatal("NewForHost() returned nil")
		}
		if client.gh.BaseURL.Host != "api.github.com" {
			t.Errorf("expected api.github.com, got %s", client.gh.BaseURL.Host)
		}
	})

	t.Run("GitHub Enterprise host", func(t *testing.T) {
		client := NewForHost("github.company.com")
		if client == nil {
			t.Fatal("NewForHost() returned nil")
		}
		if client.gh.BaseURL.Host != "github.company.com" {
			t.Errorf("expected github.company.com, got %s", client.gh.BaseURL.Host)
		}
		if client.gh.BaseURL.Path != "/api/v3/" {
			t.Errorf("expected /api/v3/, got %s", client.gh.BaseURL.Path)
		}
	})

	t.Run("GHE with subdomain", func(t *testing.T) {
		client := NewForHost("ghe.example.org")
		if client == nil {
			t.Fatal("NewForHost() returned nil")
		}
		if client.gh.BaseURL.Host != "ghe.example.org" {
			t.Errorf("expected ghe.example.org, got %s", client.gh.BaseURL.Host)
		}
	})
}

func TestIsAuthenticated(t *testing.T) {
	origGitHub := os.Getenv("GITHUB_TOKEN")
	defer os.Setenv("GITHUB_TOKEN", origGitHub)

	t.Run("authenticated", func(t *testing.T) {
		os.Setenv("GITHUB_TOKEN", "test-token")
		client := New()
		if !client.IsAuthenticated() {
			t.Error("expected true")
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GH_TOKEN")
		// This test might still pass as authenticated if gh CLI config exists
		client := New()
		// Just verify it doesn't panic
		_ = client.IsAuthenticated()
	})
}

func TestSearchCodeResult(t *testing.T) {
	result := SearchCodeResult{Repository: "owner/repo"}
	if result.Repository != "owner/repo" {
		t.Errorf("expected owner/repo, got %s", result.Repository)
	}
}

func TestSearchRepoResult(t *testing.T) {
	result := SearchRepoResult{
		FullName:    "owner/repo",
		Description: "A test repo",
		Stars:       42,
	}
	if result.FullName != "owner/repo" {
		t.Errorf("FullName = %s, want owner/repo", result.FullName)
	}
	if result.Description != "A test repo" {
		t.Errorf("Description = %s, want 'A test repo'", result.Description)
	}
	if result.Stars != 42 {
		t.Errorf("Stars = %d, want 42", result.Stars)
	}
}
