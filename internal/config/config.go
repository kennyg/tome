package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kennyg/tome/internal/artifact"
)

// Following the dot-config specification: https://dot-config.github.io/
// User config:    ~/.config/tome/ (or $XDG_CONFIG_HOME/tome/)
// Project config: .config/tome/ (in project root)

const (
	// ConfigDir is the subdirectory name under .config
	ConfigDir = "tome"
	// StateFile is the filename for tracking installed artifacts
	StateFile = "state.json"
)

// Paths holds the various paths tome uses
type Paths struct {
	// Home is the user's home directory
	Home string

	// UserConfigDir is ~/.config/tome (or $XDG_CONFIG_HOME/tome)
	UserConfigDir string
	// StateFile is ~/.config/tome/state.json
	StateFile string

	// ProjectConfigDir is .config/tome in the current project (if exists)
	ProjectConfigDir string

	// Agent being used
	Agent Agent

	// Agent-specific paths (where artifacts get installed)
	AgentDir    string // e.g., ~/.claude, ~/.opencode
	SkillsDir   string // e.g., ~/.claude/skills
	CommandsDir string // e.g., ~/.claude/commands
}

// State represents the current installation state
type State struct {
	Version   string                       `json:"version"`
	Installed []artifact.InstalledArtifact `json:"installed"`
}

// GetPaths returns the standard paths for tome
func GetPaths() (*Paths, error) {
	return GetPathsForAgent(DefaultAgent())
}

// GetPathsForAgent returns paths configured for a specific agent
func GetPathsForAgent(agent Agent) (*Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Follow XDG Base Directory spec
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(home, ".config")
	}

	userConfigDir := filepath.Join(configHome, ConfigDir)

	// Check for project-level .config/tome
	projectConfigDir := findProjectConfig()

	// Get agent-specific paths
	agentDir, skillsDir, commandsDir := AgentPaths(home, agent)

	return &Paths{
		Home:             home,
		UserConfigDir:    userConfigDir,
		StateFile:        filepath.Join(userConfigDir, StateFile),
		ProjectConfigDir: projectConfigDir,
		Agent:            agent,
		AgentDir:         agentDir,
		SkillsDir:        skillsDir,
		CommandsDir:      commandsDir,
	}, nil
}

// findProjectConfig looks for .config/tome in the current directory or parents
func findProjectConfig() string {
	projectRoot := findProjectRoot()
	if projectRoot == "" {
		return ""
	}
	candidate := filepath.Join(projectRoot, ".config", ConfigDir)
	if info, err := os.Stat(candidate); err == nil && info.IsDir() {
		return candidate
	}
	return ""
}

// findProjectRoot finds the project root by looking for .config/tome or .git
func findProjectRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Walk up the directory tree
	dir := cwd
	for {
		// Check for .config/tome (attuned project)
		candidate := filepath.Join(dir, ".config", ConfigDir)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return dir
		}

		// Also check for .git to stop at repo root
		gitDir := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached filesystem root
		}
		dir = parent
	}

	return ""
}

// IsAttuned returns true if we're in an attuned project with local agent dirs
func IsAttuned(agent Agent) bool {
	projectRoot := findProjectRoot()
	if projectRoot == "" {
		return false
	}

	cfg := GetAgentConfig(agent)
	if cfg == nil {
		return false
	}

	// Check if project has local agent directories
	skillsDir := filepath.Join(projectRoot, cfg.ConfigDir, cfg.SkillsDir)
	if _, err := os.Stat(skillsDir); err != nil {
		return false
	}

	return true
}

// GetLocalPaths returns paths for project-local installation
func GetLocalPaths(agent Agent) (*Paths, error) {
	projectRoot := findProjectRoot()
	if projectRoot == "" {
		return nil, fmt.Errorf("not in a project directory")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Follow XDG Base Directory spec for user config
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(home, ".config")
	}
	userConfigDir := filepath.Join(configHome, ConfigDir)

	cfg := GetAgentConfig(agent)
	if cfg == nil {
		cfg = GetAgentConfig(AgentClaude)
	}

	// Project-local paths
	projectConfigDir := filepath.Join(projectRoot, ".config", ConfigDir)
	agentDir := filepath.Join(projectRoot, cfg.ConfigDir)
	skillsDir := filepath.Join(agentDir, cfg.SkillsDir)
	commandsDir := filepath.Join(agentDir, cfg.CommandsDir)

	return &Paths{
		Home:             home,
		UserConfigDir:    userConfigDir,
		StateFile:        filepath.Join(projectConfigDir, StateFile), // Project-local state
		ProjectConfigDir: projectConfigDir,
		Agent:            agent,
		AgentDir:         agentDir,
		SkillsDir:        skillsDir,
		CommandsDir:      commandsDir,
	}, nil
}

// EnsureDirs creates all necessary directories
func (p *Paths) EnsureDirs() error {
	dirs := []string{
		p.UserConfigDir,
		p.AgentDir,
		p.SkillsDir,
		p.CommandsDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// HasProjectConfig returns true if a project-level config exists
func (p *Paths) HasProjectConfig() bool {
	return p.ProjectConfigDir != ""
}

// LoadState loads the current state from disk
func LoadState(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{Version: "1"}, nil
		}
		return nil, err
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// SaveState saves the current state to disk
func SaveState(path string, state *State) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// AddInstalled adds an artifact to the installed list
func (s *State) AddInstalled(a artifact.InstalledArtifact) {
	// Remove existing entry with same name and type
	s.RemoveInstalled(a.Name, a.Type)
	s.Installed = append(s.Installed, a)
}

// RemoveInstalled removes an artifact from the installed list
func (s *State) RemoveInstalled(name string, t artifact.Type) {
	filtered := make([]artifact.InstalledArtifact, 0, len(s.Installed))
	for _, a := range s.Installed {
		if !(a.Name == name && a.Type == t) {
			filtered = append(filtered, a)
		}
	}
	s.Installed = filtered
}

// FindInstalled finds an installed artifact by name
func (s *State) FindInstalled(name string) *artifact.InstalledArtifact {
	for i := range s.Installed {
		if s.Installed[i].Name == name {
			return &s.Installed[i]
		}
	}
	return nil
}
