package schema

import (
	"path/filepath"
	"strings"
)

// CursorSkill represents a Cursor rule/skill (.cursor/rules/*.md format).
// Cursor uses a simpler format, often just markdown with optional frontmatter.
type CursorSkill struct {
	// Core fields
	Name        string `yaml:"name,omitempty"`
	Description string `yaml:"description,omitempty"`

	// Content
	Body string `yaml:"-"` // Markdown body (not in frontmatter)
}

// Ensure CursorSkill implements Skill interface
var _ Skill = (*CursorSkill)(nil)

// GetName returns the skill name
func (s *CursorSkill) GetName() string {
	return s.Name
}

// GetDescription returns the skill description
func (s *CursorSkill) GetDescription() string {
	return s.Description
}

// GetBody returns the markdown body content
func (s *CursorSkill) GetBody() string {
	return s.Body
}

// GetFormat returns the source format
func (s *CursorSkill) GetFormat() Format {
	return FormatCursor
}

// Serialize returns the skill as .md content for Cursor
func (s *CursorSkill) Serialize() ([]byte, error) {
	// Cursor often doesn't use frontmatter, but we'll include it if there's metadata
	if s.Name == "" && s.Description == "" {
		// No frontmatter needed
		return []byte(s.Body), nil
	}

	fm := &cursorFrontmatter{
		Name:        s.Name,
		Description: s.Description,
	}
	return SerializeFrontmatter(fm, s.Body)
}

// cursorFrontmatter controls YAML field ordering
type cursorFrontmatter struct {
	Name        string `yaml:"name,omitempty"`
	Description string `yaml:"description,omitempty"`
}

// ParseCursorSkill parses content as a Cursor rule file
func ParseCursorSkill(content []byte) (*CursorSkill, error) {
	skill := &CursorSkill{}

	text := string(content)

	// Check if there's frontmatter
	if strings.HasPrefix(text, "---") {
		body, err := ParseFrontmatterTyped(content, skill)
		if err != nil {
			return nil, err
		}
		skill.Body = body
	} else {
		// No frontmatter, treat entire content as body
		skill.Body = text
	}

	return skill, nil
}

// ToMetadata extracts common metadata from the skill
func (s *CursorSkill) ToMetadata() SkillMetadata {
	return SkillMetadata{
		Name:        s.Name,
		Description: s.Description,
		Body:        s.Body,
	}
}

// FromMetadata populates the skill from common metadata
func (s *CursorSkill) FromMetadata(m SkillMetadata) {
	s.Name = m.Name
	s.Description = m.Description
	s.Body = m.Body
}

// Filename returns the expected filename for this skill
func (s *CursorSkill) Filename() string {
	if s.Name == "" {
		return "rules.md"
	}
	// Convert name to kebab-case
	name := strings.ToLower(s.Name)
	name = strings.ReplaceAll(name, " ", "-")
	return name + ".md"
}

// IsCursorFile checks if a filename matches Cursor patterns
func IsCursorFile(filename string) bool {
	return strings.Contains(filename, ".cursor") ||
		strings.Contains(filepath.Dir(filename), "cursor")
}
