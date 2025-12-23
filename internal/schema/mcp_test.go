package schema

import (
	"encoding/json"
	"testing"
)

func TestParseClaudeMCP(t *testing.T) {
	input := `{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"],
      "env": {
        "DEBUG": "true"
      }
    },
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "ghp_xxx"
      },
      "disabled": true
    }
  }
}`

	config, err := ParseClaudeMCP([]byte(input))
	if err != nil {
		t.Fatalf("ParseClaudeMCP failed: %v", err)
	}

	if config.GetFormat() != FormatClaude {
		t.Errorf("format = %v, want %v", config.GetFormat(), FormatClaude)
	}

	if len(config.Servers) != 2 {
		t.Errorf("server count = %d, want 2", len(config.Servers))
	}

	fs := config.Servers["filesystem"]
	if fs == nil {
		t.Fatal("missing filesystem server")
	}
	if fs.Command != "npx" {
		t.Errorf("command = %q, want %q", fs.Command, "npx")
	}
	if len(fs.Args) != 3 {
		t.Errorf("args count = %d, want 3", len(fs.Args))
	}
	if fs.Env["DEBUG"] != "true" {
		t.Errorf("env DEBUG = %q, want %q", fs.Env["DEBUG"], "true")
	}

	gh := config.Servers["github"]
	if gh == nil {
		t.Fatal("missing github server")
	}
	if !gh.Disabled {
		t.Error("github server should be disabled")
	}
}

func TestParseOpenCodeMCP(t *testing.T) {
	input := `{
  "mcp": {
    "filesystem": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-filesystem", "/tmp"],
      "enabled": true,
      "environment": {
        "DEBUG": "true"
      }
    },
    "remote-api": {
      "type": "remote",
      "url": "https://api.example.com/mcp",
      "headers": {
        "Authorization": "Bearer xxx"
      }
    }
  }
}`

	config, err := ParseOpenCodeMCP([]byte(input))
	if err != nil {
		t.Fatalf("ParseOpenCodeMCP failed: %v", err)
	}

	if config.GetFormat() != FormatOpenCode {
		t.Errorf("format = %v, want %v", config.GetFormat(), FormatOpenCode)
	}

	if len(config.Servers) != 2 {
		t.Errorf("server count = %d, want 2", len(config.Servers))
	}

	fs := config.Servers["filesystem"]
	if fs == nil {
		t.Fatal("missing filesystem server")
	}
	if fs.Command != "npx" {
		t.Errorf("command = %q, want %q", fs.Command, "npx")
	}
	if len(fs.Args) != 3 {
		t.Errorf("args count = %d, want 3", len(fs.Args))
	}
	if fs.Type != "local" {
		t.Errorf("type = %q, want %q", fs.Type, "local")
	}
	if fs.Enabled == nil || !*fs.Enabled {
		t.Error("filesystem server should be enabled")
	}

	remote := config.Servers["remote-api"]
	if remote == nil {
		t.Fatal("missing remote-api server")
	}
	if remote.Type != "remote" {
		t.Errorf("type = %q, want %q", remote.Type, "remote")
	}
	if remote.URL != "https://api.example.com/mcp" {
		t.Errorf("url = %q, want %q", remote.URL, "https://api.example.com/mcp")
	}
	if remote.Headers["Authorization"] != "Bearer xxx" {
		t.Errorf("header = %q, want %q", remote.Headers["Authorization"], "Bearer xxx")
	}
}

func TestSerializeClaudeMCP(t *testing.T) {
	config := &MCPConfig{
		Servers: map[string]*MCPServer{
			"test": {
				Name:    "test",
				Command: "npx",
				Args:    []string{"-y", "test-server"},
				Env:     map[string]string{"KEY": "value"},
			},
		},
		sourceFormat: FormatClaude,
	}

	data, err := SerializeClaudeMCP(config)
	if err != nil {
		t.Fatalf("SerializeClaudeMCP failed: %v", err)
	}

	// Parse it back to verify
	var parsed ClaudeMCPConfig
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to parse serialized output: %v", err)
	}

	server := parsed.MCPServers["test"]
	if server == nil {
		t.Fatal("missing test server in output")
	}
	if server.Command != "npx" {
		t.Errorf("command = %q, want %q", server.Command, "npx")
	}
	if len(server.Args) != 2 {
		t.Errorf("args count = %d, want 2", len(server.Args))
	}
}

