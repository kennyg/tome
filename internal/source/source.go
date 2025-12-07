package source

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Type represents the source type
type Type string

const (
	TypeGitHub Type = "github"
	TypeURL    Type = "url"
	TypeLocal  Type = "local"
)

// Source represents a parsed artifact source
type Source struct {
	Type     Type
	Host     string // GitHub host (github.com or GHE hostname)
	Owner    string // GitHub owner
	Repo     string // GitHub repo
	Path     string // Subpath within repo or local path
	URL      string // Full URL for URL type
	Ref      string // Git ref (branch, tag, commit)
	Original string // Original input string
}

var (
	// Matches owner/repo or owner/repo:path
	githubShorthand = regexp.MustCompile(`^([a-zA-Z0-9_-]+)/([a-zA-Z0-9_.-]+)(?::(.+))?$`)

	// Matches owner/repo@ref or owner/repo:path@ref
	githubWithRef = regexp.MustCompile(`^([a-zA-Z0-9_-]+)/([a-zA-Z0-9_.-]+)(?::([^@]+))?@(.+)$`)
)

// Parse parses a source string into a Source struct
func Parse(input string) (*Source, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty source")
	}

	// Check for local path first
	if isLocalPath(input) {
		absPath, err := filepath.Abs(input)
		if err != nil {
			return nil, fmt.Errorf("invalid local path: %w", err)
		}
		return &Source{
			Type:     TypeLocal,
			Path:     absPath,
			Original: input,
		}, nil
	}

	// Check for full URL
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		return parseURL(input)
	}

	// Try GitHub shorthand with ref (owner/repo:path@ref)
	if matches := githubWithRef.FindStringSubmatch(input); matches != nil {
		return &Source{
			Type:     TypeGitHub,
			Host:     "github.com",
			Owner:    matches[1],
			Repo:     matches[2],
			Path:     matches[3],
			Ref:      matches[4],
			Original: input,
		}, nil
	}

	// Try GitHub shorthand (owner/repo or owner/repo:path)
	if matches := githubShorthand.FindStringSubmatch(input); matches != nil {
		return &Source{
			Type:     TypeGitHub,
			Host:     "github.com",
			Owner:    matches[1],
			Repo:     matches[2],
			Path:     matches[3],
			Ref:      "main", // Default to main
			Original: input,
		}, nil
	}

	return nil, fmt.Errorf("unable to parse source: %s", input)
}

// parseURL parses a full URL into a Source
func parseURL(input string) (*Source, error) {
	u, err := url.Parse(input)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Check if it's a GitHub URL (public or enterprise)
	if isGitHubHost(u.Host) {
		return parseGitHubURL(u, input)
	}

	// Generic URL
	return &Source{
		Type:     TypeURL,
		URL:      input,
		Original: input,
	}, nil
}

// isGitHubHost checks if a host is GitHub (public or enterprise)
func isGitHubHost(host string) bool {
	// Public GitHub
	if host == "github.com" || host == "raw.githubusercontent.com" {
		return true
	}

	// GitHub Enterprise patterns
	// - github.company.com
	// - git.company.com
	// - ghe.company.com
	// - company.github.com (cloud)
	// - raw.github.company.com (GHE raw)
	lowerHost := strings.ToLower(host)
	if strings.Contains(lowerHost, "github") {
		return true
	}
	if strings.HasPrefix(lowerHost, "git.") || strings.HasPrefix(lowerHost, "ghe.") {
		return true
	}
	// Check for raw.* pattern with github in rest
	if strings.HasPrefix(lowerHost, "raw.") && strings.Contains(lowerHost, "github") {
		return true
	}

	return false
}

