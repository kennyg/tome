package fetch

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/kennyg/tome/internal/artifact"
	"github.com/kennyg/tome/internal/ghclient"
)

// Client handles fetching artifacts from remote sources
type Client struct {
	http *http.Client
	gh   *ghclient.Client
}

// NewClient creates a new fetch client
func NewClient() *Client {
	return &Client{
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
		gh: ghclient.New(),
	}
}

// FetchURL fetches content from a URL
func (c *Client) FetchURL(rawURL string) ([]byte, error) {
	// Try direct fetch first
	resp, err := c.http.Get(rawURL)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return io.ReadAll(resp.Body)
		}
	}

	// If failed and it's a GitHub URL, try go-github
	if strings.Contains(rawURL, "github.com") || strings.Contains(rawURL, "githubusercontent.com") {
		content, ghErr := c.fetchWithGitHub(rawURL)
		if ghErr == nil {
			return content, nil
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", rawURL, err)
	}
	return nil, fmt.Errorf("failed to fetch %s: status %d", rawURL, resp.StatusCode)
}

// fetchWithGitHub fetches file content using go-github
func (c *Client) fetchWithGitHub(rawURL string) ([]byte, error) {
	owner, repo, path, hostname, err := ghclient.ParseGitHubURL(rawURL)
	if err != nil {
		return nil, err
	}

	// Use appropriate client for the host
	client := c.gh
	if hostname != "" {
		client = ghclient.NewForHost(hostname)
	}

	return client.GetContents(context.Background(), owner, repo, path, nil)
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
	SkillDir    string `json:"-"` // For skills: the directory containing SKILL.md
}

// ListGitHubContents lists files in a GitHub directory
func (c *Client) ListGitHubContents(apiURL string) ([]GitHubContent, error) {
	// Try go-github first for authenticated access
	contents, err := c.listWithGitHub(apiURL)
	if err == nil {
		return contents, nil
	}

	// Fall back to direct HTTP (unauthenticated)
	resp, err := c.http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to list contents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list contents: status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return nil, fmt.Errorf("failed to parse contents: %w", err)
	}

	return contents, nil
}

// listWithGitHub uses go-github for GitHub API access
func (c *Client) listWithGitHub(apiURL string) ([]GitHubContent, error) {
	owner, repo, path, hostname, err := ghclient.ParseGitHubURL(apiURL)
	if err != nil {
		return nil, err
	}

	// Use appropriate client for the host
	client := c.gh
	if hostname != "" {
		client = ghclient.NewForHost(hostname)
	}

	repoContents, err := client.ListContents(context.Background(), owner, repo, path, nil)
	if err != nil {
		return nil, err
	}

	// Convert to GitHubContent
	var contents []GitHubContent
	for _, rc := range repoContents {
		content := GitHubContent{
			Type: rc.GetType(),
		}
		if rc.Name != nil {
			content.Name = *rc.Name
		}
		if rc.Path != nil {
			content.Path = *rc.Path
		}
		if rc.DownloadURL != nil {
			content.DownloadURL = *rc.DownloadURL
		}
		contents = append(contents, content)
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

// agentSkillDirs are directories where different AI agents store skills
var agentSkillDirs = []string{
	"skills",          // Generic/tome standard
	".agent/skills",   // agentskills.io standard
	".github/skills",  // GitHub Copilot
	".claude/skills",  // Claude Code
	".opencode/skills", // OpenCode
	".cursor/skills",  // Cursor
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
			c.scanSkillsDir(apiURL, "skills", &artifacts)
			continue
		}

		// Scan agent-specific directories (e.g., .agent, .github, .claude, etc.)
		if item.Type == "dir" && strings.HasPrefix(item.Name, ".") {
			for _, agentDir := range agentSkillDirs {
				if strings.HasPrefix(agentDir, item.Name+"/") {
					// This is an agent directory that might contain skills
					// e.g., .agent -> .agent/skills, .github -> .github/skills
					skillsSubdir := strings.TrimPrefix(agentDir, item.Name+"/")
					agentURL := appendPath(apiURL, item.Name)
					agentContents, err := c.ListGitHubContents(agentURL)
					if err == nil {
						for _, agentItem := range agentContents {
							if agentItem.Type == "dir" && agentItem.Name == skillsSubdir {
								c.scanSkillsDir(apiURL, agentDir, &artifacts)
							}
						}
					}
				}
			}
		}
	}

	return artifacts, nil
}

