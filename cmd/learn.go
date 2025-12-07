package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/kennyg/tome/internal/artifact"
	"github.com/kennyg/tome/internal/config"
	"github.com/kennyg/tome/internal/fetch"
	"github.com/kennyg/tome/internal/source"
	"github.com/kennyg/tome/internal/ui"
)

var learnCmd = &cobra.Command{
	Use:     "learn <source>",
	Aliases: []string{"inscribe", "add", "install"},
	Short:   "Inscribe a new artifact into the tome",
	Long: `Learn and inscribe artifacts from various sources.

Sources can be:
  owner/repo              GitHub repository (installs all artifacts)
  owner/repo:path         Specific path in a repo
  owner/repo@ref          Specific branch/tag/commit
  https://...             Direct URL to a file
  ./local/path            Local file or directory

Artifact types are auto-detected:
  SKILL.md files       → Skills (passive agent knowledge)
  Other .md files      → Commands (invokable with /name)

Examples:
  tome learn kennyg/yegges-tips                    # All commands from repo
  tome inscribe steveyegge/beads:examples/claude-code-skill
  tome learn https://raw.githubusercontent.com/.../SKILL.md
  tome learn ./my-local-skill`,
	Args: cobra.ExactArgs(1),
	Run:  runLearn,
}

var (
	learnGlobal bool
	learnAgent  string
)

func init() {
	learnCmd.Flags().BoolVarP(&learnGlobal, "global", "g", true, "Install globally (default)")
	learnCmd.Flags().StringVarP(&learnAgent, "agent", "a", "", "Target agent (claude, opencode, crush, cursor, windsurf)")
}

func runLearn(cmd *cobra.Command, args []string) {
	src, err := source.Parse(args[0])
	if err != nil {
		exitWithError(err.Error())
	}

	fmt.Println()
	fmt.Println(ui.SectionHeader("Inscribing", 56))
	fmt.Println()
	fmt.Println(ui.InfoLine("Source: " + src.String()))
	fmt.Println()

	// Determine which agent to use
	agent := config.DefaultAgent()
	if learnAgent != "" {
		agent = config.Agent(learnAgent)
		if config.GetAgentConfig(agent) == nil {
			exitWithError(fmt.Sprintf("unknown agent: %s (try: claude, opencode, crush, cursor, windsurf)", learnAgent))
		}
	}

	paths, err := config.GetPathsForAgent(agent)
	if err != nil {
		exitWithError(err.Error())
	}

	agentCfg := config.GetAgentConfig(agent)
	fmt.Println(ui.Muted.Render(fmt.Sprintf("  Target agent: %s", agentCfg.DisplayName)))
	fmt.Println()

	// Ensure directories exist
	if err := paths.EnsureDirs(); err != nil {
		exitWithError(fmt.Sprintf("failed to create directories: %v", err))
	}

	client := fetch.NewClient()

	switch src.Type {
	case source.TypeGitHub:
		learnFromGitHub(client, src, paths)
	case source.TypeURL:
		learnFromURL(client, src, paths)
	case source.TypeLocal:
		learnFromLocal(src, paths)
	}
}

