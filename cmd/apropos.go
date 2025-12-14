package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/kennyg/tome/internal/apropos"
	"github.com/kennyg/tome/internal/config"
	"github.com/kennyg/tome/internal/ui"
)

var aproposCmd = &cobra.Command{
	Use:     "apropos <query>",
	Aliases: []string{"whatis", "findskill"},
	Short:   "Find skills by keyword",
	Long: `Search installed skills by keyword or description.

Like Unix apropos, this helps you discover which skill to use for a task.
Searches skill names, descriptions, and extracted keywords.

Examples:
  tome apropos pdf          # Find skills related to PDF
  tome apropos "create chart"  # Find skills for creating charts
  tome apropos spreadsheet  # Find spreadsheet-related skills`,
	Args: cobra.MinimumNArgs(1),
	Run:  runApropos,
}

var aproposRebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Rebuild the skill index",
	Long:  `Force rebuild the apropos index, scanning all skill directories.`,
	Run:   runAproposRebuild,
}

var aproposListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all indexed skills",
	Long:  `Show all skills in the apropos index.`,
	Run:   runAproposList,
}

func init() {
	aproposCmd.AddCommand(aproposRebuildCmd)
	aproposCmd.AddCommand(aproposListCmd)
}

func runApropos(cmd *cobra.Command, args []string) {
	query := strings.Join(args, " ")

	fmt.Println()
	fmt.Println(ui.SectionHeader("Apropos: "+query, 56))
	fmt.Println()

	paths, err := config.GetPaths()
	if err != nil {
		exitWithError("Failed to get paths: " + err.Error())
	}

	// Collect skills directories (user + project)
	skillsDirs := []string{paths.SkillsDir}
	if paths.HasProjectConfig() {
		projectSkillsDir := filepath.Join(filepath.Dir(paths.ProjectConfigDir), ".claude", "skills")
		if info, err := os.Stat(projectSkillsDir); err == nil && info.IsDir() {
			skillsDirs = append(skillsDirs, projectSkillsDir)
		}
	}

	// Load or build index
	index, err := getOrBuildIndex(skillsDirs, paths.SkillsDir, false)
	if err != nil {
		exitWithError("Failed to load index: " + err.Error())
	}

	if index == nil || len(index.Skills) == 0 {
		fmt.Println(ui.WarningLine("No skills indexed"))
		fmt.Println()
		hint := lipgloss.NewStyle().Foreground(ui.Cyan).Render("tome learn <source>")
		fmt.Printf("  Install skills with %s first.\n", hint)
		fmt.Println(ui.PageFooter())
		return
	}

	results := apropos.Search(index, query)

	if len(results) == 0 {
		fmt.Print(ui.NoResults(query))
		fmt.Println(ui.PageFooter())
		return
	}

	fmt.Println(ui.SuccessLine(fmt.Sprintf("Found %d matching skills", len(results))))
	fmt.Println()

	for _, result := range results {
		printSkillResult(result.Skill)
	}

	fmt.Println(ui.PageFooter())
}

func runAproposRebuild(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(ui.SectionHeader("Rebuilding Index", 56))
	fmt.Println()

	paths, err := config.GetPaths()
	if err != nil {
		exitWithError("Failed to get paths: " + err.Error())
	}

	skillsDirs := []string{paths.SkillsDir}
	if paths.HasProjectConfig() {
		projectSkillsDir := filepath.Join(filepath.Dir(paths.ProjectConfigDir), ".claude", "skills")
		if info, err := os.Stat(projectSkillsDir); err == nil && info.IsDir() {
			skillsDirs = append(skillsDirs, projectSkillsDir)
		}
	}

	index, err := getOrBuildIndex(skillsDirs, paths.SkillsDir, true)
	if err != nil {
		exitWithError("Failed to rebuild index: " + err.Error())
	}

	fmt.Println(ui.SuccessLine(fmt.Sprintf("Indexed %d skills", len(index.Skills))))
	fmt.Println(ui.PageFooter())
}

func runAproposList(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(ui.SectionHeader("Indexed Skills", 56))
	fmt.Println()

	paths, err := config.GetPaths()
	if err != nil {
		exitWithError("Failed to get paths: " + err.Error())
	}

	skillsDirs := []string{paths.SkillsDir}
	if paths.HasProjectConfig() {
		projectSkillsDir := filepath.Join(filepath.Dir(paths.ProjectConfigDir), ".claude", "skills")
		if info, err := os.Stat(projectSkillsDir); err == nil && info.IsDir() {
			skillsDirs = append(skillsDirs, projectSkillsDir)
		}
	}

	index, err := getOrBuildIndex(skillsDirs, paths.SkillsDir, false)
	if err != nil {
		exitWithError("Failed to load index: " + err.Error())
	}

	if index == nil || len(index.Skills) == 0 {
		fmt.Println(ui.WarningLine("No skills indexed"))
		fmt.Println(ui.PageFooter())
		return
	}

	skills := apropos.List(index)
	fmt.Println(ui.InfoLine(fmt.Sprintf("%d skills indexed", len(skills))))
	fmt.Println()

	for _, skill := range skills {
		printSkillResult(skill)
	}

	fmt.Println(ui.PageFooter())
}

func getOrBuildIndex(skillsDirs []string, primaryDir string, forceRebuild bool) (*apropos.Index, error) {
	if !forceRebuild {
		index, err := apropos.LoadIndex(primaryDir)
		if err != nil {
			return nil, err
		}

		if index != nil {
			stale, err := apropos.IsStale(primaryDir, index)
			if err == nil && !stale {
				return index, nil
			}
			fmt.Println(ui.InfoLine("Index stale, rebuilding..."))
		} else {
			fmt.Println(ui.InfoLine("Building index..."))
		}
	} else {
		fmt.Println(ui.InfoLine("Force rebuilding index..."))
	}

	index, err := apropos.BuildIndex(skillsDirs)
	if err != nil {
		return nil, err
	}

	if err := apropos.SaveIndex(primaryDir, index); err != nil {
		// Non-fatal, just warn
		fmt.Println(ui.WarningLine("Could not save index: " + err.Error()))
	}

	return index, nil
}

func printSkillResult(skill apropos.Skill) {
	name := lipgloss.NewStyle().Foreground(ui.White).Bold(true).Render(skill.Name)
	fmt.Printf("  %s  %s\n", ui.SkillBadge, name)

	// Truncate description for display
	desc := skill.Description
	maxLen := ui.DescriptionWidth()
	if len(desc) > maxLen {
		desc = desc[:maxLen-3] + "..."
	}
	descStyled := lipgloss.NewStyle().Foreground(ui.Gray).Render(desc)
	fmt.Printf("       %s\n", descStyled)

	// Show invoke command
	cmd := lipgloss.NewStyle().Foreground(ui.Cyan).Render("Skill: " + skill.Name)
	fmt.Printf("       %s\n", cmd)
	fmt.Println()
}
