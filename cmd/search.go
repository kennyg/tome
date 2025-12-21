package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/kennyg/tome/internal/ghclient"
	"github.com/kennyg/tome/internal/ui"
)

var searchCmd = &cobra.Command{
	Use:     "seek <query>",
	Aliases: []string{"scry", "divine", "search", "find"},
	Short:   "Seek artifacts in the archives",
	Long: `Seek skills, commands, and prompts in the archives.

Searches GitHub for repositories containing artifacts.

Examples:
  tome seek memory
  tome seek "code review"
  tome scry deploy`,
	Args: cobra.MinimumNArgs(1),
	Run:  runSearch,
}

var searchLimit int

func init() {
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "n", 10, "Maximum results to show")
}

func runSearch(cmd *cobra.Command, args []string) {
	query := strings.Join(args, " ")

	fmt.Println()
	fmt.Println(ui.SectionHeader("Seeking: "+query, 56))
	fmt.Println()
	fmt.Println(ui.InfoLine("Searching the archives..."))
	fmt.Println()

	gh := ghclient.New()
	ctx := context.Background()

	// Search for repositories with SKILL.md files
	searchQuery := fmt.Sprintf("%s filename:SKILL.md", query)

	codeResults, err := gh.SearchCode(ctx, searchQuery, searchLimit*2)
	if err != nil || len(codeResults) == 0 {
		searchRepos(gh, ctx, query)
		return
	}

	// Deduplicate repos
	seen := make(map[string]bool)
	var repos []string
	for _, r := range codeResults {
		if !seen[r.Repository] {
			seen[r.Repository] = true
			repos = append(repos, r.Repository)
			if len(repos) >= searchLimit {
				break
			}
		}
	}

	fmt.Println(ui.SuccessLine(fmt.Sprintf("Found %d grimoires with artifacts", len(repos))))
	fmt.Println()

	for _, repo := range repos {
		name := lipgloss.NewStyle().Foreground(ui.White).Bold(true).Render(repo)
		cmdText := lipgloss.NewStyle().Foreground(ui.Cyan).Render("tome learn " + repo)
		fmt.Printf("  %s  %s\n", ui.SkillBadge, name)
		fmt.Printf("       %s\n", cmdText)
		fmt.Println()
	}

	fmt.Println(ui.PageFooter())
}

func searchRepos(gh *ghclient.Client, ctx context.Context, query string) {
	// Fallback: search for repos mentioning claude-code or skills
	searchQuery := fmt.Sprintf("%s claude-code OR SKILL.md in:readme,name,description", query)

	repos, err := gh.SearchRepos(ctx, searchQuery, searchLimit)
	if err != nil {
		fmt.Print(ui.NoResults(query))
		return
	}

	if len(repos) == 0 {
		fmt.Print(ui.NoResults(query))
		return
	}

	fmt.Println(ui.SuccessLine(fmt.Sprintf("Found %d repositories", len(repos))))
	fmt.Println()

	for _, repo := range repos {
		name := lipgloss.NewStyle().Foreground(ui.White).Bold(true).Render(repo.FullName)
		stars := ""
		if repo.Stars > 0 {
			stars = lipgloss.NewStyle().Foreground(ui.Gold).Render(fmt.Sprintf(" ★ %d", repo.Stars))
		}
		fmt.Printf("  %s%s\n", name, stars)

		if repo.Description != "" {
			desc := ui.Truncate(repo.Description, 55)
			descStyled := lipgloss.NewStyle().Foreground(ui.Gray).Render(desc)
			fmt.Printf("  %s\n", descStyled)
		}

		cmdText := lipgloss.NewStyle().Foreground(ui.Cyan).Render("tome learn " + repo.FullName)
		fmt.Printf("  %s\n", cmdText)
		fmt.Println()
	}

	fmt.Println(ui.PageFooter())
}
