// Package ghclient provides a GitHub API client using go-github
package ghclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v67/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

// Client wraps the go-github client
type Client struct {
	gh            *github.Client
	authenticated bool
}

// New creates a new GitHub client
// Token resolution order: GITHUB_TOKEN, GH_TOKEN, gh CLI config, unauthenticated
func New() *Client {
	token := getToken()

	var httpClient *http.Client
	authenticated := false

	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		httpClient = oauth2.NewClient(context.Background(), ts)
		authenticated = true
	}

	return &Client{
		gh:            github.NewClient(httpClient),
		authenticated: authenticated,
	}
}

// NewForHost creates a GitHub client for a specific host (GitHub Enterprise)
func NewForHost(host string) *Client {
	c := New()

	// Configure for GHE if not github.com
	if host != "" && host != "github.com" && host != "api.github.com" {
		baseURL := fmt.Sprintf("https://%s/api/v3/", host)
		c.gh.BaseURL, _ = url.Parse(baseURL)
		uploadURL := fmt.Sprintf("https://%s/api/uploads/", host)
		c.gh.UploadURL, _ = url.Parse(uploadURL)
	}

	return c
}

// IsAuthenticated returns true if the client has a token
func (c *Client) IsAuthenticated() bool {
	return c.authenticated
}

// GetContents fetches a file's content from a repository
func (c *Client) GetContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) ([]byte, error) {
	fileContent, _, _, err := c.gh.Repositories.GetContents(ctx, owner, repo, path, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get contents: %w", err)
	}

	if fileContent == nil {
		return nil, fmt.Errorf("path is a directory, not a file")
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return nil, fmt.Errorf("failed to decode content: %w", err)
	}

	return []byte(content), nil
}

// ListContents lists directory contents in a repository
func (c *Client) ListContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) ([]*github.RepositoryContent, error) {
	_, dirContents, _, err := c.gh.Repositories.GetContents(ctx, owner, repo, path, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list contents: %w", err)
	}

	return dirContents, nil
}

// SearchCodeResult represents a code search result
type SearchCodeResult struct {
	Repository string
}

// SearchCode searches for code on GitHub
func (c *Client) SearchCode(ctx context.Context, query string, limit int) ([]SearchCodeResult, error) {
	opts := &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: limit,
		},
	}

	result, _, err := c.gh.Search.Code(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("code search failed: %w", err)
	}

	var results []SearchCodeResult
	for _, r := range result.CodeResults {
		if r.Repository != nil && r.Repository.FullName != nil {
			results = append(results, SearchCodeResult{
				Repository: *r.Repository.FullName,
			})
		}
	}

	return results, nil
}

// SearchRepoResult represents a repository search result
type SearchRepoResult struct {
	FullName    string
	Description string
	Stars       int
}

// SearchRepos searches for repositories on GitHub
func (c *Client) SearchRepos(ctx context.Context, query string, limit int) ([]SearchRepoResult, error) {
	opts := &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: limit,
		},
	}

	result, _, err := c.gh.Search.Repositories(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("repository search failed: %w", err)
	}

	var results []SearchRepoResult
	for _, r := range result.Repositories {
		res := SearchRepoResult{}
		if r.FullName != nil {
			res.FullName = *r.FullName
		}
		if r.Description != nil {
			res.Description = *r.Description
		}
		if r.StargazersCount != nil {
			res.Stars = *r.StargazersCount
		}
		results = append(results, res)
	}

	return results, nil
}

// getToken attempts to get a GitHub token from various sources
func getToken() string {
	// 1. GITHUB_TOKEN env var
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token
	}

	// 2. GH_TOKEN env var (gh CLI compat)
	if token := os.Getenv("GH_TOKEN"); token != "" {
		return token
	}

	// 3. Try gh CLI config
	if token := readGhToken(); token != "" {
		return token
	}

	// 4. Unauthenticated (60 req/hr)
	return ""
}

// ghHostsConfig represents the gh CLI hosts.yml config
type ghHostsConfig map[string]struct {
	OAuthToken string `yaml:"oauth_token"`
}

// readGhToken reads the GitHub token from gh CLI config
func readGhToken() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Try hosts.yml (newer gh CLI versions)
	hostsPath := filepath.Join(homeDir, ".config", "gh", "hosts.yml")
	if data, err := os.ReadFile(hostsPath); err == nil {
		var hosts ghHostsConfig
		if err := yaml.Unmarshal(data, &hosts); err == nil {
			if host, ok := hosts["github.com"]; ok && host.OAuthToken != "" {
				return host.OAuthToken
			}
		}
	}

	return ""
}

// ParseGitHubURL parses a GitHub URL and returns owner, repo, path, and hostname
// Supports:
//   - https://raw.githubusercontent.com/owner/repo/ref/path (public)
//   - https://github.company.com/owner/repo/raw/ref/path (GHE)
//   - https://api.github.com/repos/owner/repo/contents/path
//   - https://github.company.com/api/v3/repos/owner/repo/contents/path
func ParseGitHubURL(rawURL string) (owner, repo, path, hostname string, err error) {
	// raw.githubusercontent.com format
	if strings.Contains(rawURL, "raw.githubusercontent.com") {
		parts := strings.Split(strings.TrimPrefix(rawURL, "https://raw.githubusercontent.com/"), "/")
		if len(parts) < 4 {
			return "", "", "", "", fmt.Errorf("invalid raw URL")
		}
		return parts[0], parts[1], strings.Join(parts[3:], "/"), "", nil
	}

	// GHE raw format: /owner/repo/raw/ref/path
	if strings.Contains(rawURL, "/raw/") {
		u, err := url.Parse(rawURL)
		if err != nil {
			return "", "", "", "", err
		}
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) < 5 || parts[2] != "raw" {
			return "", "", "", "", fmt.Errorf("invalid GHE raw URL")
		}
		return parts[0], parts[1], strings.Join(parts[4:], "/"), u.Host, nil
	}

	// API URL format: /repos/owner/repo/contents/path
	if strings.Contains(rawURL, "/repos/") && strings.Contains(rawURL, "/contents") {
		u, err := url.Parse(rawURL)
		if err != nil {
			return "", "", "", "", err
		}

		// Remove /api/v3 prefix if present (GHE)
		apiPath := strings.TrimPrefix(u.Path, "/api/v3")
		apiPath = strings.TrimPrefix(apiPath, "/")

		// Format: repos/owner/repo/contents or repos/owner/repo/contents/path
		parts := strings.Split(apiPath, "/")
		if len(parts) < 4 || parts[0] != "repos" {
			return "", "", "", "", fmt.Errorf("invalid API URL format")
		}

		owner = parts[1]
		repo = parts[2]
		hostname = ""
		if u.Host != "api.github.com" {
			hostname = u.Host
		}

		// Find "contents" and extract path after it
		for i, p := range parts {
			if p == "contents" {
				if i+1 < len(parts) {
					path = strings.Join(parts[i+1:], "/")
				}
				break
			}
		}

		return owner, repo, path, hostname, nil
	}

	return "", "", "", "", fmt.Errorf("unsupported GitHub URL format")
}
