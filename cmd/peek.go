package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kennyg/tome/internal/artifact"
	"github.com/kennyg/tome/internal/fetch"
	"github.com/kennyg/tome/internal/source"
	"github.com/kennyg/tome/internal/ui"
)

var peekCmd = &cobra.Command{
	Use:     "peek <source>",
	Aliases: []string{"preview", "inspect"},
	Short:   "Preview an artifact before installing",
	Long: `Peek at an artifact's details without installing it.

Shows name, description, type, and other metadata from the source.

Sources can be:
  owner/repo              GitHub repository
  owner/repo:path         Specific path in a repo
  https://...             Direct URL to a file

Examples:
  tome peek kennyg/agent-skill-gist
  tome peek steveyegge/beads:examples/claude-code-skill`,
	Args: cobra.ExactArgs(1),
	Run:  runPeek,
}

func init() {
	rootCmd.AddCommand(peekCmd)
}

func runPeek(cmd *cobra.Command, args []string) {
	src, err := source.Parse(args[0])
	if err != nil {
		exitWithError(err.Error())
	}

	fmt.Println()
	fmt.Println(ui.SectionHeader("Peeking", 56))
	fmt.Println()
	fmt.Println(ui.InfoLine("Source: " + src.String()))
	fmt.Println()

	client := fetch.NewClient()

	switch src.Type {
	case source.TypeGitHub:
		peekGitHub(client, src)
	case source.TypeURL:
		peekURL(client, src)
	case source.TypeLocal:
		peekLocal(src)
	}
}

func peekGitHub(client *fetch.Client, src *source.Source) {
	// Check if path points to a specific file
	if src.Path != "" && strings.HasSuffix(strings.ToLower(src.Path), ".md") {
		url := src.GitHubRawURL("")
		peekSingleFile(client, url, filepath.Base(src.Path), src.String())
		return
	}

	apiURL := src.GitHubAPIURL()

	// Try to fetch manifest for collection info
	manifest, _ := client.FetchManifest(apiURL)
	if manifest != nil && manifest.Name != "" {
		fmt.Println(ui.Info.Render("  Collection: " + manifest.Name))
		if manifest.Description != "" {
			fmt.Println(ui.Muted.Render("    " + manifest.Description))
		}
		if manifest.Author != "" {
			fmt.Println(ui.Muted.Render(fmt.Sprintf("    by %s", manifest.Author)))
		}
		fmt.Println()
	}

	fmt.Println(ui.Muted.Render("  Scanning for artifacts..."))

	artifacts, err := client.FindArtifacts(apiURL)
	if err != nil {
		// Maybe it's a directory with SKILL.md
		skillURL := src.GitHubRawURL(artifact.SkillFilename)
		content, fetchErr := client.FetchURL(skillURL)
		if fetchErr != nil {
			exitWithError(fmt.Sprintf("failed to scan source: %v", err))
		}
		art, parseErr := fetch.ParseSkill(content, skillURL)
		if parseErr != nil {
			exitWithError(parseErr.Error())
		}
		displayArtifact(art, src.String())
		return
	}

	if len(artifacts) == 0 {
		skillURL := src.GitHubRawURL(artifact.SkillFilename)
		content, err := client.FetchURL(skillURL)
		if err != nil {
			exitWithError("no artifacts found at source")
		}
		art, parseErr := fetch.ParseSkill(content, skillURL)
		if parseErr != nil {
			exitWithError(parseErr.Error())
		}
		displayArtifact(art, src.String())
		return
	}

	fmt.Println(ui.Success.Render(fmt.Sprintf("  Found %d artifact(s)", len(artifacts))))
	fmt.Println()

	// Fetch and display each artifact
	for _, item := range artifacts {
		url := item.DownloadURL
		if url == "" {
			url = src.GitHubRawURL(item.Path)
		}

		content, err := client.FetchURL(url)
		if err != nil {
			fmt.Println(ui.Warning.Render(fmt.Sprintf("  Skipping %s: %v", item.Name, err)))
			continue
		}

		art, err := parseArtifact(content, item.Name, url)
		if err != nil {
			fmt.Println(ui.Warning.Render(fmt.Sprintf("  Skipping %s: %v", item.Name, err)))
			continue
		}

		badge := getBadge(art.Type)

		// Count included files for skills
		includesInfo := ""
		if art.Type == artifact.TypeSkill && item.SkillDir != "" {
			var baseAPIURL string
			if src.Host == "github.com" || src.Host == "" {
				baseAPIURL = fmt.Sprintf("https://api.github.com/repos/%s/%s/contents", src.Owner, src.Repo)
			} else {
				baseAPIURL = fmt.Sprintf("https://%s/api/v3/repos/%s/%s/contents", src.Host, src.Owner, src.Repo)
			}
			if src.Ref != "" {
				baseAPIURL += "?ref=" + src.Ref
			}
			includes, _ := client.DiscoverSkillFiles(baseAPIURL, item.SkillDir)
			if len(includes) > 0 {
				includesInfo = fmt.Sprintf(" (+%d files)", len(includes))
			}
		}

		fmt.Printf("  %s %s%s\n", badge, ui.Highlight.Render(art.Name), ui.Muted.Render(includesInfo))
		if art.Description != "" {
			// Truncate long descriptions
			desc := art.Description
			if len(desc) > 70 {
				desc = desc[:67] + "..."
			}
			fmt.Printf("     %s\n", ui.Muted.Render(desc))
		}
		fmt.Println()
	}

	fmt.Println(ui.Dim.Render(fmt.Sprintf("  Run `tome learn %s` to install", src.String())))
	fmt.Println(ui.PageFooter())
}

func peekURL(client *fetch.Client, src *source.Source) {
	fmt.Println(ui.Muted.Render("  Fetching..."))

	content, err := client.FetchURL(src.URL)
	if err != nil {
		exitWithError(err.Error())
	}

	filename := filepath.Base(src.URL)
	art, err := parseArtifact(content, filename, src.URL)
	if err != nil {
		exitWithError(err.Error())
	}

	displayArtifact(art, src.String())
}

func peekLocal(src *source.Source) {
	exitWithError("local peek not yet implemented - use `tome learn` for local files")
}

func peekSingleFile(client *fetch.Client, url, filename, sourceStr string) {
	fmt.Println(ui.Muted.Render("  Fetching " + filename))

	content, err := client.FetchURL(url)
	if err != nil {
		exitWithError(err.Error())
	}

	art, err := parseArtifact(content, filename, url)
	if err != nil {
		exitWithError(err.Error())
	}

	displayArtifact(art, sourceStr)
}

func displayArtifact(art *artifact.Artifact, sourceStr string) {
	badge := getBadge(art.Type)

	fmt.Println()
	fmt.Println(ui.Title.Render(art.Name))
	fmt.Println()
	fmt.Printf("%s %s\n", badge, ui.Muted.Render(string(art.Type)))
	fmt.Println()

	if art.Description != "" {
		fmt.Println(art.Description)
		fmt.Println()
	}

	fmt.Println(ui.Subtitle.Render("Details"))
	fmt.Println(ui.Divider(40))

	if art.Author != "" {
		fmt.Printf("  Author:  %s\n", art.Author)
	}
	if art.Version != "" {
		fmt.Printf("  Version: %s\n", art.Version)
	}
	fmt.Printf("  Source:  %s\n", sourceStr)

	fmt.Println()
	fmt.Println(ui.Dim.Render(fmt.Sprintf("  Run `tome learn %s` to install", sourceStr)))
	fmt.Println(ui.PageFooter())
}