// scanSkillsDir scans a skills directory for SKILL.md files
func (c *Client) scanSkillsDir(apiURL string, skillsPath string, artifacts *[]GitHubContent) {
	subURL := appendPath(apiURL, skillsPath)
	subContents, err := c.ListGitHubContents(subURL)
	if err != nil {
		return
	}

	for _, sub := range subContents {
		// Check for SKILL.md directly in skills/ (flat structure)
		if sub.Type == "file" && strings.ToUpper(sub.Name) == "SKILL.MD" {
			*artifacts = append(*artifacts, sub)
			continue
		}
		// Check for skill subdirectories with SKILL.md
		if sub.Type == "dir" {
			skillURL := appendPath(subURL, sub.Name)
			skillContents, err := c.ListGitHubContents(skillURL)
			if err == nil {
				for _, skillFile := range skillContents {
					if skillFile.Type == "file" && strings.ToUpper(skillFile.Name) == "SKILL.MD" {
						// Track the skill directory for fetching includes
						skillFile.SkillDir = skillsPath + "/" + sub.Name
						*artifacts = append(*artifacts, skillFile)
					}
				}
			}
		}
	}
}

// FetchManifest tries to fetch and parse a tome.yaml manifest from a GitHub repo
func (c *Client) FetchManifest(apiURL string) (*artifact.Manifest, error) {
	// Try to fetch tome.yaml
	manifestURL := appendPath(apiURL, "tome.yaml")
	contents, err := c.ListGitHubContents(apiURL)
	if err != nil {
		return nil, err
	}

	// Look for tome.yaml in the listing
	var downloadURL string
	for _, item := range contents {
		if item.Type == "file" && (item.Name == "tome.yaml" || item.Name == "tome.yml") {
			downloadURL = item.DownloadURL
			break
		}
	}

	if downloadURL == "" {
		return nil, nil // No manifest, not an error
	}

	// Fetch the manifest content
	content, err := c.FetchURL(downloadURL)
	if err != nil {
		// Try using the API URL directly
		_ = manifestURL // suppress unused warning
		return nil, nil
	}

	var manifest artifact.Manifest
	if err := yaml.Unmarshal(content, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse tome.yaml: %w", err)
	}

	return &manifest, nil
}

// Frontmatter represents the YAML frontmatter in a skill file
type Frontmatter struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	Version      string   `yaml:"version,omitempty"`
	Author       string   `yaml:"author,omitempty"`
	License      string   `yaml:"license,omitempty"`
	Globs        []string `yaml:"globs,omitempty"`
	Includes     []string `yaml:"includes,omitempty"`      // Optional: limit which files to install
	AllowedTools []string `yaml:"allowed-tools,omitempty"` // Pre-approved tools for Claude Code
}

// Allowed file extensions for skill includes (security whitelist)
var allowedExtensions = map[string]bool{
	// Safe text files
	".md":   true,
	".txt":  true,
	".json": true,
	".yaml": true,
	".yml":  true,
	".toml": true,
	".tmpl": true,
	// Scripts (installed as non-executable 0644)
	".py": true,
	".sh": true,
	".js": true,
	".ts": true,
	".rb": true,
}

// Script extensions that should be installed non-executable
var scriptExtensions = map[string]bool{
	".py": true,
	".sh": true,
	".js": true,
	".ts": true,
	".rb": true,
}

// IsScriptFile returns true if the file is a script
func IsScriptFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return scriptExtensions[ext]
}

// ValidateIncludePath checks if an include path is safe
func ValidateIncludePath(path string) error {
	// No absolute paths
	if strings.HasPrefix(path, "/") {
		return fmt.Errorf("absolute paths not allowed: %s", path)
	}
	// No parent directory traversal
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal not allowed: %s", path)
	}
	// Check extension whitelist
	ext := strings.ToLower(filepath.Ext(path))
	if !allowedExtensions[ext] {
		return fmt.Errorf("file type not allowed: %s (allowed: .md, .txt, .json, .yaml, .yml, .toml, .tmpl)", ext)
	}
	return nil
}

