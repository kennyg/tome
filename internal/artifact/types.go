package artifact

import "time"

// Type represents the kind of artifact
type Type string

const (
	TypeSkill   Type = "skill"
	TypeCommand Type = "command"
	TypePrompt  Type = "prompt"
	TypeHook    Type = "hook"
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
	Globs []string `yaml:"globs,omitempty" json:"globs,omitempty"`

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

// Manifest represents the tome.yaml file in a repository
type Manifest struct {
	Name        string     `yaml:"name"`
	Description string     `yaml:"description,omitempty"`
	Author      string     `yaml:"author,omitempty"`
	Version     string     `yaml:"version,omitempty"`
	Artifacts   []Artifact `yaml:"artifacts,omitempty"`
}

// InstalledArtifact tracks what's been installed
type InstalledArtifact struct {
	Artifact
	LocalPath string    `json:"local_path"`
	Hash      string    `json:"hash,omitempty"` // For update detection
}
