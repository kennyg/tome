package schema

import (
	"strings"
)

// ClaudeCommand represents a Claude Code / OpenCode command.
// Claude: .claude/commands/<name>.md
// OpenCode: .opencode/command/<name>.md
type ClaudeCommand struct {
	// Core fields
	Name        string `yaml:"name"`
	Description string `yaml:"description"`

	// Optional metadata
	Version string `yaml:"version,omitempty"`
	Author  string `yaml:"author,omitempty"`

	// Command-specific fields
	AllowedTools []string `yaml:"allowed-tools,omitempty"` // Pre-approved tools

	// Content
	Body string `yaml:"-"` // Markdown body (not in frontmatter)

	// Source format tracking
	sourceFormat Format
}

// Ensure ClaudeCommand implements Skill interface
var _ Skill = (*ClaudeCommand)(nil)

// GetName returns the command name
func (c *ClaudeCommand) GetName() string {
	return c.Name
}

// GetDescription returns the command description
func (c *ClaudeCommand) GetDescription() string {
	return c.Description
}

// GetBody returns the markdown body content
func (c *ClaudeCommand) GetBody() string {
	return c.Body
}

// GetFormat returns the source format
func (c *ClaudeCommand) GetFormat() Format {
	if c.sourceFormat != "" {
		return c.sourceFormat
	}
	return FormatClaude
}

// SetFormat sets the source format
func (c *ClaudeCommand) SetFormat(f Format) {
	c.sourceFormat = f
}

// Serialize returns the command as markdown content
func (c *ClaudeCommand) Serialize() ([]byte, error) {
	fm := &commandFrontmatter{
		Name:         c.Name,
		Description:  c.Description,
		Version:      c.Version,
		Author:       c.Author,
		AllowedTools: c.AllowedTools,
	}
	return SerializeFrontmatter(fm, c.Body)
}

// commandFrontmatter controls YAML field ordering
type commandFrontmatter struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	Version      string   `yaml:"version,omitempty"`
	Author       string   `yaml:"author,omitempty"`
	AllowedTools []string `yaml:"allowed-tools,omitempty"`
}

// ParseClaudeCommand parses content as a Claude command file
func ParseClaudeCommand(content []byte) (*ClaudeCommand, error) {
	cmd := &ClaudeCommand{}
	body, err := ParseFrontmatterTyped(content, cmd)
	if err != nil {
		return nil, err
	}
	cmd.Body = body
	cmd.sourceFormat = FormatClaude
	return cmd, nil
}

// ParseOpenCodeCommand parses content as an OpenCode command file
func ParseOpenCodeCommand(content []byte) (*ClaudeCommand, error) {
	cmd, err := ParseClaudeCommand(content)
	if err != nil {
		return nil, err
	}
	cmd.sourceFormat = FormatOpenCode
	return cmd, nil
}

// ToMetadata extracts common metadata from the command
func (c *ClaudeCommand) ToMetadata() SkillMetadata {
	return SkillMetadata{
		Name:        c.Name,
		Description: c.Description,
		Version:     c.Version,
		Author:      c.Author,
		Body:        c.Body,
	}
}

// FromMetadata populates the command from common metadata
func (c *ClaudeCommand) FromMetadata(m SkillMetadata) {
	c.Name = m.Name
	c.Description = m.Description
	c.Version = m.Version
	c.Author = m.Author
	c.Body = m.Body
}

// Filename returns the expected filename for this command
func (c *ClaudeCommand) Filename() string {
	name := c.Name
	if name == "" {
		name = "command"
	}
	// Convert to kebab-case
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	return name + ".md"
}

// IsCommandFile checks if a filename matches command patterns
func IsCommandFile(filename string) bool {
	// Check for Claude/OpenCode command paths
	if containsPath(filename, ".claude/commands") ||
		containsPath(filename, ".opencode/command") {
		return true
	}
	// Check for Copilot prompt files
	if hasExtension(filename, ".prompt.md") {
		return true
	}
	return false
}