func learnFromGitHub(client *fetch.Client, src *source.Source, paths *config.Paths) {
	fmt.Println(ui.Info.Render("  Source: GitHub"))
	fmt.Println(ui.Muted.Render(fmt.Sprintf("    %s/%s", src.Owner, src.Repo)))
	if src.Path != "" {
		fmt.Println(ui.Muted.Render(fmt.Sprintf("    Path: %s", src.Path)))
	}
	fmt.Println()

	// Check if path points to a specific file
	if src.Path != "" && strings.HasSuffix(strings.ToLower(src.Path), ".md") {
		// Single file
		url := src.GitHubRawURL("")
		learnSingleFile(client, url, filepath.Base(src.Path), src.String(), paths)
		return
	}

	// Try to list directory contents
	apiURL := src.GitHubAPIURL()
	fmt.Println(ui.Muted.Render("  Scanning for artifacts..."))

	artifacts, err := client.FindArtifacts(apiURL)
	if err != nil {
		// Maybe it's a directory with SKILL.md
		skillURL := src.GitHubRawURL("SKILL.md")
		content, fetchErr := client.FetchURL(skillURL)
		if fetchErr != nil {
			exitWithError(fmt.Sprintf("failed to scan repo: %v", err))
		}
		// Found a SKILL.md
		art, parseErr := fetch.ParseSkill(content, skillURL)
		if parseErr != nil {
			exitWithError(parseErr.Error())
		}
		art.Source = src.String()
		installArtifact(art, paths)
		return
	}

	if len(artifacts) == 0 {
		// Check for SKILL.md specifically
		skillURL := src.GitHubRawURL("SKILL.md")
		content, err := client.FetchURL(skillURL)
		if err != nil {
			exitWithError("no artifacts found in repository")
		}
		art, err := fetch.ParseSkill(content, skillURL)
		if err != nil {
			exitWithError(err.Error())
		}
		art.Source = src.String()
		installArtifact(art, paths)
		return
	}

	fmt.Println(ui.Success.Render(fmt.Sprintf("  Found %d artifact(s)", len(artifacts))))
	fmt.Println()

	// Install each artifact
	var installed []string
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

		art.Source = src.String()
		installArtifactQuiet(art, paths)
		installed = append(installed, art.Name)
	}

	// Summary
	fmt.Println()
	fmt.Println(ui.SuccessLine(fmt.Sprintf("Inscribed %d artifact(s)", len(installed))))
	for _, name := range installed {
		fmt.Println(ui.Muted.Render("    • " + name))
	}
	fmt.Println()
	fmt.Println(ui.Dim.Render("  Your tome grows stronger."))
	fmt.Println(ui.PageFooter())
}

func learnSingleFile(client *fetch.Client, url, filename, source string, paths *config.Paths) {
	fmt.Println(ui.Muted.Render("  Fetching " + filename))

	content, err := client.FetchURL(url)
	if err != nil {
		exitWithError(err.Error())
	}

	art, err := parseArtifact(content, filename, url)
	if err != nil {
		exitWithError(err.Error())
	}

	art.Source = source
	installArtifact(art, paths)
}

func parseArtifact(content []byte, filename, sourceURL string) (*artifact.Artifact, error) {
	artType := fetch.DetectArtifactType(filename)

	switch artType {
	case artifact.TypeSkill:
		return fetch.ParseSkill(content, sourceURL)
	case artifact.TypeCommand:
		return fetch.ParseCommand(content, filename, sourceURL)
	default:
		// Default to command for unknown .md files
		if strings.HasSuffix(strings.ToLower(filename), ".md") {
			return fetch.ParseCommand(content, filename, sourceURL)
		}
		return nil, fmt.Errorf("unknown artifact type for %s", filename)
	}
}

func learnFromURL(client *fetch.Client, src *source.Source, paths *config.Paths) {
	fmt.Println(ui.Info.Render("  Source: URL"))
	fmt.Println(ui.Muted.Render("    " + src.URL))
	fmt.Println()

	filename := filepath.Base(src.URL)
	learnSingleFile(client, src.URL, filename, src.Original, paths)
}