// IncludedFile represents an additional file to install with a skill
type IncludedFile struct {
	Path    string // Relative path within skill directory
	Content []byte
}

// MaxIncludeFileSize is the maximum size for an included file (100KB)
const MaxIncludeFileSize = 100 * 1024

// MaxTotalIncludeSize is the maximum total size for all includes (1MB)
const MaxTotalIncludeSize = 1024 * 1024

// FetchSkillIncludes fetches additional files declared in a skill's includes
func (c *Client) FetchSkillIncludes(baseURL string, skillDir string, includes []string) ([]IncludedFile, error) {
	var files []IncludedFile
	var totalSize int64

	for _, inc := range includes {
		// Build URL for the include file
		incURL := baseURL
		if skillDir != "" {
			incURL = appendPath(baseURL, skillDir)
		}
		incURL = appendPath(incURL, inc)

		// Fetch the file
		content, err := c.FetchURL(incURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch include %s: %w", inc, err)
		}

		// Check file size
		if len(content) > MaxIncludeFileSize {
			return nil, fmt.Errorf("include %s exceeds max size (%d > %d bytes)", inc, len(content), MaxIncludeFileSize)
		}

		totalSize += int64(len(content))
		if totalSize > MaxTotalIncludeSize {
			return nil, fmt.Errorf("total include size exceeds max (%d > %d bytes)", totalSize, MaxTotalIncludeSize)
		}

		files = append(files, IncludedFile{
			Path:    inc,
			Content: content,
		})
	}

	return files, nil
}

// DiscoverSkillFiles auto-discovers all files in a skill directory
func (c *Client) DiscoverSkillFiles(apiURL string, skillDir string) ([]IncludedFile, error) {
	var files []IncludedFile
	var totalSize int64

	// Recursively discover files
	err := c.discoverFilesRecursive(apiURL, skillDir, "", &files, &totalSize)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (c *Client) discoverFilesRecursive(apiURL string, skillDir string, subPath string, files *[]IncludedFile, totalSize *int64) error {
	// Build URL for this directory
	dirURL := apiURL
	if skillDir != "" {
		dirURL = appendPath(dirURL, skillDir)
	}
	if subPath != "" {
		dirURL = appendPath(dirURL, subPath)
	}

	contents, err := c.ListGitHubContents(dirURL)
	if err != nil {
		return err
	}

	for _, item := range contents {
		relPath := item.Name
		if subPath != "" {
			relPath = subPath + "/" + item.Name
		}

		if item.Type == "dir" {
			// Recurse into subdirectory
			if err := c.discoverFilesRecursive(apiURL, skillDir, relPath, files, totalSize); err != nil {
				// Skip directories we can't access
				continue
			}
		} else if item.Type == "file" {
			// Skip SKILL.md - it's handled separately as the main file
			if strings.ToUpper(item.Name) == "SKILL.MD" {
				continue
			}

			// Validate extension
			if err := ValidateIncludePath(relPath); err != nil {
				// Skip files with disallowed extensions
				continue
			}

			// Fetch the file
			content, err := c.FetchURL(item.DownloadURL)
			if err != nil {
				continue // Skip files we can't fetch
			}

			// Check file size
			if len(content) > MaxIncludeFileSize {
				continue // Skip oversized files
			}

			*totalSize += int64(len(content))
			if *totalSize > MaxTotalIncludeSize {
				return fmt.Errorf("total skill size exceeds max (%d bytes)", MaxTotalIncludeSize)
			}

			*files = append(*files, IncludedFile{
				Path:    relPath,
				Content: content,
			})
		}
	}

	return nil
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

	// Validate includes
	var validIncludes []string
	for _, inc := range fm.Includes {
		if err := ValidateIncludePath(inc); err != nil {
			return nil, fmt.Errorf("invalid include: %w", err)
		}
		validIncludes = append(validIncludes, inc)
	}

	return &artifact.Artifact{
		Name:        name,
		Type:        artifact.TypeSkill,
		Description: description,
		Version:     fm.Version,
		Author:      fm.Author,
		Globs:       fm.Globs,
		Includes:    validIncludes,
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
