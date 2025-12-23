// Package schema defines artifact schemas for multiple AI agent formats.
// It provides parsing, serialization, and conversion between formats.
package schema

// Format represents an AI agent artifact format
type Format string

const (
	FormatClaude   Format = "claude"   // Claude Code (.claude/skills/*/SKILL.md)
	FormatOpenCode Format = "opencode" // OpenCode (.opencode/skill/*/SKILL.md) - same as Claude
	FormatCopilot  Format = "copilot"  // GitHub Copilot (agents/*.agent.md)
	FormatCursor   Format = "cursor"   // Cursor (.cursor/rules/*.md)
)

// AllFormats returns all supported formats
func AllFormats() []Format {
	return []Format{FormatClaude, FormatOpenCode, FormatCopilot, FormatCursor}
}

// String returns the string representation of the format
func (f Format) String() string {
	return string(f)
}

// IsValid returns true if the format is recognized
func (f Format) IsValid() bool {
	switch f {
	case FormatClaude, FormatOpenCode, FormatCopilot, FormatCursor:
		return true
	default:
		return false
	}
}

// Skill is the common interface for all skill formats
type Skill interface {
	// GetName returns the skill name
	GetName() string

	// GetDescription returns the skill description
	GetDescription() string

	// GetBody returns the markdown body content
	GetBody() string

	// GetFormat returns the source format of this skill
	GetFormat() Format

	// Serialize returns the skill as formatted content (frontmatter + body)
	Serialize() ([]byte, error)
}

// SkillMetadata holds common metadata fields across all formats
type SkillMetadata struct {
	Name        string
	Description string
	Version     string
	Author      string
	Body        string
}

// ArtifactType represents the type of artifact (skill, command, etc.)
type ArtifactType string

const (
	ArtifactSkill   ArtifactType = "skill"
	ArtifactCommand ArtifactType = "command"
)

// DetectFormat attempts to detect the format from file path or content
func DetectFormat(filename string, content []byte) Format {
	// Check by filename extension/pattern
	switch {
	case hasExtension(filename, ".agent.md"):
		return FormatCopilot
	case hasExtension(filename, ".prompt.md"):
		return FormatCopilot
	case containsPath(filename, ".cursor"):
		return FormatCursor
	case containsPath(filename, ".opencode"):
		return FormatOpenCode
	case containsPath(filename, ".claude"):
		return FormatClaude
	case hasBasename(filename, "SKILL.md"):
		// Could be Claude or OpenCode, default to Claude
		return FormatClaude
	}

	// Default to Claude format
	return FormatClaude
}

// DetectArtifactType attempts to detect whether a file is a skill or command
func DetectArtifactType(filename string) ArtifactType {
	switch {
	// Copilot patterns
	case hasExtension(filename, ".agent.md"):
		return ArtifactSkill
	case hasExtension(filename, ".prompt.md"):
		return ArtifactCommand

	// Claude/OpenCode patterns
	case containsPath(filename, "commands"):
		return ArtifactCommand
	case containsPath(filename, "command"):
		return ArtifactCommand
	case containsPath(filename, "skills"):
		return ArtifactSkill
	case containsPath(filename, "skill"):
		return ArtifactSkill
	case hasBasename(filename, "SKILL.md"):
		return ArtifactSkill

	// Cursor - all rules are skills
	case containsPath(filename, ".cursor"):
		return ArtifactSkill
	}

	// Default to skill
	return ArtifactSkill
}

// helper functions for path detection
func hasExtension(path, ext string) bool {
	return len(path) >= len(ext) && path[len(path)-len(ext):] == ext
}

func containsPath(path, segment string) bool {
	return contains(path, segment+"/") || contains(path, "/"+segment)
}

func hasBasename(path, name string) bool {
	// Simple check for filename at end of path
	return len(path) >= len(name) && path[len(path)-len(name):] == name
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