func learnFromLocal(src *source.Source, paths *config.Paths) {
	fmt.Println(ui.Info.Render("  Source: Local"))
	fmt.Println(ui.Muted.Render("    " + src.Path))
	fmt.Println()

	// Check if it's a file or directory
	info, err := os.Stat(src.Path)
	if err != nil {
		exitWithError(fmt.Sprintf("cannot access %s: %v", src.Path, err))
	}

	if !info.IsDir() {
		// Single file
		content, err := os.ReadFile(src.Path)
		if err != nil {
			exitWithError(fmt.Sprintf("cannot read %s: %v", src.Path, err))
		}

		filename := filepath.Base(src.Path)
		art, err := parseArtifact(content, filename, src.Path)
		if err != nil {
			exitWithError(err.Error())
		}

		art.Source = src.Original
		installArtifact(art, paths)
		return
	}

	// Directory - scan for artifacts
	fmt.Println(ui.Muted.Render("  Scanning for artifacts..."))

	entries, err := os.ReadDir(src.Path)
	if err != nil {
		exitWithError(fmt.Sprintf("cannot read directory: %v", err))
	}

	var installed []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !fetch.IsArtifactFile(entry.Name()) {
			continue
		}

		filePath := filepath.Join(src.Path, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Println(ui.Warning.Render(fmt.Sprintf("  Skipping %s: %v", entry.Name(), err)))
			continue
		}

		art, err := parseArtifact(content, entry.Name(), filePath)
		if err != nil {
			fmt.Println(ui.Warning.Render(fmt.Sprintf("  Skipping %s: %v", entry.Name(), err)))
			continue
		}

		art.Source = src.Original
		installArtifactQuiet(art, paths)
		installed = append(installed, art.Name)
	}

	if len(installed) == 0 {
		exitWithError("no artifacts found in directory")
	}

	// Summary
	fmt.Println()
	fmt.Println(ui.SuccessLine(fmt.Sprintf("Inscribed %d artifact(s)", len(installed))))
	for _, name := range installed {
		fmt.Println(ui.Muted.Render("    • " + name))
	}
	fmt.Println()
	fmt.Println(ui.Dim.Render("  Your tome grows stronger."))
	fmt.Println(ui.PageFooter())
}

func installArtifact(art *artifact.Artifact, paths *config.Paths) {
	doInstall(art, paths)

	// Success output
	badge := getBadge(art.Type)
	fmt.Printf("\n  %s %s\n", badge, ui.Highlight.Render(art.Name))
	if art.Description != "" {
		desc := ui.Truncate(art.Description, 55)
		fmt.Println(ui.Muted.Render("  " + desc))
	}
	fmt.Println()
	fmt.Println(ui.SuccessLine("Inscribed successfully"))
	fmt.Println(ui.Dim.Render("  " + getInstallPath(art, paths)))
	fmt.Println()
	fmt.Println(ui.Dim.Render("  Your tome grows stronger."))
	fmt.Println(ui.PageFooter())
}

func installArtifactQuiet(art *artifact.Artifact, paths *config.Paths) {
	doInstall(art, paths)

	badge := getBadge(art.Type)
	fmt.Printf("  %s %s\n", badge, ui.Highlight.Render(art.Name))
}

func doInstall(art *artifact.Artifact, paths *config.Paths) {
	installPath := getInstallPath(art, paths)
	installDir := filepath.Dir(installPath)

	// Create directory if needed
	if err := os.MkdirAll(installDir, 0755); err != nil {
		exitWithError(fmt.Sprintf("failed to create directory: %v", err))
	}

	// Write the file
	if err := os.WriteFile(installPath, []byte(art.Content), 0644); err != nil {
		exitWithError(fmt.Sprintf("failed to write file: %v", err))
	}

	// Update state
	state, err := config.LoadState(paths.StateFile)
	if err != nil {
		exitWithError(fmt.Sprintf("failed to load state: %v", err))
	}

	installed := artifact.InstalledArtifact{
		Artifact:  *art,
		LocalPath: installPath,
	}
	installed.InstalledAt = time.Now()

	state.AddInstalled(installed)

	// Ensure state directory exists
	stateDir := filepath.Dir(paths.StateFile)
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		exitWithError(fmt.Sprintf("failed to create state directory: %v", err))
	}

	if err := config.SaveState(paths.StateFile, state); err != nil {
		exitWithError(fmt.Sprintf("failed to save state: %v", err))
	}
}

func getInstallPath(art *artifact.Artifact, paths *config.Paths) string {
	switch art.Type {
	case artifact.TypeSkill:
		// Skills go in a directory with SKILL.md
		safeDir := fetch.SanitizeFilename(art.Name)
		return filepath.Join(paths.SkillsDir, safeDir, "SKILL.md")
	case artifact.TypeCommand:
		// Commands are just .md files
		safeName := fetch.SanitizeFilename(art.Name) + ".md"
		return filepath.Join(paths.CommandsDir, safeName)
	default:
		return filepath.Join(paths.CommandsDir, art.Filename)
	}
}
