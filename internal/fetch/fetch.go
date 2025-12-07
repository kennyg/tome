package fetch

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/kennyg/tome/internal/artifact"
)

// Client handles fetching artifacts from remote sources
type Client struct {
	http   *http.Client
	useGH  bool   // Use gh CLI for auth
	ghPath string // Path to gh CLI
}

// NewClient creates a new fetch client
func NewClient() *Client {
	c := &Client{
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Check if gh CLI is available
	if path, err := exec.LookPath("gh"); err == nil {
		c.useGH = true
		c.ghPath = path
	}

	return c
}

// FetchURL fetches content from a URL
func (c *Client) FetchURL(url string) ([]byte, error) {
	// Try direct fetch first
	resp, err := c.http.Get(url)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return io.ReadAll(resp.Body)
		}
	}

	// If failed and it's a GitHub URL, try gh CLI
	if c.useGH && (strings.Contains(url, "github.com") || strings.Contains(url, "githubusercontent.com")) {
		content, err := c.fetchWithGH(url)
		if err == nil {
			return content, nil
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	return nil, fmt.Errorf("failed to fetch %s: status %d", url, resp.StatusCode)
}

// fetchWithGH fetches file content using gh CLI
func (c *Client) fetchWithGH(url string) ([]byte, error) {
	// Parse the URL to get owner/repo/path
	// raw.githubusercontent.com/owner/repo/ref/path
	// or github.com/owner/repo/raw/ref/path

	var owner, repo, path string

	if strings.Contains(url, "raw.githubusercontent.com") {
		// https://raw.githubusercontent.com/owner/repo/ref/path/to/file
		parts := strings.Split(strings.TrimPrefix(url, "https://raw.githubusercontent.com/"), "/")
		if len(parts) < 4 {
			return nil, fmt.Errorf("invalid raw URL")
		}
		owner = parts[0]
		repo = parts[1]
		// parts[2] is the ref
		path = strings.Join(parts[3:], "/")
	} else {
		return nil, fmt.Errorf("unsupported GitHub URL format for gh fetch")
	}

	// Use gh api to get file content
	apiPath := fmt.Sprintf("repos/%s/%s/contents/%s", owner, repo, path)
	cmd := exec.Command(c.ghPath, "api", apiPath, "--jq", ".content")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh api failed: %w", err)
	}

	// Content is base64 encoded
	content := strings.TrimSpace(string(output))
	// Remove quotes if present
	content = strings.Trim(content, "\"")

	// Decode base64
	decoded, err := base64Decode(content)
	if err != nil {
		return nil, fmt.Errorf("failed to decode content: %w", err)
	}

	return decoded, nil
}

// base64Decode decodes base64 content (handles newlines in GitHub's response)
func base64Decode(s string) ([]byte, error) {
	// GitHub returns base64 with newlines, need to remove them
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\\n", "")

	return base64.StdEncoding.DecodeString(s)
}

// GitHubContent represents a file/directory in GitHub API response
type GitHubContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"` // "file" or "dir"
	DownloadURL string `json:"download_url"`
}

// ListGitHubContents lists files in a GitHub directory
func (c *Client) ListGitHubContents(apiURL string) ([]GitHubContent, error) {
	// Try gh CLI first for authenticated access
	if c.useGH {
		contents, err := c.listWithGH(apiURL)
		if err == nil {
			return contents, nil
		}
		// Fall back to unauthenticated
	}

	resp, err := c.http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to list contents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list contents: status %d", resp.StatusCode)
	}

	var contents []GitHubContent
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return nil, fmt.Errorf("failed to parse contents: %w", err)
	}

	return contents, nil
}

