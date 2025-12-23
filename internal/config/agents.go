package config

import (
	"os"
	"path/filepath"

	"github.com/kennyg/tome/internal/schema"
)

// Agent represents a supported AI coding agent
type Agent string

const (
	AgentClaude   Agent = "claude"
	AgentOpenCode Agent = "opencode"
	AgentCopilot  Agent = "copilot"
	AgentCursor   Agent = "cursor"
	AgentWindsurf Agent = "windsurf"
	AgentGemini   Agent = "gemini"
	AgentAmp      Agent = "amp"
)

// AgentCapabilities tracks what artifact types Tome supports installing for each agent.
//
// This matrix represents Tome's current implementation support, NOT the agent's native
// capabilities. For example:
//   - Cursor natively supports .cursorrules (similar to skills), but Tome doesn't
//     convert to that format yet, so Skills=false
//   - GitHub Copilot uses copilot-instructions.md, not SKILL.md, so Skills=false
//   - An agent might support MCP natively but Tome doesn't handle MCP installation
//
// The fields indicate whether Tome can:
//   - Install artifacts of that type to the agent's directory structure
//   - Use the agent's native format for that artifact type
//
// Note: The actual capability checks in the code primarily use empty/non-empty
// directory field strings (e.g., agentCfg.AgentsDir != "") rather than these booleans.
// This struct serves as documentation of Tome's support matrix and may be used for
// future feature validation.
type AgentCapabilities struct {
	Skills   bool // SKILL.md files in skills/
	Commands bool // Slash commands (.md files) in commands/
	Prompts  bool // Prompt templates in prompts/
	Hooks    bool // Event hooks (hooks.json) in hooks/
	Agents   bool // Agent definitions (.md files) in agents/
	Plugins  bool // Full plugin format (.claude-plugin/)
	MCP      bool // Model Context Protocol servers
}

// AgentConfig holds the configuration for a specific agent
type AgentConfig struct {
	Name         Agent
	DisplayName  string
	ConfigDir    string // Relative to home, e.g., ".claude"
	SkillsDir    string // Relative to ConfigDir
	CommandsDir  string // Relative to ConfigDir
	PromptsDir   string // Relative to ConfigDir (if supported)
	HooksDir     string // Relative to ConfigDir (if supported)
	AgentsDir    string // Relative to ConfigDir (if supported)
	PluginsDir   string // Relative to ConfigDir (if supported)
	Capabilities AgentCapabilities
}

