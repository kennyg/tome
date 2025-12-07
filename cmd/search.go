package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kennyg/tome/internal/ui"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for artifacts",
	Long: `Search for skills, commands, and prompts.

Searches GitHub for repositories containing Claude Code artifacts.

Examples:
  tome search memory
  tome search "code review"
  tome search deploy`,
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
	fmt.Println(ui.Title.Render("  Searching for: " + query))
	fmt.Println()

	// Check if gh CLI is available
	ghPath, err := exec.LookPath("gh")
	if err != nil {
		exitWithError("GitHub CLI (gh) not found. Install it from https://cli.github.com")
	}

	// Search for repositories with SKILL.md files
	// We search for repos containing SKILL.md and matching the query
	searchQuery := fmt.Sprintf("%s filename:SKILL.md", query)

	ghCmd := exec.Command(ghPath, "search", "code", searchQuery, "--json", "repository", "--limit", fmt.Sprintf("%d", searchLimit*2))
	output, err := ghCmd.Output()
	if err != nil {
		// Try searching repos directly as fallback
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

	fmt.Println(ui.Success.Render(fmt.Sprintf("  Found %d repositories with skills:", len(repos))))
	fmt.Println()

	for _, repo := range repos {
		fmt.Printf("  %s %s\n", ui.SkillBadge, ui.Highlight.Render(repo))
		fmt.Println(ui.Muted.Render(fmt.Sprintf("      tome learn %s", repo)))
		fmt.Println()
	}
}

func searchRepos(ghPath, query string) {
	// Fallback: search for repos mentioning claude-code or skills
	searchQuery := fmt.Sprintf("%s claude-code OR SKILL.md in:readme,name,description", query)

	ghCmd := exec.Command(ghPath, "search", "repos", searchQuery, "--json", "fullName,description,stargazersCount", "--limit", fmt.Sprintf("%d", searchLimit))
	output, err := ghCmd.Output()
	if err != nil {
		fmt.Println(ui.Warning.Render("  No results found."))
		fmt.Println()
		fmt.Println(ui.Muted.Render("  Try a different search term or browse GitHub directly."))
		return
	}

	var repos []GitHubSearchResult
	if err := json.Unmarshal(output, &repos); err != nil {
		fmt.Println(ui.Warning.Render("  Failed to parse search results."))
		return
	}

	if len(repos) == 0 {
		fmt.Println(ui.Muted.Render("  No repositories found matching your query."))
		fmt.Println()
		fmt.Println(ui.Info.Render("  Tip: Try broader search terms or check GitHub directly."))
		return
	}

	fmt.Println(ui.Success.Render(fmt.Sprintf("  Found %d repositories:", len(repos))))
	fmt.Println()

	for _, repo := range repos {
		stars := ""
		if repo.Stars > 0 {
			stars = ui.Muted.Render(fmt.Sprintf(" ★ %d", repo.Stars))
		}
		fmt.Printf("  %s%s\n", ui.Highlight.Render(repo.FullName), stars)
		if repo.Description != "" {
			desc := repo.Description
			if len(desc) > 60 {
				desc = desc[:60] + "..."
			}
			fmt.Println(ui.Muted.Render("      " + desc))
		}
		fmt.Println(ui.Info.Render(fmt.Sprintf("      tome learn %s", repo.FullName)))
		fmt.Println()
	}
}