// parseGitHubURL parses GitHub URLs (public or enterprise) into a Source
func parseGitHubURL(u *url.URL, original string) (*Source, error) {
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid GitHub URL: %s", original)
	}

	host := u.Host
	// Normalize raw.githubusercontent.com to github.com for host tracking
	if host == "raw.githubusercontent.com" {
		host = "github.com"
	}
	// Normalize raw.* GHE hosts
	if strings.HasPrefix(strings.ToLower(host), "raw.") {
		host = strings.TrimPrefix(strings.ToLower(host), "raw.")
	}

	src := &Source{
		Type:     TypeGitHub,
		Host:     host,
		Owner:    parts[0],
		Repo:     parts[1],
		Ref:      "main",
		URL:      original,
		Original: original,
	}

	// Handle raw.githubusercontent.com URLs
	// Format: raw.githubusercontent.com/owner/repo/ref/path
	if strings.Contains(u.Host, "raw.githubusercontent.com") && len(parts) >= 3 {
		src.Ref = parts[2]
		if len(parts) > 3 {
			src.Path = strings.Join(parts[3:], "/")
		}
		return src, nil
	}

	// Handle GHE raw URLs (raw.github.company.com/owner/repo/ref/path)
	if strings.HasPrefix(strings.ToLower(u.Host), "raw.") && len(parts) >= 3 {
		src.Ref = parts[2]
		if len(parts) > 3 {
			src.Path = strings.Join(parts[3:], "/")
		}
		return src, nil
	}

	// Handle github.com or GHE URLs
	// Format: github.com/owner/repo/blob/ref/path or github.com/owner/repo/tree/ref/path
	if len(parts) >= 4 && (parts[2] == "blob" || parts[2] == "tree") {
		src.Ref = parts[3]
		if len(parts) > 4 {
			src.Path = strings.Join(parts[4:], "/")
		}
	}

	// Handle GHE /raw/ URLs (github.company.com/owner/repo/raw/ref/path)
	if len(parts) >= 4 && parts[2] == "raw" {
		src.Ref = parts[3]
		if len(parts) > 4 {
			src.Path = strings.Join(parts[4:], "/")
		}
	}

	return src, nil
}

// isLocalPath checks if the input looks like a local path
func isLocalPath(input string) bool {
	// Starts with . or / or ~ or is a Windows path
	if strings.HasPrefix(input, ".") ||
		strings.HasPrefix(input, "/") ||
		strings.HasPrefix(input, "~") ||
		(len(input) >= 2 && input[1] == ':') {
		return true
	}

	// Check if path exists on filesystem
	_, err := os.Stat(input)
	return err == nil
}

// GitHubRawURL returns the raw content URL for a GitHub source
func (s *Source) GitHubRawURL(path string) string {
	if s.Type != TypeGitHub {
		return ""
	}
	fullPath := path
	if s.Path != "" && path == "" {
		fullPath = s.Path
	} else if s.Path != "" {
		fullPath = s.Path + "/" + path
	}

	// Public GitHub
	if s.Host == "github.com" || s.Host == "" {
		return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s",
			s.Owner, s.Repo, s.Ref, fullPath)
	}

	// GitHub Enterprise - use /raw/ path
	return fmt.Sprintf("https://%s/%s/%s/raw/%s/%s",
		s.Host, s.Owner, s.Repo, s.Ref, fullPath)
}

// GitHubAPIURL returns the GitHub API URL for listing contents
func (s *Source) GitHubAPIURL() string {
	if s.Type != TypeGitHub {
		return ""
	}

	var base string
	if s.Host == "github.com" || s.Host == "" {
		base = fmt.Sprintf("https://api.github.com/repos/%s/%s/contents", s.Owner, s.Repo)
	} else {
		// GitHub Enterprise API
		base = fmt.Sprintf("https://%s/api/v3/repos/%s/%s/contents", s.Host, s.Owner, s.Repo)
	}

	if s.Path != "" {
		base += "/" + s.Path
	}
	if s.Ref != "" {
		base += "?ref=" + s.Ref
	}
	return base
}

// IsEnterprise returns true if this is a GitHub Enterprise source
func (s *Source) IsEnterprise() bool {
	return s.Host != "" && s.Host != "github.com"
}

// String returns a human-readable representation
func (s *Source) String() string {
	switch s.Type {
	case TypeGitHub:
		result := fmt.Sprintf("%s/%s", s.Owner, s.Repo)
		if s.Path != "" {
			result += ":" + s.Path
		}
		if s.Ref != "" && s.Ref != "main" {
			result += "@" + s.Ref
		}
		return result
	case TypeLocal:
		return s.Path
	case TypeURL:
		return s.URL
	default:
		return s.Original
	}
}
