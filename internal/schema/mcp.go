package schema

import (
	"encoding/json"
	"fmt"
	"sort"
)

// MCPServer represents a single MCP server configuration
// This is the canonical internal representation used for conversion
type MCPServer struct {
	Name        string            `json:"-"`                       // Server name (used as key)
	Command     string            `json:"command,omitempty"`       // Executable command
	Args        []string          `json:"args,omitempty"`          // Command arguments
	Env         map[string]string `json:"env,omitempty"`           // Environment variables
	Type        string            `json:"type,omitempty"`          // "local" or "remote" (OpenCode)
	URL         string            `json:"url,omitempty"`           // Remote server URL (OpenCode)
	Headers     map[string]string `json:"headers,omitempty"`       // HTTP headers (OpenCode remote)
	Enabled     *bool             `json:"enabled,omitempty"`       // Enabled state (OpenCode)
	Disabled    bool              `json:"disabled,omitempty"`      // Disabled state (Claude)
	Timeout     int               `json:"timeout,omitempty"`       // Timeout in seconds
	Description string            `json:"description,omitempty"`   // Optional description
}

// MCPConfig represents a collection of MCP servers
type MCPConfig struct {
	Servers      map[string]*MCPServer
	sourceFormat Format
}

// GetFormat returns the source format
func (c *MCPConfig) GetFormat() Format {
	return c.sourceFormat
}

// SetFormat sets the source format
func (c *MCPConfig) SetFormat(f Format) {
	c.sourceFormat = f
}

