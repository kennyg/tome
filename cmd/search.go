package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

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

// GitHubSearchResult represents a repository from GitHub search
type GitHubSearchResult struct {
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	HTMLURL     string `json:"html_url"`
	Stars       int    `json:"stargazers_count"`
	UpdatedAt   string `json:"updated_at"`
}

func runSearch(cmd *cobra.Command, args []string) {
	query := strings.Join(args, " ")

	fmt.Println()
	fmt.Println(ui.SectionHeader("Seeking: "+query, 56))
	fmt.Println()
	fmt.Println(ui.InfoLine("Searching the archives..."))
	fmt.Println()

	// Check if gh CLI is available
	ghPath, err := exec.LookPath("gh")
	if err != nil {
		exitWithError("GitHub CLI (gh) not found. Install it from https://cli.github.com")
	}

	// Search for repositories with SKILL.md files
	searchQuery := fmt.Sprintf("%s filename:SKILL.md", query)

	ghCmd := exec.Command(ghPath, "search", "code", searchQuery, "--json", "repository", "--limit", fmt.Sprintf("%d", searchLimit*2))
	output, err := ghCmd.Output()
	if err != nil {
		searchRepos(ghPath, query)
		return
	}

	var codeResults []struct {
		Repository struct {
			NameWithOwner string `json:"nameWithOwner"`
			Description   string `json:"description"`
		} `json:"repository"`
	}

	if err := json.Unmarshal(output, &codeResults); err != nil {
		searchRepos(ghPath, query)
		return
	}

	if len(codeResults) == 0 {
		searchRepos(ghPath, query)
		return
	}

	// Deduplicate repos
	seen := make(map[string]bool)
	var repos []string
	for _, r := range codeResults {
		if !seen[r.Repository.NameWithOwner] {
			seen[r.Repository.NameWithOwner] = true
			repos = append(repos, r.Repository.NameWithOwner)
			if len(repos) >= searchLimit {
				break
			}
		}
	}

	fmt.Println(ui.SuccessLine(fmt.Sprintf("Found %d grimoires with artifacts", len(repos))))
	fmt.Println()

	for _, repo := range repos {
		name := lipgloss.NewStyle().Foreground(ui.White).Bold(true).Render(repo)
		cmd := lipgloss.NewStyle().Foreground(ui.Cyan).Render("tome learn " + repo)
		fmt.Printf("  %s  %s\n", ui.SkillBadge, name)
		fmt.Printf("       %s\n", cmd)
		fmt.Println()
	}

	fmt.Println(ui.PageFooter())
}

func searchRepos(ghPath, query string) {
	// Fallback: search for repos mentioning claude-code or skills
	searchQuery := fmt.Sprintf("%s claude-code OR SKILL.md in:readme,name,description", query)

	ghCmd := exec.Command(ghPath, "search", "repos", searchQuery, "--json", "fullName,description,stargazersCount", "--limit", fmt.Sprintf("%d", searchLimit))
	output, err := ghCmd.Output()
	if err != nil {
		fmt.Print(ui.NoResults(query))
		return
	}

	var repos []GitHubSearchResult
	if err := json.Unmarshal(output, &repos); err != nil {
		fmt.Println(ui.ErrorLine("Failed to parse search results"))
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

		cmd := lipgloss.NewStyle().Foreground(ui.Cyan).Render("tome learn " + repo.FullName)
		fmt.Printf("  %s\n", cmd)
		fmt.Println()
	}

	fmt.Println(ui.PageFooter())
}
