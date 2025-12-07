package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/kennyg/tome/internal/artifact"
	"github.com/kennyg/tome/internal/fetch"
	"github.com/kennyg/tome/internal/ui"
)

var buildCmd = &cobra.Command{
	Use:     "bind",
	Aliases: []string{"build", "compile", "validate"},
	Short:   "Bind and validate the tome collection",
	Long: `Scan and validate all artifacts in the current tome collection.

Checks:
  - tome.yaml manifest exists and is valid
  - All commands have valid frontmatter
  - All skills have SKILL.md with valid frontmatter
  - No duplicate artifact names

Examples:
  tome bind
  tome build --json`,
	Run: runBuild,
}

var (
	buildJSON  bool
	buildWrite bool
)

func init() {
	buildCmd.Flags().BoolVar(&buildJSON, "json", false, "Output as JSON")
	buildCmd.Flags().BoolVarP(&buildWrite, "write", "w", false, "Update tome.yaml with discovered artifacts")
}

func runBuild(cmd *cobra.Command, args []string) {
	if !buildJSON {
		fmt.Println()
		fmt.Println(ui.SectionHeader("Binding Tome", 56))
		fmt.Println()
	}

	// Load manifest
	manifestData, err := os.ReadFile("tome.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			exitWithError("tome.yaml not found - run 'tome conjure' first")
		}
		exitWithError(fmt.Sprintf("failed to read tome.yaml: %v", err))
	}

	var manifest artifact.Manifest
	if err := yaml.Unmarshal(manifestData, &manifest); err != nil {
		exitWithError(fmt.Sprintf("failed to parse tome.yaml: %v", err))
	}

	if !buildJSON {
		fmt.Println(ui.Info.Render("  Collection: " + manifest.Name))
		if manifest.Description != "" {
			fmt.Println(ui.Muted.Render("    " + manifest.Description))
		}
		fmt.Println()
	}

	// Scan for artifacts
	var artifacts []artifact.Artifact
	var warnings []string
	var errors []string

	// Scan commands/
	commandsDir := manifest.CommandsDir
	if commandsDir == "" {
		commandsDir = "commands"
	}

	if info, err := os.Stat(commandsDir); err == nil && info.IsDir() {
		entries, _ := os.ReadDir(commandsDir)
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
				continue
			}

			filePath := filepath.Join(commandsDir, entry.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("cannot read %s: %v", filePath, err))
				continue
			}

			art, err := fetch.ParseCommand(content, entry.Name(), filePath)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("%s: %v", entry.Name(), err))
				continue
			}

			if art.Description == "" {
				warnings = append(warnings, fmt.Sprintf("%s: missing description", entry.Name()))
			}

			artifacts = append(artifacts, *art)
		}
	}

	// Scan skills/
	skillsDir := manifest.SkillsDir
	if skillsDir == "" {
		skillsDir = "skills"
	}

	if info, err := os.Stat(skillsDir); err == nil && info.IsDir() {
		entries, _ := os.ReadDir(skillsDir)
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			skillPath := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
			content, err := os.ReadFile(skillPath)
			if err != nil {
				// Check for skill.md (lowercase)
				skillPath = filepath.Join(skillsDir, entry.Name(), "skill.md")
				content, err = os.ReadFile(skillPath)
				if err != nil {
					warnings = append(warnings, fmt.Sprintf("skills/%s: no SKILL.md found", entry.Name()))
					continue
				}
			}

			art, err := fetch.ParseSkill(content, skillPath)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("skills/%s: %v", entry.Name(), err))
				continue
			}

			if art.Description == "" {
				warnings = append(warnings, fmt.Sprintf("skills/%s: missing description", entry.Name()))
			}

			artifacts = append(artifacts, *art)
		}
	}

	// Check for duplicates
	seen := make(map[string]bool)
	for _, art := range artifacts {
		if seen[art.Name] {
			errors = append(errors, fmt.Sprintf("duplicate artifact name: %s", art.Name))
		}
		seen[art.Name] = true
	}

	// Convert to summaries for manifest
	var summaries []artifact.ArtifactSummary
	for _, art := range artifacts {
		// Calculate SHA256 hash of content
		hash := ""
		if art.Content != "" {
			h := sha256.Sum256([]byte(art.Content))
			hash = "sha256:" + hex.EncodeToString(h[:])
		}
		summaries = append(summaries, artifact.ArtifactSummary{
			Name:        art.Name,
			Type:        art.Type,
			Description: art.Description,
			Hash:        hash,
		})
	}
	manifest.Artifacts = summaries

	// Write manifest if requested
	if buildWrite && len(errors) == 0 {
		output, err := yaml.Marshal(&manifest)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to marshal manifest: %v", err))
		} else {
			header := "# Tome Collection Manifest\n# https://github.com/kennyg/tome\n\n"
			if err := os.WriteFile("tome.yaml", []byte(header+string(output)), 0644); err != nil {
				errors = append(errors, fmt.Sprintf("failed to write tome.yaml: %v", err))
			} else if !buildJSON {
				fmt.Println(ui.Muted.Render("  Updated tome.yaml with artifact index"))
				fmt.Println()
			}
		}
	}

	// Output
	if buildJSON {
		// JSON output
		fmt.Printf(`{"name":"%s","artifacts":%d,"warnings":%d,"errors":%d}`,
			manifest.Name, len(artifacts), len(warnings), len(errors))
		fmt.Println()
		return
	}

	// Display results
	if len(artifacts) == 0 {
		fmt.Println(ui.WarningLine("No artifacts found"))
	} else {
		// Count by type
		var skills, commands int
		for _, art := range artifacts {
			switch art.Type {
			case artifact.TypeSkill:
				skills++
			case artifact.TypeCommand:
				commands++
			}
		}

		fmt.Println(ui.Muted.Render("  Artifacts found:"))
		if commands > 0 {
			fmt.Println(ui.Muted.Render(fmt.Sprintf("    %s %d command(s)", ui.CmdBadge, commands)))
		}
		if skills > 0 {
			fmt.Println(ui.Muted.Render(fmt.Sprintf("    %s %d skill(s)", ui.SkillBadge, skills)))
		}
		fmt.Println()

		// List artifacts
		for _, art := range artifacts {
			badge := getBadge(art.Type)
			status := ui.Success.Render("✓")
			if art.Description == "" {
				status = ui.Warning.Render("!")
			}
			fmt.Printf("  %s %s %s\n", status, badge, art.Name)
		}
	}

	// Warnings
	if len(warnings) > 0 {
		fmt.Println()
		fmt.Println(ui.WarningLine(fmt.Sprintf("%d warning(s)", len(warnings))))
		for _, w := range warnings {
			fmt.Println(ui.Muted.Render("    " + w))
		}
	}

	// Errors
	if len(errors) > 0 {
		fmt.Println()
		fmt.Println(ui.ErrorLine(fmt.Sprintf("%d error(s)", len(errors))))
		for _, e := range errors {
			fmt.Println(ui.Muted.Render("    " + e))
		}
	}

	// Summary
	fmt.Println()
	if len(errors) > 0 {
		fmt.Println(ui.ErrorLine("Binding failed"))
	} else if len(warnings) > 0 {
		fmt.Println(ui.SuccessLine(fmt.Sprintf("Bound with %d warning(s)", len(warnings))))
	} else {
		fmt.Println(ui.SuccessLine("Tome bound successfully"))
	}

	fmt.Println(ui.PageFooter())

	if len(errors) > 0 {
		os.Exit(1)
	}
}
