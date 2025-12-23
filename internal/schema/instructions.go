package schema

import (
	"path/filepath"
	"strings"
)

// Instructions represents project-level instruction files.
// These are different from skills/commands - they provide global context.
// - Claude: CLAUDE.md
// - OpenCode: AGENTS.md
// - Copilot: *.instructions.md (with applyTo glob)
// - Cursor: .cursorrules or .cursor/rules/*.mdc

// ClaudeInstructions represents Claude/OpenCode project instructions.
// File: CLAUDE.md or AGENTS.md
type ClaudeInstructions struct {
	// Content
	Body string `yaml:"-"`

	// Source format tracking
	sourceFormat Format
}

// Ensure ClaudeInstructions implements Skill interface
var _ Skill = (*ClaudeInstructions)(nil)

// GetName returns a default name for instructions
func (i *ClaudeInstructions) GetName() string {
	return "instructions"
}

// GetDescription returns empty (instructions don't have descriptions)
func (i *ClaudeInstructions) GetDescription() string {
	return ""
}

// GetBody returns the markdown body content
func (i *ClaudeInstructions) GetBody() string {
	return i.Body
}

// GetFormat returns the source format
func (i *ClaudeInstructions) GetFormat() Format {
	if i.sourceFormat != "" {
		return i.sourceFormat
	}
	return FormatClaude
}

// SetFormat sets the source format
func (i *ClaudeInstructions) SetFormat(f Format) {
	i.sourceFormat = f
}

// Serialize returns the instructions as markdown content
func (i *ClaudeInstructions) Serialize() ([]byte, error) {
	// Claude/OpenCode instructions are just plain markdown
	return []byte(i.Body), nil
}

// ParseClaudeInstructions parses content as Claude instructions
func ParseClaudeInstructions(content []byte) (*ClaudeInstructions, error) {
	return &ClaudeInstructions{
		Body:         string(content),
		sourceFormat: FormatClaude,
	}, nil
}

// ParseOpenCodeInstructions parses content as OpenCode instructions (AGENTS.md)
func ParseOpenCodeInstructions(content []byte) (*ClaudeInstructions, error) {
	return &ClaudeInstructions{
		Body:         string(content),
		sourceFormat: FormatOpenCode,
	}, nil
}

// ToMetadata extracts common metadata
func (i *ClaudeInstructions) ToMetadata() SkillMetadata {
	return SkillMetadata{
		Name: "instructions",
		Body: i.Body,
	}
}

// FromMetadata populates from common metadata
func (i *ClaudeInstructions) FromMetadata(m SkillMetadata) {
	i.Body = m.Body
}

// CopilotInstructions represents GitHub Copilot instructions (.instructions.md).
// Has frontmatter with description and applyTo glob pattern.
type CopilotInstructions struct {
	// Core fields
	Description string `yaml:"description"`
	ApplyTo     string `yaml:"applyTo,omitempty"` // Glob pattern like "**/*.cs"

	// Content
	Body string `yaml:"-"`
}

// Ensure CopilotInstructions implements Skill interface
var _ Skill = (*CopilotInstructions)(nil)

// GetName returns the name (derived from description or default)
func (i *CopilotInstructions) GetName() string {
	if i.Description != "" {
		// Use first few words of description
		words := strings.Fields(i.Description)
		if len(words) > 3 {
			words = words[:3]
		}
		return strings.ToLower(strings.Join(words, "-"))
	}
	return "instructions"
}

// GetDescription returns the description
func (i *CopilotInstructions) GetDescription() string {
	return i.Description
}

// GetBody returns the markdown body content
func (i *CopilotInstructions) GetBody() string {
	return i.Body
}

// GetFormat returns FormatCopilot
func (i *CopilotInstructions) GetFormat() Format {
	return FormatCopilot
}

// Serialize returns the instructions as .instructions.md content
func (i *CopilotInstructions) Serialize() ([]byte, error) {
	fm := &copilotInstructionsFrontmatter{
		Description: i.Description,
		ApplyTo:     i.ApplyTo,
	}
	return SerializeFrontmatter(fm, i.Body)
}

type copilotInstructionsFrontmatter struct {
	Description string `yaml:"description"`
	ApplyTo     string `yaml:"applyTo,omitempty"`
}

// ParseCopilotInstructions parses content as Copilot .instructions.md
func ParseCopilotInstructions(content []byte) (*CopilotInstructions, error) {
	inst := &CopilotInstructions{}
	body, err := ParseFrontmatterTyped(content, inst)
	if err != nil {
		return nil, err
	}
	inst.Body = body
	return inst, nil
}

// ToMetadata extracts common metadata
func (i *CopilotInstructions) ToMetadata() SkillMetadata {
	return SkillMetadata{
		Name:        i.GetName(),
		Description: i.Description,
		Body:        i.Body,
	}
}

// FromMetadata populates from common metadata
func (i *CopilotInstructions) FromMetadata(m SkillMetadata) {
	i.Description = m.Description
	i.Body = m.Body
}

