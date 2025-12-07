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
	AgentCrush    Agent = "crush"
	AgentCursor   Agent = "cursor"
	AgentWindsurf Agent = "windsurf"
)

// AgentConfig holds the configuration for a specific agent
type AgentConfig struct {
	Name        Agent
	DisplayName string
	ConfigDir   string // Relative to home, e.g., ".claude"
	SkillsDir   string // Relative to ConfigDir
	CommandsDir string // Relative to ConfigDir
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
		},
		{
			Name:        AgentOpenCode,
			DisplayName: "OpenCode",
			ConfigDir:   ".opencode",
			SkillsDir:   "skills",
			CommandsDir: "commands",
		},
		{
			Name:        AgentCrush,
			DisplayName: "Crush",
			ConfigDir:   ".crush",
			SkillsDir:   "skills",
			CommandsDir: "commands",
		},
		{
			Name:        AgentCursor,
			DisplayName: "Cursor",
			ConfigDir:   ".cursor",
			SkillsDir:   "skills",
			CommandsDir: "commands",
		},
		{
			Name:        AgentWindsurf,
			DisplayName: "Windsurf",
			ConfigDir:   ".windsurf",
			SkillsDir:   "skills",
			CommandsDir: "commands",
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
