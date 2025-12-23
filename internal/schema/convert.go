package schema

import (
	"fmt"
)

// Convert transforms a skill from one format to another.
// Returns the converted skill as bytes ready to write.
func Convert(skill Skill, targetFormat Format) ([]byte, error) {
	// Extract common metadata
	meta := SkillMetadata{
		Name:        skill.GetName(),
		Description: skill.GetDescription(),
		Body:        skill.GetBody(),
	}

	// Get version/author if available
	switch s := skill.(type) {
	case *ClaudeSkill:
		meta.Version = s.Version
		meta.Author = s.Author
	case *CopilotAgent:
		meta.Version = s.Version
	}

	// Create target format skill
	var target Skill
	switch targetFormat {
	case FormatClaude, FormatOpenCode:
		cs := &ClaudeSkill{}
		cs.FromMetadata(meta)
		cs.SetFormat(targetFormat)
		target = cs
	case FormatCopilot:
		ca := &CopilotAgent{}
		ca.FromMetadata(meta)
		target = ca
	case FormatCursor:
		cs := &CursorSkill{}
		cs.FromMetadata(meta)
		target = cs
	default:
		return nil, fmt.Errorf("unsupported target format: %s", targetFormat)
	}

	return target.Serialize()
}

// ConvertToClaudeSkill converts any skill to ClaudeSkill
func ConvertToClaudeSkill(skill Skill) *ClaudeSkill {
	if cs, ok := skill.(*ClaudeSkill); ok {
		return cs
	}

	cs := &ClaudeSkill{
		Name:        skill.GetName(),
		Description: skill.GetDescription(),
		Body:        skill.GetBody(),
	}

	// Copy additional fields if available
	switch s := skill.(type) {
	case *CopilotAgent:
		cs.Version = s.Version
	}

	return cs
}

// ConvertToCopilotAgent converts any skill to CopilotAgent
func ConvertToCopilotAgent(skill Skill) *CopilotAgent {
	if ca, ok := skill.(*CopilotAgent); ok {
		return ca
	}

	ca := &CopilotAgent{
		Name:        skill.GetName(),
		Description: skill.GetDescription(),
		Body:        skill.GetBody(),
	}

	// Copy additional fields if available
	switch s := skill.(type) {
	case *ClaudeSkill:
		ca.Version = s.Version
	}

	return ca
}

// ConvertToCursorSkill converts any skill to CursorSkill
func ConvertToCursorSkill(skill Skill) *CursorSkill {
	if cs, ok := skill.(*CursorSkill); ok {
		return cs
	}

	return &CursorSkill{
		Name:        skill.GetName(),
		Description: skill.GetDescription(),
		Body:        skill.GetBody(),
	}
}

// Parse parses content based on the detected or specified format
func Parse(content []byte, format Format) (Skill, error) {
	switch format {
	case FormatClaude:
		return ParseClaudeSkill(content)
	case FormatOpenCode:
		return ParseOpenCodeSkill(content)
	case FormatCopilot:
		// Try agent format first, could also be prompt
		return ParseCopilotAgent(content)
	case FormatCursor:
		return ParseCursorSkill(content)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// ParseAuto attempts to detect the format and parse accordingly
func ParseAuto(content []byte, filename string) (Skill, error) {
	format := DetectFormat(filename, content)
	return Parse(content, format)
}

// ConversionResult holds the result of a conversion operation
type ConversionResult struct {
	SourceFormat Format
	TargetFormat Format
	SourceName   string
	TargetName   string
	Content      []byte
	Warnings     []string
}

// ConvertWithInfo converts a skill and returns detailed information
func ConvertWithInfo(skill Skill, targetFormat Format) (*ConversionResult, error) {
	content, err := Convert(skill, targetFormat)
	if err != nil {
		return nil, err
	}

	result := &ConversionResult{
		SourceFormat: skill.GetFormat(),
		TargetFormat: targetFormat,
		SourceName:   skill.GetName(),
		TargetName:   skill.GetName(),
		Content:      content,
	}

	// Check for potential data loss
	if cs, ok := skill.(*ClaudeSkill); ok {
		if len(cs.Globs) > 0 && targetFormat == FormatCopilot {
			result.Warnings = append(result.Warnings,
				"globs field not supported in Copilot format (will be omitted)")
		}
		if len(cs.Includes) > 0 && targetFormat == FormatCopilot {
			result.Warnings = append(result.Warnings,
				"includes field not supported in Copilot format (will be omitted)")
		}
		if len(cs.AllowedTools) > 0 && targetFormat != FormatClaude {
			result.Warnings = append(result.Warnings,
				"allowed-tools field is Claude-specific (will be omitted)")
		}
	}

	return result, nil
}

// OutputFilename returns the appropriate filename for a skill in the target format
func OutputFilename(skill Skill, targetFormat Format) string {
	name := skill.GetName()
	if name == "" {
		name = "skill"
	}

	switch targetFormat {
	case FormatClaude, FormatOpenCode:
		return "SKILL.md"
	case FormatCopilot:
		return toKebabCase(name) + ".agent.md"
	case FormatCursor:
		return toKebabCase(name) + ".md"
	default:
		return name + ".md"
	}
}

// OutputDirectory returns the appropriate directory structure for a skill
func OutputDirectory(skill Skill, targetFormat Format) string {
	name := skill.GetName()
	if name == "" {
		name = "skill"
	}

	switch targetFormat {
	case FormatClaude:
		return "skills/" + toKebabCase(name)
	case FormatOpenCode:
		return ".opencode/skill/" + toKebabCase(name)
	case FormatCopilot:
		return "agents"
	case FormatCursor:
		return ".cursor/rules"
	default:
		return ""
	}
}

// toKebabCase converts a string to kebab-case
func toKebabCase(s string) string {
	result := make([]byte, 0, len(s))
	for i, c := range s {
		if c >= 'A' && c <= 'Z' {
			if i > 0 {
				result = append(result, '-')
			}
			result = append(result, byte(c+32)) // lowercase
		} else if c == ' ' || c == '_' {
			result = append(result, '-')
		} else {
			result = append(result, byte(c))
		}
	}
	return string(result)
}