// listWithGH uses gh CLI for authenticated GitHub API access
func (c *Client) listWithGH(apiURL string) ([]GitHubContent, error) {
	// Convert full URL to gh api path
	// https://api.github.com/repos/owner/repo/contents -> repos/owner/repo/contents
	path := strings.TrimPrefix(apiURL, "https://api.github.com/")

	cmd := exec.Command(c.ghPath, "api", path)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh api failed: %w", err)
	}

	var contents []GitHubContent
	if err := json.Unmarshal(output, &contents); err != nil {
		return nil, fmt.Errorf("failed to parse gh output: %w", err)
	}

	return contents, nil
}

// appendPath appends a path segment to a URL, handling query strings properly
func appendPath(baseURL, segment string) string {
	// Split URL and query string
	parts := strings.SplitN(baseURL, "?", 2)
	base := parts[0]
	query := ""
	if len(parts) > 1 {
		query = "?" + parts[1]
	}
	return base + "/" + segment + query
}

// FindArtifacts finds all artifact files in a GitHub directory
func (c *Client) FindArtifacts(apiURL string) ([]GitHubContent, error) {
	contents, err := c.ListGitHubContents(apiURL)
	if err != nil {
		return nil, err
	}

	var artifacts []GitHubContent

	for _, item := range contents {
		// Direct artifact files at root
		if item.Type == "file" && IsArtifactFile(item.Name) {
			artifacts = append(artifacts, item)
			continue
		}

		// Scan commands/ directory for .md files
		if item.Type == "dir" && item.Name == "commands" {
			subURL := appendPath(apiURL, "commands")
			subContents, err := c.ListGitHubContents(subURL)
			if err == nil {
				for _, sub := range subContents {
					if sub.Type == "file" && strings.HasSuffix(strings.ToLower(sub.Name), ".md") {
						artifacts = append(artifacts, sub)
					}
				}
			}
			continue
		}

		// Scan skills/ directory for skill subdirectories with SKILL.md
		if item.Type == "dir" && item.Name == "skills" {
			subURL := appendPath(apiURL, "skills")
			subContents, err := c.ListGitHubContents(subURL)
			if err == nil {
				for _, sub := range subContents {
					// Check for SKILL.md directly in skills/ (flat structure)
					if sub.Type == "file" && strings.ToUpper(sub.Name) == "SKILL.MD" {
						artifacts = append(artifacts, sub)
						continue
					}
					// Check for skill subdirectories with SKILL.md
					if sub.Type == "dir" {
						skillURL := appendPath(subURL, sub.Name)
						skillContents, err := c.ListGitHubContents(skillURL)
						if err == nil {
							for _, skillFile := range skillContents {
								if skillFile.Type == "file" && strings.ToUpper(skillFile.Name) == "SKILL.MD" {
									artifacts = append(artifacts, skillFile)
								}
							}
						}
					}
				}
			}
			continue
		}
	}

	return artifacts, nil
}

// Frontmatter represents the YAML frontmatter in a skill file
type Frontmatter struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Version     string   `yaml:"version,omitempty"`
	Author      string   `yaml:"author,omitempty"`
	Globs       []string `yaml:"globs,omitempty"`
}

// ParseSkill parses a SKILL.md file and returns an artifact
func ParseSkill(content []byte, sourceURL string) (*artifact.Artifact, error) {
	fm, body, err := parseFrontmatter(content)
	if err != nil {
		return nil, err
	}

	// If no frontmatter, try to extract name from filename or content
	name := fm.Name
	if name == "" {
		name = extractNameFromContent(body)
	}

	description := fm.Description
	if description == "" {
		description = extractDescriptionFromContent(body)
	}

	return &artifact.Artifact{
		Name:        name,
		Type:        artifact.TypeSkill,
		Description: description,
		Version:     fm.Version,
		Author:      fm.Author,
		Globs:       fm.Globs,
		SourceURL:   sourceURL,
		Content:     string(content),
		Filename:    "SKILL.md",
	}, nil
}

