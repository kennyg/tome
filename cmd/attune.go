package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kennyg/tome/internal/ui"
)

var attuneCmd = &cobra.Command{
	Use:   "attune",
	Short: "Attune your project to the Tome",
	Long: `Attune your project to work with Tome and AI agents.

This command prepares your project by:
  - Creating .config/tome/ directory (following dot-config spec)
  - Generating AGENTS.md with Tome usage instructions
  - Setting up any necessary configuration

Run this once in a new project to enable AI agents to use Tome effectively.`,
	Run: runAttune,
}

var (
	attuneForce bool
)

func init() {
	attuneCmd.Flags().BoolVarP(&attuneForce, "force", "f", false, "Overwrite existing AGENTS.md")
}

func runAttune(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(ui.Title.Render("  Attuning your project to the Tome..."))
	fmt.Println()

	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		exitWithError(err.Error())
	}

	// Check if we're in a git repo (recommended but not required)
	gitDir := filepath.Join(cwd, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		fmt.Println(ui.Warning.Render("  Warning: Not a git repository. Tome works best in a git repo."))
		fmt.Println()
	}

	// Create .config/tome/ directory
	configDir := filepath.Join(cwd, ".config", "tome")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		exitWithError(fmt.Sprintf("failed to create config directory: %v", err))
	}
	fmt.Println(ui.Success.Render("  Created .config/tome/"))

	// Create .gitkeep in config dir
	gitkeepPath := filepath.Join(configDir, ".gitkeep")
	if _, err := os.Stat(gitkeepPath); os.IsNotExist(err) {
		content := "# Project-level tome configuration\n# Following dot-config spec: https://dot-config.github.io/\n"
		if err := os.WriteFile(gitkeepPath, []byte(content), 0644); err != nil {
			fmt.Println(ui.Warning.Render("  Warning: Could not create .gitkeep"))
		}
	}

	// Generate or update AGENTS.md
	agentsPath := filepath.Join(cwd, "AGENTS.md")
	agentsExists := false
	if _, err := os.Stat(agentsPath); err == nil {
		agentsExists = true
	}

	if agentsExists && !attuneForce {
		// Check if Tome section already exists
		content, err := os.ReadFile(agentsPath)
		if err == nil && strings.Contains(string(content), "## Using Tome") {
			fmt.Println(ui.Info.Render("  AGENTS.md already contains Tome instructions"))
		} else {
			// Append Tome section
			if err := appendTomeInstructions(agentsPath); err != nil {
				exitWithError(fmt.Sprintf("failed to update AGENTS.md: %v", err))
			}
			fmt.Println(ui.Success.Render("  Updated AGENTS.md with Tome instructions"))
		}
	} else {
		// Create new AGENTS.md
		if err := createAgentsMd(agentsPath); err != nil {
			exitWithError(fmt.Sprintf("failed to create AGENTS.md: %v", err))
		}
		if agentsExists {
			fmt.Println(ui.Success.Render("  Overwrote AGENTS.md with Tome instructions"))
		} else {
			fmt.Println(ui.Success.Render("  Created AGENTS.md with Tome instructions"))
		}
	}

	// Print success message
	fmt.Println()
	fmt.Println(ui.Divider(50))
	fmt.Println()
	fmt.Println(ui.Highlight.Render("  Your project is now attuned."))
	fmt.Println()
	fmt.Println(ui.Muted.Render("  AI agents will understand how to:"))
	fmt.Println(ui.Info.Render("    • Learn new skills with 'tome learn'"))
	fmt.Println(ui.Info.Render("    • List installed artifacts with 'tome list'"))
	fmt.Println(ui.Info.Render("    • Search for capabilities with 'tome search'"))
	fmt.Println()
	fmt.Println(ui.Muted.Render("  May your spellbook grow ever stronger."))
	fmt.Println()
}

func createAgentsMd(path string) error {
	content := getTomeAgentsContent()
	return os.WriteFile(path, []byte(content), 0644)
}

func appendTomeInstructions(path string) error {
	existing, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	content := string(existing)
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += "\n" + getTomeSection()

	return os.WriteFile(path, []byte(content), 0644)
}

func getTomeAgentsContent() string {
	return `# AGENTS.md

Instructions for AI agents working in this repository.

` + getTomeSection()
}

func getTomeSection() string {
	return `## Using Tome

This project uses [Tome](https://github.com/kennyg/tome) to manage AI agent skills, commands, and prompts.

### What is Tome?

Tome is a CLI tool for discovering and installing AI agent capabilities from GitHub repositories. Think of it as a package manager for your AI agent's abilities.

### Quick Reference

` + "```" + `bash
# Search for skills
tome search <query>

# Learn (install) from a source
tome learn owner/repo              # Full repository
tome learn owner/repo:path         # Specific path
tome learn owner/repo@branch       # Specific branch/tag
tome learn ./local/path            # Local directory

# Manage installed artifacts
tome list                          # Show all installed
tome list --skills                 # Filter by type
tome list --commands
tome info <name>                   # Show details
tome remove <name>                 # Uninstall

# Keep everything updated
tome sync                          # Update all artifacts
` + "```" + `

### Artifact Types

| Type | Description | Install Location |
|------|-------------|------------------|
| **Skill** | Passive abilities the agent learns | ` + "`~/.claude/skills/`" + ` |
| **Command** | Slash commands you can invoke | ` + "`~/.claude/commands/`" + ` |
| **Prompt** | Reusable prompt templates | TBD |
| **Hook** | Event-triggered actions | ` + "`settings.json`" + ` |

### Configuration

Tome follows the [dot-config specification](https://dot-config.github.io/):

- **User config**: ` + "`~/.config/tome/`" + ` (or ` + "`$XDG_CONFIG_HOME/tome/`" + `)
- **Project config**: ` + "`.config/tome/`" + `

### For AI Agents

When asked to add capabilities or skills:
1. Use ` + "`tome search <query>`" + ` to find relevant skills
2. Use ` + "`tome learn <source>`" + ` to install them
3. Use ` + "`tome list`" + ` to verify installation

When exploring what's available:
- Check ` + "`tome list`" + ` for currently installed artifacts
- Use ` + "`tome info <name>`" + ` for details about specific artifacts
`
}
