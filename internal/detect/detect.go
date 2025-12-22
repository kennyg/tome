// Package detect provides automatic detection of setup requirements from artifact content.
// It scans markdown content and included files for patterns that indicate
// dependencies, environment variables, and other setup requirements.
package detect

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// RequirementType represents the kind of requirement detected
type RequirementType string

const (
	TypeCommand RequirementType = "command" // Binary must exist on PATH
	TypeNPM     RequirementType = "npm"     // Node.js package (npm, bun, yarn, pnpm)
	TypePip     RequirementType = "pip"     // Python package
	TypeBrew    RequirementType = "brew"    // Homebrew formula
	TypeCargo   RequirementType = "cargo"   // Rust crate
	TypeEnv     RequirementType = "env"     // Environment variable
	TypeRuntime RequirementType = "runtime" // Runtime (node, python, etc.)
)

// PackageManager tracks which package manager was used
type PackageManager string

const (
	PMnpm  PackageManager = "npm"
	PMbun  PackageManager = "bun"
	PMyarn PackageManager = "yarn"
	PMpnpm PackageManager = "pnpm"
	PMpip  PackageManager = "pip"
	PMpip3 PackageManager = "pip3"
)

// Requirement represents a detected setup requirement
type Requirement struct {
	Type           RequirementType `json:"type"`
	Value          string          `json:"value"`                     // Package name, env var name, command name
	Source         string          `json:"source"`                    // Where detected: "content", "include:file.py"
	Line           int             `json:"line"`                      // Line number (0 if not from content)
	Context        string          `json:"context"`                   // The line/snippet where it was found
	PackageManager PackageManager  `json:"package_manager,omitempty"` // Which package manager (npm, bun, yarn, pnpm, pip, pip3)
}

// VerifyResult contains the result of verifying a requirement
type VerifyResult struct {
	Requirement Requirement
	Satisfied   bool
	Message     string // Help message if not satisfied
}

// Patterns for detecting requirements
var (
	// Package manager install patterns - capture the package manager name
	npmInstallRe   = regexp.MustCompile(`(npm)\s+(?:install|i)\s+(?:-[gGdD]\s+)?([a-zA-Z0-9@/_-]+)`)
	bunInstallRe   = regexp.MustCompile(`(bun)\s+(?:add|install)\s+(?:-[gGdD]\s+)?([a-zA-Z0-9@/_-]+)`)
	yarnInstallRe  = regexp.MustCompile(`(yarn)\s+add\s+(?:-[gGdD]\s+)?([a-zA-Z0-9@/_-]+)`)
	pnpmInstallRe  = regexp.MustCompile(`(pnpm)\s+(?:add|install)\s+(?:-[gGdD]\s+)?([a-zA-Z0-9@/_-]+)`)
	pipInstallRe   = regexp.MustCompile(`(pip3?)\s+install\s+([a-zA-Z0-9_-]+)`)
	pythonPipRe    = regexp.MustCompile(`python3?\s+-m\s+(pip)\s+install\s+([a-zA-Z0-9_-]+)`)
	brewInstallRe  = regexp.MustCompile(`brew\s+install\s+([a-zA-Z0-9_-]+)`)
	cargoInstallRe = regexp.MustCompile(`cargo\s+install\s+([a-zA-Z0-9_-]+)`)

	// Environment variable patterns
	envVarRe      = regexp.MustCompile(`\$\{?([A-Z][A-Z0-9_]{2,})\}?`)
	envExportRe   = regexp.MustCompile(`export\s+([A-Z][A-Z0-9_]+)=`)
	apiKeyMention = regexp.MustCompile(`(?i)((?:[A-Z]+_)?API_KEY|(?:[A-Z]+_)?SECRET|(?:[A-Z]+_)?TOKEN)`)

	// Common env vars to ignore (too generic or system-level)
	ignoredEnvVars = map[string]bool{
		"PATH": true, "HOME": true, "USER": true, "SHELL": true,
		"PWD": true, "OLDPWD": true, "TERM": true, "LANG": true,
		"LC_ALL": true, "EDITOR": true, "VISUAL": true,
		"XDG_CONFIG_HOME": true, "XDG_DATA_HOME": true,
		"TMPDIR": true, "TMP": true, "TEMP": true,
	}
)

