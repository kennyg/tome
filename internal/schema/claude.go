package schema

// ClaudeSkill represents a Claude Code / OpenCode skill (SKILL.md format).
// OpenCode uses the same format, so this struct serves both.
type ClaudeSkill struct {
	// Core fields
	Name        string `yaml:"name"`
	Description string `yaml:"description"`

	// Optional metadata
	Version string `yaml:"version,omitempty"`
	Author  string `yaml:"author,omitempty"`
	License string `yaml:"license,omitempty"`

	// Skill-specific fields
	Globs    []string `yaml:"globs,omitempty"`    // File patterns this skill applies to
	Includes []string `yaml:"includes,omitempty"` // Additional files to include

	// Claude-specific fields
	AllowedTools []string `yaml:"allowed-tools,omitempty"` // Pre-approved tools

	// Content
	Body string `yaml:"-"` // Markdown body (not in frontmatter)

	// Source format tracking
	sourceFormat Format
}

// Ensure ClaudeSkill implements Skill interface
var _ Skill = (*ClaudeSkill)(nil)

// GetName returns the skill name
func (s *ClaudeSkill) GetName() string {
	return s.Name
}

// GetDescription returns the skill description
func (s *ClaudeSkill) GetDescription() string {
	return s.Description
}

// GetBody returns the markdown body content
func (s *ClaudeSkill) GetBody() string {
	return s.Body
}

// GetFormat returns the source format
func (s *ClaudeSkill) GetFormat() Format {
	if s.sourceFormat != "" {
		return s.sourceFormat
	}
	return FormatClaude
}

// SetFormat sets the source format
func (s *ClaudeSkill) SetFormat(f Format) {
	s.sourceFormat = f
}

// Serialize returns the skill as SKILL.md content
func (s *ClaudeSkill) Serialize() ([]byte, error) {
	// Create a copy for serialization to control field order
	fm := &claudeFrontmatter{
		Name:         s.Name,
		Description:  s.Description,
		Version:      s.Version,
		Author:       s.Author,
		License:      s.License,
		Globs:        s.Globs,
		Includes:     s.Includes,
		AllowedTools: s.AllowedTools,
	}
	return SerializeFrontmatter(fm, s.Body)
}

// claudeFrontmatter controls YAML field ordering
type claudeFrontmatter struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	Version      string   `yaml:"version,omitempty"`
	Author       string   `yaml:"author,omitempty"`
	License      string   `yaml:"license,omitempty"`
	Globs        []string `yaml:"globs,omitempty"`
	Includes     []string `yaml:"includes,omitempty"`
	AllowedTools []string `yaml:"allowed-tools,omitempty"`
}

// ParseClaudeSkill parses content as a Claude/OpenCode SKILL.md file
func ParseClaudeSkill(content []byte) (*ClaudeSkill, error) {
	skill := &ClaudeSkill{}
	body, err := ParseFrontmatterTyped(content, skill)
	if err != nil {
		return nil, err
	}
	skill.Body = body
	skill.sourceFormat = FormatClaude
	return skill, nil
}

// ParseOpenCodeSkill is an alias for ParseClaudeSkill (same format)
func ParseOpenCodeSkill(content []byte) (*ClaudeSkill, error) {
	skill, err := ParseClaudeSkill(content)
	if err != nil {
		return nil, err
	}
	skill.sourceFormat = FormatOpenCode
	return skill, nil
}

// ToMetadata extracts common metadata from the skill
func (s *ClaudeSkill) ToMetadata() SkillMetadata {
	return SkillMetadata{
		Name:        s.Name,
		Description: s.Description,
		Version:     s.Version,
		Author:      s.Author,
		Body:        s.Body,
	}
}

// FromMetadata populates the skill from common metadata
func (s *ClaudeSkill) FromMetadata(m SkillMetadata) {
	s.Name = m.Name
	s.Description = m.Description
	s.Version = m.Version
	s.Author = m.Author
	s.Body = m.Body
}
