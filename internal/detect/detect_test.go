package detect

import (
	"testing"
)

func TestFromContent_NPM(t *testing.T) {
	content := `
# Installation

Run the following:

` + "```bash" + `
npm install opencode-orchestrator
bun add some-package
yarn add another-package
` + "```" + `
`
	reqs := FromContent(content)

	// Should find npm packages
	found := make(map[string]bool)
	for _, req := range reqs {
		if req.Type == TypeNPM {
			found[req.Value] = true
		}
	}

	if !found["opencode-orchestrator"] {
		t.Error("expected to find opencode-orchestrator")
	}
	if !found["some-package"] {
		t.Error("expected to find some-package")
	}
	if !found["another-package"] {
		t.Error("expected to find another-package")
	}
}

func TestFromContent_Pip(t *testing.T) {
	content := `
# Setup

` + "```bash" + `
pip install requests
pip3 install numpy
python -m pip install pandas
` + "```" + `
`
	reqs := FromContent(content)

	found := make(map[string]bool)
	for _, req := range reqs {
		if req.Type == TypePip {
			found[req.Value] = true
		}
	}

	if !found["requests"] {
		t.Error("expected to find requests")
	}
	if !found["numpy"] {
		t.Error("expected to find numpy")
	}
	if !found["pandas"] {
		t.Error("expected to find pandas")
	}
}

func TestFromContent_EnvVars(t *testing.T) {
	content := `
# Configuration

Set your API key:

export OPENAI_API_KEY=your-key-here

You can also use $ANTHROPIC_API_KEY for Claude.
`
	reqs := FromContent(content)

	found := make(map[string]bool)
	for _, req := range reqs {
		if req.Type == TypeEnv {
			found[req.Value] = true
		}
	}

	if !found["OPENAI_API_KEY"] {
		t.Error("expected to find OPENAI_API_KEY")
	}
	if !found["ANTHROPIC_API_KEY"] {
		t.Error("expected to find ANTHROPIC_API_KEY")
	}
}

func TestFromContent_IgnoresSystemEnvVars(t *testing.T) {
	content := `
The command runs in $HOME directory.
Check your $PATH settings.
`
	reqs := FromContent(content)

	for _, req := range reqs {
		if req.Type == TypeEnv && (req.Value == "HOME" || req.Value == "PATH") {
			t.Errorf("should not detect system env var: %s", req.Value)
		}
	}
}

func TestFromContent_Brew(t *testing.T) {
	content := `
# Prerequisites

` + "```bash" + `
brew install jq
brew install fzf
` + "```" + `
`
	reqs := FromContent(content)

	found := make(map[string]bool)
	for _, req := range reqs {
		if req.Type == TypeBrew {
			found[req.Value] = true
		}
	}

	if !found["jq"] {
		t.Error("expected to find jq")
	}
	if !found["fzf"] {
		t.Error("expected to find fzf")
	}
}

func TestFromIncludes_Python(t *testing.T) {
	includes := []string{"helper.py", "utils/parser.py"}
	reqs := FromIncludes(includes)

	found := false
	for _, req := range reqs {
		if req.Type == TypeRuntime && req.Value == "python3" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected to detect python3 runtime requirement")
	}

	// Should only have one python requirement (deduplicated)
	count := 0
	for _, req := range reqs {
		if req.Type == TypeRuntime && req.Value == "python3" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 python3 requirement, got %d", count)
	}
}

func TestFromIncludes_Node(t *testing.T) {
	includes := []string{"index.js", "lib/utils.ts"}
	reqs := FromIncludes(includes)

	found := false
	for _, req := range reqs {
		if req.Type == TypeRuntime && req.Value == "node" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected to detect node runtime requirement")
	}
}

func TestMerge_Deduplicates(t *testing.T) {
	contentReqs := []Requirement{
		{Type: TypeNPM, Value: "foo"},
		{Type: TypeEnv, Value: "API_KEY"},
	}
	includeReqs := []Requirement{
		{Type: TypeNPM, Value: "foo"},      // duplicate
		{Type: TypeRuntime, Value: "node"}, // new
	}

	merged := Merge(contentReqs, includeReqs)

	if len(merged) != 3 {
		t.Errorf("expected 3 merged requirements, got %d", len(merged))
	}

	// Count occurrences
	counts := make(map[string]int)
	for _, req := range merged {
		key := string(req.Type) + ":" + req.Value
		counts[key]++
	}

	if counts["npm:foo"] != 1 {
		t.Error("expected npm:foo to appear exactly once")
	}
}

func TestFromContent_LineNumbers(t *testing.T) {
	content := `Line 1
Line 2
npm install test-package
Line 4`

	reqs := FromContent(content)

	for _, req := range reqs {
		if req.Type == TypeNPM && req.Value == "test-package" {
			if req.Line != 3 {
				t.Errorf("expected line 3, got %d", req.Line)
			}
			return
		}
	}
	t.Error("expected to find test-package requirement")
}

func TestFromContent_BunPackageManager(t *testing.T) {
	content := `
# Installation

` + "```bash" + `
bun add opencode-orchestrator
` + "```" + `
`
	reqs := FromContent(content)

	for _, req := range reqs {
		if req.Type == TypeNPM && req.Value == "opencode-orchestrator" {
			if req.PackageManager != PMbun {
				t.Errorf("expected package manager 'bun', got '%s'", req.PackageManager)
			}
			return
		}
	}
	t.Error("expected to find opencode-orchestrator with bun package manager")
}

func TestFromContent_PackageManagerPreserved(t *testing.T) {
	testCases := []struct {
		content    string
		expectedPM PackageManager
		pkg        string
	}{
		{"npm install foo", PMnpm, "foo"},
		{"bun add bar", PMbun, "bar"},
		{"yarn add baz", PMyarn, "baz"},
		{"pnpm add qux", PMpnpm, "qux"},
		{"pip install requests", PMpip, "requests"},
		{"pip3 install numpy", PMpip3, "numpy"},
	}

	for _, tc := range testCases {
		reqs := FromContent(tc.content)
		found := false
		for _, req := range reqs {
			if req.Value == tc.pkg {
				found = true
				if req.PackageManager != tc.expectedPM {
					t.Errorf("for '%s': expected package manager '%s', got '%s'",
						tc.content, tc.expectedPM, req.PackageManager)
				}
				break
			}
		}
		if !found {
			t.Errorf("expected to find package '%s' in '%s'", tc.pkg, tc.content)
		}
	}
}