func TestSerializeOpenCodeMCP(t *testing.T) {
	enabled := true
	config := &MCPConfig{
		Servers: map[string]*MCPServer{
			"test": {
				Name:    "test",
				Command: "npx",
				Args:    []string{"-y", "test-server"},
				Env:     map[string]string{"KEY": "value"},
				Enabled: &enabled,
			},
		},
		sourceFormat: FormatOpenCode,
	}

	data, err := SerializeOpenCodeMCP(config)
	if err != nil {
		t.Fatalf("SerializeOpenCodeMCP failed: %v", err)
	}

	// Parse it back to verify
	var parsed OpenCodeMCPConfig
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to parse serialized output: %v", err)
	}

	server := parsed.MCP["test"]
	if server == nil {
		t.Fatal("missing test server in output")
	}
	if server.Type != "local" {
		t.Errorf("type = %q, want %q", server.Type, "local")
	}
	if len(server.Command) != 3 {
		t.Errorf("command count = %d, want 3", len(server.Command))
	}
	if server.Command[0] != "npx" {
		t.Errorf("command[0] = %q, want %q", server.Command[0], "npx")
	}
	if server.Environment["KEY"] != "value" {
		t.Errorf("environment KEY = %q, want %q", server.Environment["KEY"], "value")
	}
}

func TestConvertClaudeToOpenCode(t *testing.T) {
	input := `{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem"],
      "env": {
        "DEBUG": "true"
      }
    }
  }
}`

	config, err := ParseClaudeMCP([]byte(input))
	if err != nil {
		t.Fatalf("ParseClaudeMCP failed: %v", err)
	}

	output, err := ConvertMCP(config, FormatOpenCode)
	if err != nil {
		t.Fatalf("ConvertMCP failed: %v", err)
	}

	// Parse output as OpenCode format
	var parsed OpenCodeMCPConfig
	if err := json.Unmarshal(output, &parsed); err != nil {
		t.Fatalf("failed to parse converted output: %v", err)
	}

	server := parsed.MCP["filesystem"]
	if server == nil {
		t.Fatal("missing filesystem server in output")
	}
	if server.Type != "local" {
		t.Errorf("type = %q, want %q", server.Type, "local")
	}
	// Command should be combined array
	if len(server.Command) != 3 {
		t.Errorf("command count = %d, want 3", len(server.Command))
	}
	// Env should be in environment field
	if server.Environment["DEBUG"] != "true" {
		t.Errorf("environment DEBUG = %q, want %q", server.Environment["DEBUG"], "true")
	}
}

func TestConvertOpenCodeToClaude(t *testing.T) {
	input := `{
  "mcp": {
    "filesystem": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-filesystem"],
      "environment": {
        "DEBUG": "true"
      }
    }
  }
}`

	config, err := ParseOpenCodeMCP([]byte(input))
	if err != nil {
		t.Fatalf("ParseOpenCodeMCP failed: %v", err)
	}

	output, err := ConvertMCP(config, FormatClaude)
	if err != nil {
		t.Fatalf("ConvertMCP failed: %v", err)
	}

	// Parse output as Claude format
	var parsed ClaudeMCPConfig
	if err := json.Unmarshal(output, &parsed); err != nil {
		t.Fatalf("failed to parse converted output: %v", err)
	}

	server := parsed.MCPServers["filesystem"]
	if server == nil {
		t.Fatal("missing filesystem server in output")
	}
	if server.Command != "npx" {
		t.Errorf("command = %q, want %q", server.Command, "npx")
	}
	if len(server.Args) != 2 {
		t.Errorf("args count = %d, want 2", len(server.Args))
	}
	if server.Env["DEBUG"] != "true" {
		t.Errorf("env DEBUG = %q, want %q", server.Env["DEBUG"], "true")
	}
}

func TestConvertMCPWithInfo_Warnings(t *testing.T) {
	// OpenCode config with remote server and headers
	enabled := true
	config := &MCPConfig{
		Servers: map[string]*MCPServer{
			"remote": {
				Name:    "remote",
				Type:    "remote",
				URL:     "https://api.example.com",
				Headers: map[string]string{"Auth": "token"},
				Enabled: &enabled,
			},
		},
		sourceFormat: FormatOpenCode,
	}

	result, err := ConvertMCPWithInfo(config, FormatClaude)
	if err != nil {
		t.Fatalf("ConvertMCPWithInfo failed: %v", err)
	}

	// Should have warnings about OpenCode-specific fields
	if len(result.Warnings) == 0 {
		t.Error("expected warnings about OpenCode-specific fields")
	}

	hasURLWarning := false
	hasHeadersWarning := false
	hasEnabledWarning := false
	for _, w := range result.Warnings {
		if contains(w, "URL") {
			hasURLWarning = true
		}
		if contains(w, "headers") {
			hasHeadersWarning = true
		}
		if contains(w, "enabled") {
			hasEnabledWarning = true
		}
	}

	if !hasURLWarning {
		t.Error("missing warning about URL field")
	}
	if !hasHeadersWarning {
		t.Error("missing warning about headers field")
	}
	if !hasEnabledWarning {
		t.Error("missing warning about enabled field")
	}
}