// ParseCommand parses a command markdown file and returns an artifact
func ParseCommand(content []byte, filename string, sourceURL string) (*artifact.Artifact, error) {
	fm, body, err := parseFrontmatter(content)
	if err != nil {
		return nil, err
	}

	name := fm.Name
	if name == "" {
		// Use filename without extension as name
		name = strings.TrimSuffix(filename, ".md")
	}

	description := fm.Description
	if description == "" {
		description = extractDescriptionFromContent(body)
	}

	return &artifact.Artifact{
		Name:        name,
		Type:        artifact.TypeCommand,
		Description: description,
		Version:     fm.Version,
		Author:      fm.Author,
		SourceURL:   sourceURL,
		Content:     string(content),
		Filename:    filename,
	}, nil
}

// parseFrontmatter extracts YAML frontmatter from content
func parseFrontmatter(content []byte) (*Frontmatter, string, error) {
	text := string(content)
	fm := &Frontmatter{}

	// Check for frontmatter delimiter
	if !strings.HasPrefix(text, "---") {
		return fm, text, nil
	}

	// Find the closing delimiter
	rest := text[3:]
	idx := strings.Index(rest, "\n---")
	if idx == -1 {
		return fm, text, nil
	}

	// Extract and parse YAML
	yamlContent := rest[:idx]
	body := strings.TrimPrefix(rest[idx+4:], "\n")

	if err := yaml.Unmarshal([]byte(yamlContent), fm); err != nil {
		return nil, "", fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return fm, body, nil
}

// extractNameFromContent tries to extract a name from the content
func extractNameFromContent(body string) string {
	// Look for first H1 heading
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return "unnamed-skill"
}

// extractDescriptionFromContent tries to extract a description
func extractDescriptionFromContent(body string) string {
	// Look for first non-empty, non-heading, non-bullet paragraph
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines, headings, bullets, numbered lists, code blocks, horizontal rules
		if line == "" ||
			strings.HasPrefix(line, "#") ||
			strings.HasPrefix(line, "- ") ||
			strings.HasPrefix(line, "* ") ||
			strings.HasPrefix(line, "> ") ||
			strings.HasPrefix(line, "```") ||
			strings.HasPrefix(line, "---") ||
			(len(line) > 2 && line[0] >= '0' && line[0] <= '9' && line[1] == '.') {
			continue
		}
		// Return first 200 chars of first paragraph
		if len(line) > 200 {
			return line[:200] + "..."
		}
		return line
	}
	return ""
}

// DetectArtifactType detects the type of artifact from a filename
func DetectArtifactType(filename string) artifact.Type {
	lower := strings.ToLower(filename)
	base := strings.ToLower(filepath.Base(filename))

	// SKILL.md files are skills
	if base == "skill.md" {
		return artifact.TypeSkill
	}

	// Any other .md file is a command
	if strings.HasSuffix(lower, ".md") {
		return artifact.TypeCommand
	}

	return ""
}

// IsArtifactFile checks if a filename is a potential artifact
func IsArtifactFile(filename string) bool {
	lower := strings.ToLower(filename)
	base := strings.ToLower(filepath.Base(filename))

	// Skip common non-artifact files
	if base == "readme.md" || base == "license.md" || base == "changelog.md" ||
		base == "contributing.md" || base == "agents.md" || base == "claude.md" {
		return false
	}

	// SKILL.md files
	if base == "skill.md" {
		return true
	}

	// Other markdown files (potential commands)
	if strings.HasSuffix(lower, ".md") {
		return true
	}

	return false
}

// CommandNameFromFile extracts a command name from a filename
func CommandNameFromFile(filename string) string {
	base := filepath.Base(filename)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	return SanitizeFilename(name)
}

// SanitizeFilename makes a filename safe for the filesystem
func SanitizeFilename(name string) string {
	// Replace unsafe characters
	re := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	safe := re.ReplaceAllString(name, "-")

	// Remove multiple dashes
	re = regexp.MustCompile(`-+`)
	safe = re.ReplaceAllString(safe, "-")

	// Trim dashes from ends
	safe = strings.Trim(safe, "-")

	if safe == "" {
		safe = "unnamed"
	}

	return safe
}
