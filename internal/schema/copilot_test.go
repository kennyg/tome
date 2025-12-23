package schema

import (
	"strings"
	"testing"
)

func TestParseCopilotAgent(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantName    string
		wantDesc    string
		wantVersion string
		wantBody    string
		wantErr     bool
	}{
		{
			name: "full agent",
			content: `---
name: "C# Expert"
description: An agent for .NET development
version: 2025-01-01
---
You are an expert C# developer.

## Guidelines

- Follow best practices
- Write clean code`,
			wantName:    "C# Expert",
			wantDesc:    "An agent for .NET development",
			wantVersion: "2025-01-01",
			wantBody:    "You are an expert C# developer.\n\n## Guidelines\n\n- Follow best practices\n- Write clean code",
		},
		{
			name: "minimal agent",
			content: `---
name: Test Agent
description: A test
---
Body`,
			wantName: "Test Agent",
			wantDesc: "A test",
			wantBody: "Body",
		},
		{
			name: "agent without version",
			content: `---
name: Simple
description: Simple agent
---
Content here`,
			wantName: "Simple",
			wantDesc: "Simple agent",
			wantBody: "Content here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := ParseCopilotAgent([]byte(tt.content))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCopilotAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if agent.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", agent.Name, tt.wantName)
			}
			if agent.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", agent.Description, tt.wantDesc)
			}
			if agent.Version != tt.wantVersion {
				t.Errorf("Version = %q, want %q", agent.Version, tt.wantVersion)
			}
			if agent.Body != tt.wantBody {
				t.Errorf("Body = %q, want %q", agent.Body, tt.wantBody)
			}
			if agent.GetFormat() != FormatCopilot {
				t.Errorf("GetFormat() = %v, want %v", agent.GetFormat(), FormatCopilot)
			}
		})
	}
}

func TestParseCopilotPrompt(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantAgent string
		wantDesc  string
		wantBody  string
		wantErr   bool
	}{
		{
			name: "full prompt",
			content: `---
agent: agent
description: Interactive prompt refinement workflow
---
You are an AI assistant designed to help users.`,
			wantAgent: "agent",
			wantDesc:  "Interactive prompt refinement workflow",
			wantBody:  "You are an AI assistant designed to help users.",
		},
		{
			name: "prompt without agent field",
			content: `---
description: Just a description
---
Content`,
			wantDesc: "Just a description",
			wantBody: "Content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt, err := ParseCopilotPrompt([]byte(tt.content))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCopilotPrompt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if prompt.Agent != tt.wantAgent {
				t.Errorf("Agent = %q, want %q", prompt.Agent, tt.wantAgent)
			}
			if prompt.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", prompt.Description, tt.wantDesc)
			}
			if prompt.Body != tt.wantBody {
				t.Errorf("Body = %q, want %q", prompt.Body, tt.wantBody)
			}
		})
	}
}

func TestCopilotAgent_Serialize(t *testing.T) {
	agent := &CopilotAgent{
		Name:        "Test Agent",
		Description: "A test agent",
		Version:     "1.0.0",
		Body:        "# Instructions\n\nDo things.",
	}

	got, err := agent.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	content := string(got)

	checks := []string{
		"name: Test Agent",
		"description: A test agent",
		"version: 1.0.0",
		"# Instructions",
		"Do things.",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("Serialized content should contain %q", check)
		}
	}
}

func TestCopilotAgent_Filename(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"C# Expert", "c#-expert.agent.md"},
		{"Simple Agent", "simple-agent.agent.md"},
		{"NoSpaces", "nospaces.agent.md"},
		{"With_Underscore", "with_underscore.agent.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &CopilotAgent{Name: tt.name}
			if got := agent.Filename(); got != tt.want {
				t.Errorf("Filename() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCopilotPrompt_GetName(t *testing.T) {
	tests := []struct {
		agent string
		want  string
	}{
		{"custom-agent", "custom-agent"},
		{"", "prompt"},
	}

	for _, tt := range tests {
		t.Run(tt.agent, func(t *testing.T) {
			prompt := &CopilotPrompt{Agent: tt.agent}
			if got := prompt.GetName(); got != tt.want {
				t.Errorf("GetName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsCopilotFile(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{"CSharp.agent.md", true},
		{"prompts/test.prompt.md", true},
		{"coding.instructions.md", true},
		{"SKILL.md", false},
		{"regular.md", false},
		{"agent.md", false}, // No dot prefix
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			if got := IsCopilotFile(tt.filename); got != tt.want {
				t.Errorf("IsCopilotFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestCopilotAgent_Interface(t *testing.T) {
	agent := &CopilotAgent{
		Name:        "Interface Test",
		Description: "Testing",
		Body:        "Content",
	}

	// Verify it implements Skill
	var _ Skill = agent

	if agent.GetName() != "Interface Test" {
		t.Errorf("GetName() = %q", agent.GetName())
	}
	if agent.GetDescription() != "Testing" {
		t.Errorf("GetDescription() = %q", agent.GetDescription())
	}
	if agent.GetBody() != "Content" {
		t.Errorf("GetBody() = %q", agent.GetBody())
	}
	if agent.GetFormat() != FormatCopilot {
		t.Errorf("GetFormat() = %v", agent.GetFormat())
	}
}