func TestDetectMCPFormat(t *testing.T) {
	tests := []struct {
		filename string
		want     Format
	}{
		{".mcp.json", FormatClaude},
		{"project/.mcp.json", FormatClaude},
		{".claude/settings.local.json", FormatClaude},
		{".cursor/mcp.json", FormatCursor},
		{"~/.cursor/mcp.json", FormatCursor},
		{"opencode.json", FormatOpenCode},
		{"~/.config/opencode/opencode.json", FormatOpenCode},
		{"random.json", FormatClaude}, // default
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := DetectMCPFormat(tt.filename)
			if got != tt.want {
				t.Errorf("DetectMCPFormat(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestIsMCPFile(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{".mcp.json", true},
		{".cursor/mcp.json", true},
		{"opencode.json", true},
		{".claude.json", true},
		{".claude/settings.local.json", true},
		{"SKILL.md", false},
		{"random.json", false},
		{"mcp.json", false}, // Only .cursor/mcp.json, not bare mcp.json
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := IsMCPFile(tt.filename)
			if got != tt.want {
				t.Errorf("IsMCPFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestMCPOutputFilename(t *testing.T) {
	tests := []struct {
		format Format
		want   string
	}{
		{FormatClaude, ".mcp.json"},
		{FormatCursor, "mcp.json"},
		{FormatOpenCode, "opencode.json"},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			got := MCPOutputFilename(tt.format)
			if got != tt.want {
				t.Errorf("MCPOutputFilename(%v) = %q, want %q", tt.format, got, tt.want)
			}
		})
	}
}

func TestMergeMCPConfigs(t *testing.T) {
	config1 := &MCPConfig{
		Servers: map[string]*MCPServer{
			"server1": {Name: "server1", Command: "cmd1"},
			"server2": {Name: "server2", Command: "cmd2"},
		},
		sourceFormat: FormatClaude,
	}

	config2 := &MCPConfig{
		Servers: map[string]*MCPServer{
			"server2": {Name: "server2", Command: "cmd2-override"},
			"server3": {Name: "server3", Command: "cmd3"},
		},
		sourceFormat: FormatOpenCode,
	}

	merged := MergeMCPConfigs(config1, config2)

	if len(merged.Servers) != 3 {
		t.Errorf("merged server count = %d, want 3", len(merged.Servers))
	}

	// server2 should be overridden
	if merged.Servers["server2"].Command != "cmd2-override" {
		t.Errorf("server2 command = %q, want %q", merged.Servers["server2"].Command, "cmd2-override")
	}

	// Format should be from last config
	if merged.GetFormat() != FormatOpenCode {
		t.Errorf("format = %v, want %v", merged.GetFormat(), FormatOpenCode)
	}
}

func TestServerNames(t *testing.T) {
	config := &MCPConfig{
		Servers: map[string]*MCPServer{
			"zebra":   {Name: "zebra"},
			"alpha":   {Name: "alpha"},
			"middle":  {Name: "middle"},
		},
	}

	names := config.ServerNames()
	expected := []string{"alpha", "middle", "zebra"}

	if len(names) != len(expected) {
		t.Fatalf("names count = %d, want %d", len(names), len(expected))
	}

	for i, name := range names {
		if name != expected[i] {
			t.Errorf("names[%d] = %q, want %q", i, name, expected[i])
		}
	}
}

func TestRoundTrip_ClaudeToOpenCodeAndBack(t *testing.T) {
	original := `{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem"],
      "env": {
        "PATH_PREFIX": "/home/user"
      }
    }
  }
}`

	// Parse as Claude
	config, err := ParseClaudeMCP([]byte(original))
	if err != nil {
		t.Fatalf("ParseClaudeMCP failed: %v", err)
	}

	// Convert to OpenCode
	openCodeData, err := ConvertMCP(config, FormatOpenCode)
	if err != nil {
		t.Fatalf("ConvertMCP to OpenCode failed: %v", err)
	}

	// Parse as OpenCode
	config2, err := ParseOpenCodeMCP(openCodeData)
	if err != nil {
		t.Fatalf("ParseOpenCodeMCP failed: %v", err)
	}

	// Convert back to Claude
	claudeData, err := ConvertMCP(config2, FormatClaude)
	if err != nil {
		t.Fatalf("ConvertMCP back to Claude failed: %v", err)
	}

	// Parse final result
	config3, err := ParseClaudeMCP(claudeData)
	if err != nil {
		t.Fatalf("ParseClaudeMCP of round-trip failed: %v", err)
	}

	// Verify data preserved
	server := config3.Servers["filesystem"]
	if server == nil {
		t.Fatal("missing filesystem server after round-trip")
	}
	if server.Command != "npx" {
		t.Errorf("command = %q, want %q", server.Command, "npx")
	}
	if len(server.Args) != 2 {
		t.Errorf("args count = %d, want 2", len(server.Args))
	}
	if server.Env["PATH_PREFIX"] != "/home/user" {
		t.Errorf("env PATH_PREFIX = %q, want %q", server.Env["PATH_PREFIX"], "/home/user")
	}
}
