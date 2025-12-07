package artifact

import "time"

// Type represents the kind of artifact
type Type string

const (
	TypeSkill   Type = "skill"
	TypeCommand Type = "command"
	TypePrompt  Type = "prompt"
	TypeHook    Type = "hook"
	TypeAgent   Type = "agent"
	TypePlugin  Type = "plugin"
)

// Artifact represents an installable item (skill, command, prompt, or hook)
type Artifact struct {
	// Core metadata
	Name        string `yaml:"name" json:"name"`
	Type        Type   `yaml:"type" json:"type"`
	Description string `yaml:"description" json:"description"`
	Version     string `yaml:"version,omitempty" json:"version,omitempty"`
	Author      string `yaml:"author,omitempty" json:"author,omitempty"`

	// Source information
	Source    string `yaml:"-" json:"source"`              // Where it was installed from
	SourceURL string `yaml:"-" json:"source_url,omitempty"` // Original URL if applicable

	// File information
	Filename string `yaml:"-" json:"filename"`
	Content  string `yaml:"-" json:"-"` // The actual file content

	// Installation metadata
	InstalledAt time.Time `yaml:"-" json:"installed_at,omitempty"`
	UpdatedAt   time.Time `yaml:"-" json:"updated_at,omitempty"`

	// Skill-specific fields
	Globs    []string `yaml:"globs,omitempty" json:"globs,omitempty"`
	Includes []string `yaml:"includes,omitempty" json:"includes,omitempty"` // Files installed with this skill

	// Command-specific fields
	Arguments []Argument `yaml:"arguments,omitempty" json:"arguments,omitempty"`

	// Hook-specific fields
	Event   string `yaml:"event,omitempty" json:"event,omitempty"`
	Command string `yaml:"command,omitempty" json:"command,omitempty"`
}

// Argument represents a command argument
type Argument struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Required    bool   `yaml:"required,omitempty" json:"required,omitempty"`
}

// ArtifactSummary is a minimal artifact representation for manifests
type ArtifactSummary struct {
	Name        string `yaml:"name" json:"name"`
	Type        Type   `yaml:"type" json:"type"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Hash        string `yaml:"hash,omitempty" json:"hash,omitempty"` // sha256:... for integrity verification
}

// Manifest represents the tome.yaml file in a repository
type Manifest struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Author      string   `yaml:"author,omitempty" json:"author,omitempty"`
	Version     string   `yaml:"version,omitempty" json:"version,omitempty"`
	License     string   `yaml:"license,omitempty" json:"license,omitempty"`
	Source      string   `yaml:"source,omitempty" json:"source,omitempty"`
	Homepage    string   `yaml:"homepage,omitempty" json:"homepage,omitempty"`
	Tags        []string `yaml:"tags,omitempty" json:"tags,omitempty"`

	// Optional custom paths (defaults: commands/, skills/)
	CommandsDir string `yaml:"commands_dir,omitempty" json:"commands_dir,omitempty"`
	SkillsDir   string `yaml:"skills_dir,omitempty" json:"skills_dir,omitempty"`

	// Artifact index (written by 'tome bind --write')
	Artifacts []ArtifactSummary `yaml:"artifacts,omitempty" json:"artifacts,omitempty"`
}

// InstalledArtifact tracks what's been installed
type InstalledArtifact struct {
	Artifact
	LocalPath string `json:"local_path"`
	Hash      string `json:"hash,omitempty"` // For update detection
}

// PluginManifest represents .claude-plugin/plugin.json
type PluginManifest struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Version     string       `json:"version,omitempty"`
	Author      PluginAuthor `json:"author,omitempty"`
	Repository  *PluginRepo  `json:"repository,omitempty"`
}

// PluginAuthor represents plugin author info
type PluginAuthor struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	URL   string `json:"url,omitempty"`
}

// PluginRepo represents plugin repository info
type PluginRepo struct {
	Type string `json:"type,omitempty"`
	URL  string `json:"url,omitempty"`
}

// Plugin represents a complete plugin with all its artifacts
type Plugin struct {
	Manifest PluginManifest
	Source   string // Where it was fetched from

	// Extracted artifacts
	Skills   []Artifact
	Commands []Artifact
	Agents   []Artifact
	Hooks    []Artifact
}