// KnownAgents returns all known agent configurations
func KnownAgents() []AgentConfig {
	return []AgentConfig{
		{
			Name:        AgentClaude,
			DisplayName: "Claude Code",
			ConfigDir:   ".claude",
			SkillsDir:   "skills",
			CommandsDir: "commands",
			PromptsDir:  "prompts",
			HooksDir:    "hooks",
			AgentsDir:   "agents",
			PluginsDir:  "plugins",
			Capabilities: AgentCapabilities{
				Skills:   true,
				Commands: true,
				Prompts:  true,
				Hooks:    true,
				Agents:   true,
				Plugins:  true,
				MCP:      true,
			},
		},
		{
			Name:        AgentOpenCode,
			DisplayName: "OpenCode",
			ConfigDir:   ".opencode",
			SkillsDir:   "skills",
			CommandsDir: "command", // Note: singular, not "commands"
			Capabilities: AgentCapabilities{
				Skills:   true,
				Commands: true,
				Prompts:  false,
				Hooks:    false,
				Agents:   false,
				Plugins:  false,
				MCP:      true,
			},
		},
		{
			Name:        AgentCopilot,
			DisplayName: "GitHub Copilot",
			ConfigDir:   ".github",
			SkillsDir:   "agents",  // .agent.md files
			CommandsDir: "prompts", // .prompt.md files
			Capabilities: AgentCapabilities{
				Skills:   true, // Now supported via transmogrify conversion
				Commands: true, // Now supported via transmogrify conversion
				Prompts:  true,
				Hooks:    false,
				Agents:   true,
				Plugins:  false,
				MCP:      false,
			},
		},
		{
			Name:        AgentCursor,
			DisplayName: "Cursor",
			ConfigDir:   ".cursor",
			SkillsDir:   "rules", // .md rules files
			CommandsDir: "rules", // Also rules (Cursor doesn't distinguish)
			Capabilities: AgentCapabilities{
				Skills:   true, // Now supported via transmogrify conversion
				Commands: true, // Now supported via transmogrify conversion
				Prompts:  false,
				Hooks:    false,
				Agents:   false,
				Plugins:  false,
				MCP:      true,
			},
		},
		{
			Name:        AgentWindsurf,
			DisplayName: "Windsurf",
			ConfigDir:   ".windsurf",
			SkillsDir:   "skills",
			CommandsDir: "commands",
			Capabilities: AgentCapabilities{
				Skills:   true,
				Commands: true,
				Prompts:  false,
				Hooks:    false,
				Agents:   false,
				Plugins:  false,
				MCP:      true,
			},
		},
		{
			Name:        AgentGemini,
			DisplayName: "Gemini CLI",
			ConfigDir:   ".gemini",
			SkillsDir:   "",
			CommandsDir: "",
			Capabilities: AgentCapabilities{
				Skills:   false, // TODO: Research
				Commands: false,
				Prompts:  false,
				Hooks:    false,
				Agents:   false,
				Plugins:  false,
				MCP:      false,
			},
		},
		{
			Name:        AgentAmp,
			DisplayName: "Amp",
			ConfigDir:   ".amp",
			SkillsDir:   "",
			CommandsDir: "",
			Capabilities: AgentCapabilities{
				Skills:   false, // TODO: Research
				Commands: false,
				Prompts:  false,
				Hooks:    false,
				Agents:   false,
				Plugins:  false,
				MCP:      false,
			},
		},
	}
}

// GetAgentConfig returns the config for a specific agent
func GetAgentConfig(agent Agent) *AgentConfig {
	for _, a := range KnownAgents() {
		if a.Name == agent {
			return &a
		}
	}
	return nil
}

// DetectInstalledAgents returns agents that appear to be installed
func DetectInstalledAgents() []AgentConfig {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	var installed []AgentConfig
	for _, agent := range KnownAgents() {
		configPath := filepath.Join(home, agent.ConfigDir)
		if _, err := os.Stat(configPath); err == nil {
			installed = append(installed, agent)
		}
	}

	return installed
}

// DefaultAgent returns the default agent to use
// Prefers Claude, falls back to first detected agent
func DefaultAgent() Agent {
	// Check if Claude is installed
	home, _ := os.UserHomeDir()
	claudePath := filepath.Join(home, ".claude")
	if _, err := os.Stat(claudePath); err == nil {
		return AgentClaude
	}

	// Fall back to first detected agent
	installed := DetectInstalledAgents()
	if len(installed) > 0 {
		return installed[0].Name
	}

	// Default to Claude even if not detected
	return AgentClaude
}

// AgentPaths returns the paths for a specific agent
func AgentPaths(home string, agent Agent) (configDir, skillsDir, commandsDir string) {
	cfg := GetAgentConfig(agent)
	if cfg == nil {
		// Fallback to Claude-style paths
		cfg = GetAgentConfig(AgentClaude)
	}

	configDir = filepath.Join(home, cfg.ConfigDir)
	skillsDir = filepath.Join(configDir, cfg.SkillsDir)
	commandsDir = filepath.Join(configDir, cfg.CommandsDir)
	return
}

// AgentToFormat converts a config.Agent to schema.Format for artifact conversion
func AgentToFormat(agent Agent) schema.Format {
	switch agent {
	case AgentClaude, AgentWindsurf:
		return schema.FormatClaude
	case AgentOpenCode:
		return schema.FormatOpenCode
	case AgentCopilot:
		return schema.FormatCopilot
	case AgentCursor:
		return schema.FormatCursor
	default:
		// Default to Claude format
		return schema.FormatClaude
	}
}