// ServerNames returns sorted server names for deterministic output
func (c *MCPConfig) ServerNames() []string {
	names := make([]string, 0, len(c.Servers))
	for name := range c.Servers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ClaudeMCPConfig represents Claude Code's MCP configuration format
// Used in ~/.claude.json, .claude/settings.local.json, .mcp.json
type ClaudeMCPConfig struct {
	MCPServers map[string]*ClaudeMCPServer `json:"mcpServers,omitempty"`
}

// ClaudeMCPServer represents a server in Claude's format
type ClaudeMCPServer struct {
	Command  string            `json:"command,omitempty"`
	Args     []string          `json:"args,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
	Type     string            `json:"type,omitempty"`     // "stdio" (default) or others
	Disabled bool              `json:"disabled,omitempty"` // Disabled state
	Timeout  int               `json:"timeout,omitempty"`  // Timeout in seconds
}

// CursorMCPConfig represents Cursor's MCP configuration format
// Used in ~/.cursor/mcp.json, .cursor/mcp.json
// Format is identical to Claude's
type CursorMCPConfig = ClaudeMCPConfig

// CursorMCPServer is identical to Claude's format
type CursorMCPServer = ClaudeMCPServer

// CopilotMCPConfig represents VS Code/GitHub Copilot's MCP configuration format
// Used in .vscode/mcp.json
type CopilotMCPConfig struct {
	Servers map[string]*CopilotMCPServer `json:"servers,omitempty"`
	Inputs  []CopilotMCPInput            `json:"inputs,omitempty"`
}

// CopilotMCPServer represents a server in VS Code/Copilot's format
type CopilotMCPServer struct {
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	EnvFile string            `json:"envFile,omitempty"` // Path to env file
	Type    string            `json:"type,omitempty"`    // "stdio", "http", "sse"
	URL     string            `json:"url,omitempty"`     // For http/sse types
	Headers map[string]string `json:"headers,omitempty"` // For http/sse types
}

// CopilotMCPInput represents an input placeholder for secrets
type CopilotMCPInput struct {
	ID          string `json:"id"`
	Type        string `json:"type,omitempty"`        // "promptString"
	Description string `json:"description,omitempty"`
	Password    bool   `json:"password,omitempty"`
}

// OpenCodeMCPConfig represents OpenCode's MCP configuration format
// Used in ~/.config/opencode/opencode.json, opencode.json
type OpenCodeMCPConfig struct {
	MCP map[string]*OpenCodeMCPServer `json:"mcp,omitempty"`
}

// OpenCodeMCPServer represents a server in OpenCode's format
type OpenCodeMCPServer struct {
	Type        string            `json:"type,omitempty"`        // "local" or "remote"
	Command     []string          `json:"command,omitempty"`     // Command as array
	Environment map[string]string `json:"environment,omitempty"` // Env vars (different key)
	Enabled     *bool             `json:"enabled,omitempty"`     // Enabled state
	URL         string            `json:"url,omitempty"`         // For remote servers
	Headers     map[string]string `json:"headers,omitempty"`     // For remote servers
}

// ParseClaudeMCP parses Claude Code MCP configuration
func ParseClaudeMCP(content []byte) (*MCPConfig, error) {
	var cfg ClaudeMCPConfig
	if err := json.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse Claude MCP config: %w", err)
	}

	config := &MCPConfig{
		Servers:      make(map[string]*MCPServer),
		sourceFormat: FormatClaude,
	}

	for name, server := range cfg.MCPServers {
		config.Servers[name] = &MCPServer{
			Name:     name,
			Command:  server.Command,
			Args:     server.Args,
			Env:      server.Env,
			Type:     server.Type,
			Disabled: server.Disabled,
			Timeout:  server.Timeout,
		}
	}

	return config, nil
}

// ParseCursorMCP parses Cursor MCP configuration (same format as Claude)
func ParseCursorMCP(content []byte) (*MCPConfig, error) {
	config, err := ParseClaudeMCP(content)
	if err != nil {
		return nil, err
	}
	config.sourceFormat = FormatCursor
	return config, nil
}

// ParseOpenCodeMCP parses OpenCode MCP configuration
func ParseOpenCodeMCP(content []byte) (*MCPConfig, error) {
	var cfg OpenCodeMCPConfig
	if err := json.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse OpenCode MCP config: %w", err)
	}

	config := &MCPConfig{
		Servers:      make(map[string]*MCPServer),
		sourceFormat: FormatOpenCode,
	}

	for name, server := range cfg.MCP {
		srv := &MCPServer{
			Name:    name,
			Type:    server.Type,
			Env:     server.Environment,
			URL:     server.URL,
			Headers: server.Headers,
			Enabled: server.Enabled,
		}

		// Convert command array to command + args
		if len(server.Command) > 0 {
			srv.Command = server.Command[0]
			if len(server.Command) > 1 {
				srv.Args = server.Command[1:]
			}
		}

		config.Servers[name] = srv
	}

	return config, nil
}

// ParseCopilotMCP parses VS Code/GitHub Copilot MCP configuration
func ParseCopilotMCP(content []byte) (*MCPConfig, error) {
	var cfg CopilotMCPConfig
	if err := json.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse Copilot MCP config: %w", err)
	}

	config := &MCPConfig{
		Servers:      make(map[string]*MCPServer),
		sourceFormat: FormatCopilot,
	}

	for name, server := range cfg.Servers {
		config.Servers[name] = &MCPServer{
			Name:    name,
			Command: server.Command,
			Args:    server.Args,
			Env:     server.Env,
			Type:    server.Type,
			URL:     server.URL,
			Headers: server.Headers,
		}
	}

	return config, nil
}

// ParseMCP parses MCP configuration based on format
func ParseMCP(content []byte, format Format) (*MCPConfig, error) {
	switch format {
	case FormatClaude:
		return ParseClaudeMCP(content)
	case FormatCopilot:
		return ParseCopilotMCP(content)
	case FormatCursor:
		return ParseCursorMCP(content)
	case FormatOpenCode:
		return ParseOpenCodeMCP(content)
	default:
		return nil, fmt.Errorf("unsupported MCP format: %s", format)
	}
}

// ParseMCPAuto attempts to detect format and parse MCP config
func ParseMCPAuto(content []byte, filename string) (*MCPConfig, error) {
	format := DetectMCPFormat(filename)
	return ParseMCP(content, format)
}

// DetectMCPFormat detects the format from filename
func DetectMCPFormat(filename string) Format {
	switch {
	case contains(filename, ".vscode"):
		return FormatCopilot
	case contains(filename, ".cursor"):
		return FormatCursor
	case contains(filename, "opencode"):
		return FormatOpenCode
	case contains(filename, ".claude") || hasBasename(filename, ".mcp.json"):
		return FormatClaude
	default:
		return FormatClaude
	}
}

// SerializeClaudeMCP serializes to Claude/Cursor format
func SerializeClaudeMCP(config *MCPConfig) ([]byte, error) {
	cfg := ClaudeMCPConfig{
		MCPServers: make(map[string]*ClaudeMCPServer),
	}

	for name, server := range config.Servers {
		srv := &ClaudeMCPServer{
			Command:  server.Command,
			Args:     server.Args,
			Env:      server.Env,
			Disabled: server.Disabled,
			Timeout:  server.Timeout,
		}
		// Set type if specified
		if server.Type != "" && server.Type != "local" {
			srv.Type = server.Type
		}
		cfg.MCPServers[name] = srv
	}

	return json.MarshalIndent(cfg, "", "  ")
}

// SerializeCursorMCP serializes to Cursor format (same as Claude)
func SerializeCursorMCP(config *MCPConfig) ([]byte, error) {
	return SerializeClaudeMCP(config)
}

// SerializeOpenCodeMCP serializes to OpenCode format
func SerializeOpenCodeMCP(config *MCPConfig) ([]byte, error) {
	cfg := OpenCodeMCPConfig{
		MCP: make(map[string]*OpenCodeMCPServer),
	}

	for name, server := range config.Servers {
		srv := &OpenCodeMCPServer{
			Environment: server.Env,
			URL:         server.URL,
			Headers:     server.Headers,
			Enabled:     server.Enabled,
		}

		// Set type (default to "local" if not specified)
		if server.Type != "" {
			srv.Type = server.Type
		} else if server.URL != "" {
			srv.Type = "remote"
		} else {
			srv.Type = "local"
		}

		// Combine command and args into command array
		if server.Command != "" {
			srv.Command = append([]string{server.Command}, server.Args...)
		}

		cfg.MCP[name] = srv
	}

	return json.MarshalIndent(cfg, "", "  ")
}

// SerializeCopilotMCP serializes to VS Code/Copilot format
func SerializeCopilotMCP(config *MCPConfig) ([]byte, error) {
	cfg := CopilotMCPConfig{
		Servers: make(map[string]*CopilotMCPServer),
	}

	for name, server := range config.Servers {
		srv := &CopilotMCPServer{
			Command: server.Command,
			Args:    server.Args,
			Env:     server.Env,
			URL:     server.URL,
			Headers: server.Headers,
		}
		// Set type for remote servers
		if server.URL != "" {
			if server.Type == "sse" {
				srv.Type = "sse"
			} else {
				srv.Type = "http"
			}
		}
		cfg.Servers[name] = srv
	}

	return json.MarshalIndent(cfg, "", "  ")
}

// SerializeMCP serializes to the specified format
func SerializeMCP(config *MCPConfig, format Format) ([]byte, error) {
	switch format {
	case FormatClaude:
		return SerializeClaudeMCP(config)
	case FormatCursor:
		return SerializeCursorMCP(config)
	case FormatCopilot:
		return SerializeCopilotMCP(config)
	case FormatOpenCode:
		return SerializeOpenCodeMCP(config)
	default:
		return nil, fmt.Errorf("unsupported MCP format: %s", format)
	}
}

// ConvertMCP converts MCP config from one format to another
func ConvertMCP(config *MCPConfig, targetFormat Format) ([]byte, error) {
	return SerializeMCP(config, targetFormat)
}

// MCPConversionResult holds the result of an MCP conversion
type MCPConversionResult struct {
	SourceFormat Format
	TargetFormat Format
	ServerCount  int
	Content      []byte
	Warnings     []string
}

// ConvertMCPWithInfo converts MCP config and returns detailed information
func ConvertMCPWithInfo(config *MCPConfig, targetFormat Format) (*MCPConversionResult, error) {
	content, err := ConvertMCP(config, targetFormat)
	if err != nil {
		return nil, err
	}

	result := &MCPConversionResult{
		SourceFormat: config.sourceFormat,
		TargetFormat: targetFormat,
		ServerCount:  len(config.Servers),
		Content:      content,
	}

	// Check for potential data loss
	for name, server := range config.Servers {
		// OpenCode-specific fields
		if targetFormat != FormatOpenCode && targetFormat != FormatCopilot {
			if server.URL != "" {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("server %q: remote URL not supported in %s (will be omitted)", name, targetFormat))
			}
			if server.Headers != nil && len(server.Headers) > 0 {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("server %q: headers not supported in %s (will be omitted)", name, targetFormat))
			}
		}
		if targetFormat != FormatOpenCode {
			if server.Enabled != nil {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("server %q: enabled field is OpenCode-specific (will be omitted)", name))
			}
		}

		// Claude-specific fields
		if targetFormat == FormatOpenCode || targetFormat == FormatCopilot {
			if server.Disabled {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("server %q: disabled field is Claude-specific (will be omitted)", name))
			}
			if server.Timeout > 0 {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("server %q: timeout field is Claude-specific (will be omitted)", name))
			}
		}
	}

	return result, nil
}

// MCPOutputFilename returns the appropriate filename for MCP config
func MCPOutputFilename(targetFormat Format) string {
	switch targetFormat {
	case FormatClaude:
		return ".mcp.json"
	case FormatCursor:
		return "mcp.json"
	case FormatCopilot:
		return "mcp.json"
	case FormatOpenCode:
		return "opencode.json"
	default:
		return "mcp.json"
	}
}

// MCPOutputDirectory returns the appropriate directory for MCP config
func MCPOutputDirectory(targetFormat Format) string {
	switch targetFormat {
	case FormatClaude:
		return "" // .mcp.json goes in project root
	case FormatCursor:
		return ".cursor"
	case FormatCopilot:
		return ".vscode"
	case FormatOpenCode:
		return "" // opencode.json goes in project root
	default:
		return ""
	}
}

// IsMCPFile checks if a filename is an MCP configuration file
func IsMCPFile(filename string) bool {
	switch {
	case hasBasename(filename, ".mcp.json"):
		return true
	case hasBasename(filename, "mcp.json") && contains(filename, ".cursor"):
		return true
	case hasBasename(filename, "mcp.json") && contains(filename, ".vscode"):
		return true
	case hasBasename(filename, "opencode.json"):
		return true
	case hasBasename(filename, ".claude.json"):
		return true
	case contains(filename, "settings.local.json"):
		return true
	default:
		return false
	}
}

// MergeMCPConfigs merges multiple MCP configs into one
// Later configs override earlier ones for the same server name
func MergeMCPConfigs(configs ...*MCPConfig) *MCPConfig {
	merged := &MCPConfig{
		Servers: make(map[string]*MCPServer),
	}

	for _, cfg := range configs {
		if cfg == nil {
			continue
		}
		for name, server := range cfg.Servers {
			merged.Servers[name] = server
		}
		// Use the last config's format
		merged.sourceFormat = cfg.sourceFormat
	}

	return merged
}
