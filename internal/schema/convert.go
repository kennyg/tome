package schema

import (
	"fmt"
	"strings"
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

// ConvertCommand transforms a command from one format to another.
// Returns the converted command as bytes ready to write.
func ConvertCommand(cmd Skill, targetFormat Format) ([]byte, error) {
	// Extract common metadata
	meta := SkillMetadata{
		Name:        cmd.GetName(),
		Description: cmd.GetDescription(),
		Body:        cmd.GetBody(),
	}

	// Get version/author if available
	switch c := cmd.(type) {
	case *ClaudeCommand:
		meta.Version = c.Version
		meta.Author = c.Author
	}

	// Create target format command
	var target Skill
	switch targetFormat {
	case FormatClaude, FormatOpenCode:
		cc := &ClaudeCommand{}
		cc.FromMetadata(meta)
		cc.SetFormat(targetFormat)
		target = cc
	case FormatCopilot:
		cp := &CopilotPrompt{}
		cp.FromMetadata(meta)
		target = cp
	case FormatCursor:
		// Cursor doesn't have a command concept, convert to rule
		cs := &CursorSkill{}
		cs.FromMetadata(meta)
		target = cs
	default:
		return nil, fmt.Errorf("unsupported target format for command: %s", targetFormat)
	}

	return target.Serialize()
}

// ConvertToClaudeCommand converts any command to ClaudeCommand
func ConvertToClaudeCommand(cmd Skill) *ClaudeCommand {
	if cc, ok := cmd.(*ClaudeCommand); ok {
		return cc
	}

	cc := &ClaudeCommand{
		Name:        cmd.GetName(),
		Description: cmd.GetDescription(),
		Body:        cmd.GetBody(),
	}

	return cc
}

// ConvertToCopilotPrompt converts any command to CopilotPrompt
func ConvertToCopilotPrompt(cmd Skill) *CopilotPrompt {
	if cp, ok := cmd.(*CopilotPrompt); ok {
		return cp
	}

	cp := &CopilotPrompt{
		Agent:       cmd.GetName(),
		Description: cmd.GetDescription(),
		Body:        cmd.GetBody(),
	}

	return cp
}

// ParseCommand parses content as a command based on the specified format
func ParseCommand(content []byte, format Format) (Skill, error) {
	switch format {
	case FormatClaude:
		return ParseClaudeCommand(content)
	case FormatOpenCode:
		return ParseOpenCodeCommand(content)
	case FormatCopilot:
		return ParseCopilotPrompt(content)
	case FormatCursor:
		// Cursor doesn't have commands, parse as skill/rule
		return ParseCursorSkill(content)
	default:
		return nil, fmt.Errorf("unsupported format for command: %s", format)
	}
}

// ParseCommandAuto attempts to detect the format and parse as command
func ParseCommandAuto(content []byte, filename string) (Skill, error) {
	format := DetectFormat(filename, content)
	return ParseCommand(content, format)
}

// CommandOutputFilename returns the appropriate filename for a command in the target format
func CommandOutputFilename(cmd Skill, targetFormat Format) string {
	name := cmd.GetName()
	if name == "" {
		name = "command"
	}

	switch targetFormat {
	case FormatClaude, FormatOpenCode:
		return toKebabCase(name) + ".md"
	case FormatCopilot:
		return toKebabCase(name) + ".prompt.md"
	case FormatCursor:
		return toKebabCase(name) + ".md"
	default:
		return name + ".md"
	}
}

// CommandOutputDirectory returns the appropriate directory structure for a command
func CommandOutputDirectory(cmd Skill, targetFormat Format) string {
	switch targetFormat {
	case FormatClaude:
		return "commands"
	case FormatOpenCode:
		return ".opencode/command"
	case FormatCopilot:
		return "prompts"
	case FormatCursor:
		return ".cursor/rules"
	default:
		return ""
	}
}

// ConvertCommandWithInfo converts a command and returns detailed information
func ConvertCommandWithInfo(cmd Skill, targetFormat Format) (*ConversionResult, error) {
	content, err := ConvertCommand(cmd, targetFormat)
	if err != nil {
		return nil, err
	}

	result := &ConversionResult{
		SourceFormat: cmd.GetFormat(),
		TargetFormat: targetFormat,
		SourceName:   cmd.GetName(),
		TargetName:   cmd.GetName(),
		Content:      content,
	}

	// Check for potential data loss
	if cc, ok := cmd.(*ClaudeCommand); ok {
		if len(cc.AllowedTools) > 0 && targetFormat != FormatClaude && targetFormat != FormatOpenCode {
			result.Warnings = append(result.Warnings,
				"allowed-tools field is Claude/OpenCode-specific (will be omitted)")
		}
		if cc.Version != "" && targetFormat == FormatCopilot {
			result.Warnings = append(result.Warnings,
				"version field not supported in Copilot prompts (will be omitted)")
		}
		if cc.Author != "" && targetFormat == FormatCopilot {
			result.Warnings = append(result.Warnings,
				"author field not supported in Copilot prompts (will be omitted)")
		}
	}

	if targetFormat == FormatCursor {
		result.Warnings = append(result.Warnings,
			"Cursor doesn't have commands; converting to rule instead")
	}

	return result, nil
}

// ConvertInstructions transforms instructions from one format to another.
// Returns the converted instructions as bytes ready to write.
func ConvertInstructions(inst Skill, targetFormat Format) ([]byte, error) {
	body := inst.GetBody()
	desc := inst.GetDescription()

	var target Skill
	switch targetFormat {
	case FormatClaude:
		ci := &ClaudeInstructions{Body: body}
		ci.SetFormat(FormatClaude)
		target = ci
	case FormatOpenCode:
		ci := &ClaudeInstructions{Body: body}
		ci.SetFormat(FormatOpenCode)
		target = ci
	case FormatCopilot:
		ci := &CopilotInstructions{
			Description: desc,
			Body:        body,
		}
		// Try to preserve applyTo if source is Copilot
		if src, ok := inst.(*CopilotInstructions); ok {
			ci.ApplyTo = src.ApplyTo
		}
		target = ci
	case FormatCursor:
		cr := &CursorRules{
			Description: desc,
			Body:        body,
		}
		// Try to preserve globs if source is Cursor MDC
		if src, ok := inst.(*CursorRules); ok {
			cr.Globs = src.Globs
			cr.AlwaysApply = src.AlwaysApply
		}
		target = cr
	default:
		return nil, fmt.Errorf("unsupported target format for instructions: %s", targetFormat)
	}

	return target.Serialize()
}

// ConvertToClaudeInstructions converts any instructions to ClaudeInstructions
func ConvertToClaudeInstructions(inst Skill) *ClaudeInstructions {
	if ci, ok := inst.(*ClaudeInstructions); ok {
		return ci
	}

	return &ClaudeInstructions{
		Body:         inst.GetBody(),
		sourceFormat: FormatClaude,
	}
}

// ConvertToCopilotInstructions converts any instructions to CopilotInstructions
func ConvertToCopilotInstructions(inst Skill) *CopilotInstructions {
	if ci, ok := inst.(*CopilotInstructions); ok {
		return ci
	}

	ci := &CopilotInstructions{
		Description: inst.GetDescription(),
		Body:        inst.GetBody(),
	}

	// Try to derive description from body if empty
	if ci.Description == "" {
		// Use first line as description
		lines := strings.SplitN(ci.Body, "\n", 2)
		if len(lines) > 0 {
			first := strings.TrimPrefix(lines[0], "# ")
			first = strings.TrimSpace(first)
			if len(first) > 100 {
				first = first[:100] + "..."
			}
			ci.Description = first
		}
	}

	return ci
}

// ConvertToCursorRules converts any instructions to CursorRules
func ConvertToCursorRules(inst Skill) *CursorRules {
	if cr, ok := inst.(*CursorRules); ok {
		return cr
	}

	return &CursorRules{
		Description: inst.GetDescription(),
		Body:        inst.GetBody(),
		isLegacy:    true, // Default to legacy format for simplicity
	}
}

// ParseInstructions parses content as instructions based on the specified format
func ParseInstructions(content []byte, format Format) (Skill, error) {
	switch format {
	case FormatClaude:
		return ParseClaudeInstructions(content)
	case FormatOpenCode:
		return ParseOpenCodeInstructions(content)
	case FormatCopilot:
		return ParseCopilotInstructions(content)
	case FormatCursor:
		return ParseCursorRules(content)
	default:
		return nil, fmt.Errorf("unsupported format for instructions: %s", format)
	}
}

// ParseInstructionsAuto attempts to detect the format and parse as instructions
func ParseInstructionsAuto(content []byte, filename string) (Skill, error) {
	format := DetectFormat(filename, content)
	return ParseInstructions(content, format)
}

// ConvertInstructionsWithInfo converts instructions and returns detailed information
func ConvertInstructionsWithInfo(inst Skill, targetFormat Format) (*ConversionResult, error) {
	content, err := ConvertInstructions(inst, targetFormat)
	if err != nil {
		return nil, err
	}

	result := &ConversionResult{
		SourceFormat: inst.GetFormat(),
		TargetFormat: targetFormat,
		SourceName:   inst.GetName(),
		TargetName:   inst.GetName(),
		Content:      content,
	}

	// Check for potential data loss
	if ci, ok := inst.(*CopilotInstructions); ok {
		if ci.ApplyTo != "" && targetFormat != FormatCopilot {
			result.Warnings = append(result.Warnings,
				"applyTo glob pattern is Copilot-specific (will be omitted)")
		}
	}

	if cr, ok := inst.(*CursorRules); ok {
		if cr.Globs != "" && targetFormat != FormatCursor {
			result.Warnings = append(result.Warnings,
				"globs field is Cursor-specific (will be omitted)")
		}
		if cr.AlwaysApply && targetFormat != FormatCursor {
			result.Warnings = append(result.Warnings,
				"alwaysApply field is Cursor-specific (will be omitted)")
		}
	}

	return result, nil
}
