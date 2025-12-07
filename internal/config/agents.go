package config

import (
	"os"
	"path/filepath"
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

// AgentCapabilities tracks what features an agent supports
type AgentCapabilities struct {
	Skills   bool // SKILL.md files
	Commands bool // Slash commands (.md files)
	Hooks    bool // Event hooks (hooks.json)
	Agents   bool // Agent definitions
	Plugins  bool // Full plugin format (.claude-plugin/)
	MCP      bool // Model Context Protocol
}

// AgentConfig holds the configuration for a specific agent
type AgentConfig struct {
	Name         Agent
	DisplayName  string
	ConfigDir    string // Relative to home, e.g., ".claude"
	SkillsDir    string // Relative to ConfigDir
	CommandsDir  string // Relative to ConfigDir
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
			HooksDir:    "hooks",
			AgentsDir:   "agents",
			PluginsDir:  "plugins",
			Capabilities: AgentCapabilities{
				Skills:   true,
				Commands: true,
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
			CommandsDir: "commands",
			Capabilities: AgentCapabilities{
				Skills:   true,
				Commands: true,
				Hooks:    false, // Not yet supported
				Agents:   false,
				Plugins:  false,
				MCP:      true,
			},
		},
		{
			Name:        AgentCopilot,
			DisplayName: "GitHub Copilot",
			ConfigDir:   ".github",
			SkillsDir:   "", // Uses copilot-instructions.md instead
			CommandsDir: "",
			Capabilities: AgentCapabilities{
				Skills:   false, // Uses different format
				Commands: false,
				Hooks:    false,
				Agents:   false,
				Plugins:  false,
				MCP:      false,
			},
		},
		{
			Name:        AgentCursor,
			DisplayName: "Cursor",
			ConfigDir:   ".cursor",
			SkillsDir:   "", // Uses .cursorrules instead
			CommandsDir: "",
			Capabilities: AgentCapabilities{
				Skills:   false, // Uses different format
				Commands: false,
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
