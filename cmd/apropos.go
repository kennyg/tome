package cmd

import (
	"encoding/json"
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

var (
	aproposJSON bool
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
  tome apropos spreadsheet  # Find spreadsheet-related skills
  tome apropos --json pdf   # Output as JSON (for AI agents)`,
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
	aproposCmd.Flags().BoolVar(&aproposJSON, "json", false, "Output as JSON (for AI agents)")
	aproposCmd.AddCommand(aproposRebuildCmd)
	aproposCmd.AddCommand(aproposListCmd)
}

// JSONResult is the structured output for AI agents
type JSONResult struct {
	Query   string       `json:"query"`
	Count   int          `json:"count"`
	Results []JSONSkill  `json:"results"`
}

// JSONSkill is a skill in JSON output
type JSONSkill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Score       int    `json:"score"`
	Invoke      string `json:"invoke"`
}

func runApropos(cmd *cobra.Command, args []string) {
	query := strings.Join(args, " ")

	paths, err := config.GetPaths()
	if err != nil {
		if aproposJSON {
			outputJSONError(err.Error())
			return
		}
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

	// Load or build index (quiet mode for JSON)
	index, err := getOrBuildIndexQuiet(skillsDirs, paths.SkillsDir, false, aproposJSON)
	if err != nil {
		if aproposJSON {
			outputJSONError(err.Error())
			return
		}
		exitWithError("Failed to load index: " + err.Error())
	}

	results := apropos.Search(index, query)

	// JSON output
	if aproposJSON {
		outputJSON(query, results)
		return
	}

	// Human-readable output
	fmt.Println()
	fmt.Println(ui.SectionHeader("Apropos: "+query, 56))
	fmt.Println()

	if index == nil || len(index.Skills) == 0 {
		fmt.Println(ui.WarningLine("No skills indexed"))
		fmt.Println()
		hint := lipgloss.NewStyle().Foreground(ui.Cyan).Render("tome learn <source>")
		fmt.Printf("  Install skills with %s first.\n", hint)
		fmt.Println(ui.PageFooter())
		return
	}

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

func outputJSON(query string, results []apropos.SearchResult) {
	out := JSONResult{
		Query:   query,
		Count:   len(results),
		Results: make([]JSONSkill, len(results)),
	}

	for i, r := range results {
		out.Results[i] = JSONSkill{
			Name:        r.Skill.Name,
			Description: r.Skill.Description,
			Score:       r.Score,
			Invoke:      fmt.Sprintf("Skill: %s", r.Skill.Name),
		}
	}

	data, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(data))
}

func outputJSONError(msg string) {
	out := map[string]string{"error": msg}
	data, _ := json.Marshal(out)
	fmt.Println(string(data))
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
	return getOrBuildIndexQuiet(skillsDirs, primaryDir, forceRebuild, false)
}

func getOrBuildIndexQuiet(skillsDirs []string, primaryDir string, forceRebuild bool, quiet bool) (*apropos.Index, error) {
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
			if !quiet {
				fmt.Println(ui.InfoLine("Index stale, rebuilding..."))
			}
		} else {
			if !quiet {
				fmt.Println(ui.InfoLine("Building index..."))
			}
		}
	} else {
		if !quiet {
			fmt.Println(ui.InfoLine("Force rebuilding index..."))
		}
	}

	index, err := apropos.BuildIndex(skillsDirs)
	if err != nil {
		return nil, err
	}

	if err := apropos.SaveIndex(primaryDir, index); err != nil {
		// Non-fatal, just warn
		if !quiet {
			fmt.Println(ui.WarningLine("Could not save index: " + err.Error()))
		}
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