// FromContent scans markdown/text content for setup requirements
func FromContent(content string) []Requirement {
	var reqs []Requirement
	seen := make(map[string]bool) // Dedupe by type:value

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lineNum := i + 1

		// Check for Node.js package managers (npm, bun, yarn, pnpm)
		// Each captures: [full match, package manager, package name]
		for _, re := range []*regexp.Regexp{npmInstallRe, bunInstallRe, yarnInstallRe, pnpmInstallRe} {
			if matches := re.FindAllStringSubmatch(line, -1); matches != nil {
				for _, m := range matches {
					pm := PackageManager(m[1])
					pkg := m[2]
					key := "npm:" + pkg
					if !seen[key] {
						seen[key] = true
						reqs = append(reqs, Requirement{
							Type:           TypeNPM,
							Value:          pkg,
							Source:         "content",
							Line:           lineNum,
							Context:        strings.TrimSpace(line),
							PackageManager: pm,
						})
					}
				}
			}
		}

		// Check for pip install (captures: [full match, pip/pip3, package])
		if matches := pipInstallRe.FindAllStringSubmatch(line, -1); matches != nil {
			for _, m := range matches {
				pm := PackageManager(m[1])
				pkg := m[2]
				key := "pip:" + pkg
				if !seen[key] {
					seen[key] = true
					reqs = append(reqs, Requirement{
						Type:           TypePip,
						Value:          pkg,
						Source:         "content",
						Line:           lineNum,
						Context:        strings.TrimSpace(line),
						PackageManager: pm,
					})
				}
			}
		}

		// Check for python -m pip install
		if matches := pythonPipRe.FindAllStringSubmatch(line, -1); matches != nil {
			for _, m := range matches {
				pkg := m[2]
				key := "pip:" + pkg
				if !seen[key] {
					seen[key] = true
					reqs = append(reqs, Requirement{
						Type:           TypePip,
						Value:          pkg,
						Source:         "content",
						Line:           lineNum,
						Context:        strings.TrimSpace(line),
						PackageManager: PMpip,
					})
				}
			}
		}

		// Check for brew install
		if matches := brewInstallRe.FindAllStringSubmatch(line, -1); matches != nil {
			for _, m := range matches {
				key := "brew:" + m[1]
				if !seen[key] {
					seen[key] = true
					reqs = append(reqs, Requirement{
						Type:    TypeBrew,
						Value:   m[1],
						Source:  "content",
						Line:    lineNum,
						Context: strings.TrimSpace(line),
					})
				}
			}
		}

		// Check for cargo install
		if matches := cargoInstallRe.FindAllStringSubmatch(line, -1); matches != nil {
			for _, m := range matches {
				key := "cargo:" + m[1]
				if !seen[key] {
					seen[key] = true
					reqs = append(reqs, Requirement{
						Type:    TypeCargo,
						Value:   m[1],
						Source:  "content",
						Line:    lineNum,
						Context: strings.TrimSpace(line),
					})
				}
			}
		}

		// Check for environment variables
		if matches := envVarRe.FindAllStringSubmatch(line, -1); matches != nil {
			for _, m := range matches {
				varName := m[1]
				if ignoredEnvVars[varName] {
					continue
				}
				key := "env:" + varName
				if !seen[key] {
					seen[key] = true
					reqs = append(reqs, Requirement{
						Type:    TypeEnv,
						Value:   varName,
						Source:  "content",
						Line:    lineNum,
						Context: strings.TrimSpace(line),
					})
				}
			}
		}

		// Check for export statements
		if matches := envExportRe.FindAllStringSubmatch(line, -1); matches != nil {
			for _, m := range matches {
				varName := m[1]
				if ignoredEnvVars[varName] {
					continue
				}
				key := "env:" + varName
				if !seen[key] {
					seen[key] = true
					reqs = append(reqs, Requirement{
						Type:    TypeEnv,
						Value:   varName,
						Source:  "content",
						Line:    lineNum,
						Context: strings.TrimSpace(line),
					})
				}
			}
		}

		// Check for API key mentions
		if matches := apiKeyMention.FindAllStringSubmatch(line, -1); matches != nil {
			for _, m := range matches {
				varName := strings.ToUpper(m[1])
				key := "env:" + varName
				if !seen[key] {
					seen[key] = true
					reqs = append(reqs, Requirement{
						Type:    TypeEnv,
						Value:   varName,
						Source:  "content",
						Line:    lineNum,
						Context: strings.TrimSpace(line),
					})
				}
			}
		}
	}

	return reqs
}

