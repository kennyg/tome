package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/kennyg/tome/internal/artifact"
	"github.com/kennyg/tome/internal/ui"
)

var initCmd = &cobra.Command{
	Use:     "conjure",
	Aliases: []string{"init", "create", "new"},
	Short:   "Conjure a new tome collection",
	Long: `Initialize a new tome collection in the current directory.

Creates a tome.yaml manifest and the standard directory structure:
  commands/    Slash commands (.md files)
  skills/      Skills (subdirectories with SKILL.md)

Examples:
  tome conjure
  tome conjure --name my-prompts
  tome init --name my-prompts --author "Your Name"`,
	Run: runInit,
}

var (
	initName        string
	initDescription string
	initAuthor      string
	initLicense     string
	initTags        string
)

func init() {
	initCmd.Flags().StringVarP(&initName, "name", "n", "", "Collection name (defaults to directory name)")
	initCmd.Flags().StringVarP(&initDescription, "description", "d", "", "Collection description")
	initCmd.Flags().StringVarP(&initAuthor, "author", "a", "", "Author name")
	initCmd.Flags().StringVarP(&initLicense, "license", "l", "MIT", "License")
	initCmd.Flags().StringVarP(&initTags, "tags", "t", "", "Comma-separated tags")
}

func runInit(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(ui.SectionHeader("Conjuring New Tome", 56))
	fmt.Println()

	// Check if tome.yaml already exists
	if _, err := os.Stat("tome.yaml"); err == nil {
		exitWithError("tome.yaml already exists in this directory")
	}

	// Get directory name for default collection name
	cwd, err := os.Getwd()
	if err != nil {
		exitWithError(fmt.Sprintf("failed to get current directory: %v", err))
	}
	dirName := filepath.Base(cwd)

	// Build manifest
	name := initName
	if name == "" {
		name = dirName
	}

	var tags []string
	if initTags != "" {
		for _, t := range strings.Split(initTags, ",") {
			tags = append(tags, strings.TrimSpace(t))
		}
	}

	manifest := artifact.Manifest{
		Name:        name,
		Description: initDescription,
		Author:      initAuthor,
		License:     initLicense,
		Tags:        tags,
	}

	// Create directories
	dirs := []string{"commands", "skills"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			exitWithError(fmt.Sprintf("failed to create %s directory: %v", dir, err))
		}
		fmt.Println(ui.Muted.Render(fmt.Sprintf("  Created %s/", dir)))
	}

	// Write tome.yaml
	content, err := yaml.Marshal(&manifest)
	if err != nil {
		exitWithError(fmt.Sprintf("failed to generate tome.yaml: %v", err))
	}

	header := `# Tome Collection Manifest
# https://github.com/kennyg/tome
#
# Tome auto-discovers:
#   commands/*.md  -> slash commands
#   skills/*/SKILL.md -> skills

`
	fullContent := header + string(content)

	if err := os.WriteFile("tome.yaml", []byte(fullContent), 0644); err != nil {
		exitWithError(fmt.Sprintf("failed to write tome.yaml: %v", err))
	}
	fmt.Println(ui.Muted.Render("  Created tome.yaml"))

	// Create example command
	exampleCmd := `---
description: Example slash command - customize or delete this file
---

# Example Command

This is an example slash command. When users invoke it with ` + "`/example`" + `, this content will be provided as context.

## Usage

Describe how to use this command.

## Template

` + "```" + `
Your prompt template here
` + "```" + `
`
	if err := os.WriteFile("commands/example.md", []byte(exampleCmd), 0644); err != nil {
		fmt.Println(ui.Warning.Render("  Could not create example command"))
	} else {
		fmt.Println(ui.Muted.Render("  Created commands/example.md"))
	}

	fmt.Println()
	fmt.Println(ui.SuccessLine("Tome conjured successfully"))
	fmt.Println()
	fmt.Println(ui.Muted.Render("  Next steps:"))
	fmt.Println(ui.Muted.Render("    1. Edit tome.yaml with your collection details"))
	fmt.Println(ui.Muted.Render("    2. Add commands to commands/*.md"))
	fmt.Println(ui.Muted.Render("    3. Add skills to skills/<name>/SKILL.md"))
	fmt.Println(ui.Muted.Render("    4. Run 'tome bind' to validate"))
	fmt.Println(ui.PageFooter())
}
