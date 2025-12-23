package schema

import (
	"path/filepath"
	"strings"
)

// CopilotAgent represents a GitHub Copilot agent (.agent.md format).
type CopilotAgent struct {
	// Core fields
	Name        string `yaml:"name"`
	Description string `yaml:"description"`

	// Optional metadata
	Version string `yaml:"version,omitempty"`

	// Content
	Body string `yaml:"-"` // Markdown body (not in frontmatter)
}

// Ensure CopilotAgent implements Skill interface
var _ Skill = (*CopilotAgent)(nil)

// GetName returns the agent name
func (a *CopilotAgent) GetName() string {
	return a.Name
}

// GetDescription returns the agent description
func (a *CopilotAgent) GetDescription() string {
	return a.Description
}

// GetBody returns the markdown body content
func (a *CopilotAgent) GetBody() string {
	return a.Body
}

// GetFormat returns the source format
func (a *CopilotAgent) GetFormat() Format {
	return FormatCopilot
}

// Serialize returns the agent as .agent.md content
func (a *CopilotAgent) Serialize() ([]byte, error) {
	fm := &copilotFrontmatter{
		Name:        a.Name,
		Description: a.Description,
		Version:     a.Version,
	}
	return SerializeFrontmatter(fm, a.Body)
}

// copilotFrontmatter controls YAML field ordering
type copilotFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version,omitempty"`
}

// ParseCopilotAgent parses content as a Copilot .agent.md file
func ParseCopilotAgent(content []byte) (*CopilotAgent, error) {
	agent := &CopilotAgent{}
	body, err := ParseFrontmatterTyped(content, agent)
	if err != nil {
		return nil, err
	}
	agent.Body = body
	return agent, nil
}

// ToMetadata extracts common metadata from the agent
func (a *CopilotAgent) ToMetadata() SkillMetadata {
	return SkillMetadata{
		Name:        a.Name,
		Description: a.Description,
		Version:     a.Version,
		Body:        a.Body,
	}
}

// FromMetadata populates the agent from common metadata
func (a *CopilotAgent) FromMetadata(m SkillMetadata) {
	a.Name = m.Name
	a.Description = m.Description
	a.Version = m.Version
	a.Body = m.Body
}

// Filename returns the expected filename for this agent
func (a *CopilotAgent) Filename() string {
	// Convert name to kebab-case and add extension
	name := strings.ToLower(a.Name)
	name = strings.ReplaceAll(name, " ", "-")
	return name + ".agent.md"
}

// CopilotPrompt represents a GitHub Copilot prompt (.prompt.md format).
// Similar to agent but used for slash commands.
type CopilotPrompt struct {
	// Core fields - note: uses "agent" field, not "name"
	Agent       string `yaml:"agent,omitempty"`
	Description string `yaml:"description"`

	// Content
	Body string `yaml:"-"`
}

// Ensure CopilotPrompt implements Skill interface
var _ Skill = (*CopilotPrompt)(nil)

// GetName returns the prompt name (derived from agent field or filename)
func (p *CopilotPrompt) GetName() string {
	if p.Agent != "" {
		return p.Agent
	}
	return "prompt"
}

// GetDescription returns the prompt description
func (p *CopilotPrompt) GetDescription() string {
	return p.Description
}

// GetBody returns the markdown body content
func (p *CopilotPrompt) GetBody() string {
	return p.Body
}

// GetFormat returns the source format
func (p *CopilotPrompt) GetFormat() Format {
	return FormatCopilot
}

// Serialize returns the prompt as .prompt.md content
func (p *CopilotPrompt) Serialize() ([]byte, error) {
	fm := &copilotPromptFrontmatter{
		Agent:       p.Agent,
		Description: p.Description,
	}
	return SerializeFrontmatter(fm, p.Body)
}

type copilotPromptFrontmatter struct {
	Agent       string `yaml:"agent,omitempty"`
	Description string `yaml:"description"`
}

// ParseCopilotPrompt parses content as a Copilot .prompt.md file
func ParseCopilotPrompt(content []byte) (*CopilotPrompt, error) {
	prompt := &CopilotPrompt{}
	body, err := ParseFrontmatterTyped(content, prompt)
	if err != nil {
		return nil, err
	}
	prompt.Body = body
	return prompt, nil
}

// IsCopilotFile checks if a filename matches Copilot patterns
func IsCopilotFile(filename string) bool {
	base := filepath.Base(filename)
	return strings.HasSuffix(base, ".agent.md") ||
		strings.HasSuffix(base, ".prompt.md") ||
		strings.HasSuffix(base, ".instructions.md")
}