// FromIncludes infers requirements from included file types
func FromIncludes(includes []string) []Requirement {
	var reqs []Requirement
	seen := make(map[string]bool)

	for _, path := range includes {
		lower := strings.ToLower(path)

		// Python files -> need python runtime
		if strings.HasSuffix(lower, ".py") {
			if !seen["runtime:python3"] {
				seen["runtime:python3"] = true
				reqs = append(reqs, Requirement{
					Type:   TypeRuntime,
					Value:  "python3",
					Source: "include:" + path,
				})
			}
		}

		// JavaScript/TypeScript files -> need node runtime
		if strings.HasSuffix(lower, ".js") || strings.HasSuffix(lower, ".ts") ||
			strings.HasSuffix(lower, ".mjs") || strings.HasSuffix(lower, ".cjs") {
			if !seen["runtime:node"] {
				seen["runtime:node"] = true
				reqs = append(reqs, Requirement{
					Type:   TypeRuntime,
					Value:  "node",
					Source: "include:" + path,
				})
			}
		}

		// Ruby files -> need ruby runtime
		if strings.HasSuffix(lower, ".rb") {
			if !seen["runtime:ruby"] {
				seen["runtime:ruby"] = true
				reqs = append(reqs, Requirement{
					Type:   TypeRuntime,
					Value:  "ruby",
					Source: "include:" + path,
				})
			}
		}

		// Shell scripts -> need bash/sh
		if strings.HasSuffix(lower, ".sh") || strings.HasSuffix(lower, ".bash") {
			if !seen["runtime:bash"] {
				seen["runtime:bash"] = true
				reqs = append(reqs, Requirement{
					Type:   TypeRuntime,
					Value:  "bash",
					Source: "include:" + path,
				})
			}
		}
	}

	return reqs
}

// Verify checks if a requirement is satisfied
func Verify(req Requirement) VerifyResult {
	result := VerifyResult{Requirement: req}

	switch req.Type {
	case TypeCommand, TypeRuntime:
		_, err := exec.LookPath(req.Value)
		result.Satisfied = err == nil
		if !result.Satisfied {
			result.Message = "Command not found: " + req.Value
		}

	case TypeEnv:
		result.Satisfied = os.Getenv(req.Value) != ""
		if !result.Satisfied {
			result.Message = "Environment variable not set: " + req.Value
		}

	case TypeNPM:
		// Check if node module is importable
		cmd := exec.Command("node", "-e", "require('"+req.Value+"')")
		result.Satisfied = cmd.Run() == nil
		if !result.Satisfied {
			// Use the original package manager from the instructions
			pm := req.PackageManager
			if pm == "" {
				pm = PMnpm // default
			}
			var installCmd string
			switch pm {
			case PMbun:
				installCmd = "bun add " + req.Value
			case PMyarn:
				installCmd = "yarn add " + req.Value
			case PMpnpm:
				installCmd = "pnpm add " + req.Value
			default:
				installCmd = "npm install " + req.Value
			}
			result.Message = "Node package not installed: " + req.Value + "\n  Run: " + installCmd
		}

	case TypePip:
		// Check if python module is importable
		cmd := exec.Command("python3", "-c", "import "+req.Value)
		result.Satisfied = cmd.Run() == nil
		if !result.Satisfied {
			pm := req.PackageManager
			if pm == "" {
				pm = PMpip
			}
			result.Message = "Python package not installed: " + req.Value + "\n  Run: " + string(pm) + " install " + req.Value
		}

	case TypeBrew:
		// Just check if the command exists (brew installs add to PATH)
		_, err := exec.LookPath(req.Value)
		result.Satisfied = err == nil
		if !result.Satisfied {
			result.Message = "Command not found: " + req.Value + "\n  Run: brew install " + req.Value
		}

	case TypeCargo:
		// Just check if the command exists
		_, err := exec.LookPath(req.Value)
		result.Satisfied = err == nil
		if !result.Satisfied {
			result.Message = "Command not found: " + req.Value + "\n  Run: cargo install " + req.Value
		}

	default:
		result.Satisfied = true // Unknown types pass by default
	}

	return result
}

// VerifyAll checks all requirements and returns results
func VerifyAll(reqs []Requirement) []VerifyResult {
	results := make([]VerifyResult, len(reqs))
	for i, req := range reqs {
		results[i] = Verify(req)
	}
	return results
}

// HasUnsatisfied returns true if any requirement is not satisfied
func HasUnsatisfied(results []VerifyResult) bool {
	for _, r := range results {
		if !r.Satisfied {
			return true
		}
	}
	return false
}

// Merge combines content and include requirements, deduplicating
func Merge(contentReqs, includeReqs []Requirement) []Requirement {
	seen := make(map[string]bool)
	var result []Requirement

	for _, req := range contentReqs {
		key := string(req.Type) + ":" + req.Value
		if !seen[key] {
			seen[key] = true
			result = append(result, req)
		}
	}

	for _, req := range includeReqs {
		key := string(req.Type) + ":" + req.Value
		if !seen[key] {
			seen[key] = true
			result = append(result, req)
		}
	}

	return result
}