// Filename returns the expected filename
func (i *CopilotInstructions) Filename() string {
	name := i.GetName()
	return strings.ToLower(strings.ReplaceAll(name, " ", "-")) + ".instructions.md"
}

// CursorRules represents Cursor project rules (.cursorrules or .mdc files).
type CursorRules struct {
	// MDC frontmatter fields (for .cursor/rules/*.mdc)
	Description string `yaml:"description,omitempty"`
	Globs       string `yaml:"globs,omitempty"`      // File patterns
	AlwaysApply bool   `yaml:"alwaysApply,omitempty"` // Always include in context

	// Content
	Body string `yaml:"-"`

	// Whether this is legacy .cursorrules (no frontmatter) or MDC
	isLegacy bool
}

// Ensure CursorRules implements Skill interface
var _ Skill = (*CursorRules)(nil)

// GetName returns a default name
func (r *CursorRules) GetName() string {
	if r.Description != "" {
		words := strings.Fields(r.Description)
		if len(words) > 3 {
			words = words[:3]
		}
		return strings.ToLower(strings.Join(words, "-"))
	}
	return "cursorrules"
}

// GetDescription returns the description
func (r *CursorRules) GetDescription() string {
	return r.Description
}

// GetBody returns the content
func (r *CursorRules) GetBody() string {
	return r.Body
}

// GetFormat returns FormatCursor
func (r *CursorRules) GetFormat() Format {
	return FormatCursor
}

// Serialize returns the rules content
func (r *CursorRules) Serialize() ([]byte, error) {
	// Legacy .cursorrules is just plain text
	if r.isLegacy || (r.Description == "" && r.Globs == "" && !r.AlwaysApply) {
		return []byte(r.Body), nil
	}

	// MDC format with frontmatter
	fm := &cursorRulesFrontmatter{
		Description: r.Description,
		Globs:       r.Globs,
		AlwaysApply: r.AlwaysApply,
	}
	return SerializeFrontmatter(fm, r.Body)
}

type cursorRulesFrontmatter struct {
	Description string `yaml:"description,omitempty"`
	Globs       string `yaml:"globs,omitempty"`
	AlwaysApply bool   `yaml:"alwaysApply,omitempty"`
}

// ParseCursorRules parses content as Cursor rules
func ParseCursorRules(content []byte) (*CursorRules, error) {
	text := string(content)
	rules := &CursorRules{}

	// Check if it has frontmatter (MDC format)
	if strings.HasPrefix(text, "---") {
		body, err := ParseFrontmatterTyped(content, rules)
		if err != nil {
			return nil, err
		}
		rules.Body = body
		rules.isLegacy = false
	} else {
		// Legacy .cursorrules - plain text
		rules.Body = text
		rules.isLegacy = true
	}

	return rules, nil
}

// ToMetadata extracts common metadata
func (r *CursorRules) ToMetadata() SkillMetadata {
	return SkillMetadata{
		Name:        r.GetName(),
		Description: r.Description,
		Body:        r.Body,
	}
}

// FromMetadata populates from common metadata
func (r *CursorRules) FromMetadata(m SkillMetadata) {
	r.Description = m.Description
	r.Body = m.Body
}

// IsInstructionsFile checks if a filename matches instruction patterns
func IsInstructionsFile(filename string) bool {
	base := filepath.Base(filename)
	baseLower := strings.ToLower(base)

	// Claude/OpenCode
	if baseLower == "claude.md" || baseLower == "agents.md" {
		return true
	}

	// Copilot
	if strings.HasSuffix(baseLower, ".instructions.md") {
		return true
	}

	// Cursor legacy
	if baseLower == ".cursorrules" {
		return true
	}

	// Cursor MDC
	if strings.HasSuffix(baseLower, ".mdc") && containsPath(filename, ".cursor/rules") {
		return true
	}

	return false
}

// InstructionsOutputFilename returns the appropriate filename for instructions
func InstructionsOutputFilename(inst Skill, targetFormat Format) string {
	switch targetFormat {
	case FormatClaude:
		return "CLAUDE.md"
	case FormatOpenCode:
		return "AGENTS.md"
	case FormatCopilot:
		name := inst.GetName()
		if name == "" || name == "instructions" || name == "cursorrules" {
			name = "project"
		}
		return strings.ToLower(strings.ReplaceAll(name, " ", "-")) + ".instructions.md"
	case FormatCursor:
		return ".cursorrules"
	default:
		return "instructions.md"
	}
}

// InstructionsOutputDirectory returns the appropriate directory for instructions
func InstructionsOutputDirectory(inst Skill, targetFormat Format) string {
	switch targetFormat {
	case FormatClaude, FormatOpenCode:
		return "" // Root directory
	case FormatCopilot:
		return "instructions"
	case FormatCursor:
		return "" // Root directory for .cursorrules
	default:
		return ""
	}
}
