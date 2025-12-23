package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/kennyg/tome/internal/artifact"
	"github.com/kennyg/tome/internal/config"
	"github.com/kennyg/tome/internal/detect"
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
	learnCmd.Flags().BoolVarP(&learnGlobal, "global", "g", false, "Install globally to ~/.<agent>/ instead of project-local")
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

	// Determine paths: local by default if attuned, global with --global flag
	var paths *config.Paths
	var installLocation string

	if learnGlobal {
		// Explicit global install
		paths, err = config.GetPathsForAgent(agent)
		if err != nil {
			exitWithError(err.Error())
		}
		installLocation = "global"
	} else if config.IsAttuned(agent) {
		// Project-local install (default when attuned)
		paths, err = config.GetLocalPaths(agent)
		if err != nil {
			exitWithError(err.Error())
		}
		installLocation = "project"
	} else {
		// Not attuned, fall back to global
		paths, err = config.GetPathsForAgent(agent)
		if err != nil {
			exitWithError(err.Error())
		}
		installLocation = "global"
	}

	agentCfg := config.GetAgentConfig(agent)
	locationInfo := fmt.Sprintf("  Target: %s (%s)", agentCfg.DisplayName, installLocation)
	fmt.Println(ui.Muted.Render(locationInfo))
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
	// Handle single file case
	if src.Path != "" && strings.HasSuffix(strings.ToLower(src.Path), ".md") {
		displayGitHubSource(src)
		url := src.GitHubRawURL("")
		learnSingleFile(client, url, filepath.Base(src.Path), src.String(), paths, nil)
		return
	}

	// Fetch README.md for requirement detection
	readmeReqs := fetchReadmeRequirements(client, src)

	// Try to list directory contents
	apiURL := src.GitHubAPIURL()

	// Check if this is a plugin
	if client.IsPlugin(apiURL) {
		learnPlugin(client, src, apiURL, paths)
		return
	}

	// Display source/collection info
	manifest, _ := client.FetchManifest(apiURL)
	displaySourceInfo(manifest, src)

	// Find artifacts
	fmt.Println(ui.Muted.Render("  Scanning for artifacts..."))
	artifacts, err := client.FindArtifacts(apiURL)

	// Handle fallback cases
	if err != nil || len(artifacts) == 0 {
		if tryFallbackSkill(client, src, paths, readmeReqs, err) {
			return
		}
	}

	// Install found artifacts
	result := installFoundArtifacts(client, src, paths, artifacts, readmeReqs)

	// Display summary
	displayInstallSummary(result, src)
}

// displayGitHubSource shows source info for a GitHub URL
func displayGitHubSource(src *source.Source) {
	fmt.Println(ui.Info.Render("  Source: GitHub"))
	fmt.Println(ui.Muted.Render(fmt.Sprintf("    %s/%s", src.Owner, src.Repo)))
	if src.Path != "" {
		fmt.Println(ui.Muted.Render(fmt.Sprintf("    Path: %s", src.Path)))
	}
	fmt.Println()
}

// fetchReadmeRequirements fetches README.md and extracts requirements
func fetchReadmeRequirements(client *fetch.Client, src *source.Source) []detect.Requirement {
	for _, readmeName := range []string{"README.md", "readme.md", "Readme.md"} {
		readmeURL := src.GitHubRawURL(readmeName)
		if content, err := client.FetchURL(readmeURL); err == nil {
			return detect.FromContent(string(content))
		}
	}
	return nil
}

// displaySourceInfo shows collection or source information
func displaySourceInfo(manifest *artifact.Manifest, src *source.Source) {
	if manifest != nil && manifest.Name != "" {
		fmt.Println(ui.Info.Render("  Collection: " + manifest.Name))
		if manifest.Description != "" {
			fmt.Println(ui.Muted.Render("    " + manifest.Description))
		}
		if manifest.Author != "" {
			fmt.Println(ui.Muted.Render(fmt.Sprintf("    by %s", manifest.Author)))
		}
		if manifest.Source != "" {
			fmt.Println(ui.Muted.Render(fmt.Sprintf("    %s", manifest.Source)))
		}
		if len(manifest.Tags) > 0 {
			fmt.Println(ui.Muted.Render(fmt.Sprintf("    tags: %s", strings.Join(manifest.Tags, ", "))))
		}
	} else {
		displayGitHubSource(src)
		return // displayGitHubSource already prints newline
	}
	fmt.Println()
}

// tryFallbackSkill attempts to find a SKILL.md when artifact scanning fails
// Returns true if a skill was found and installed
func tryFallbackSkill(client *fetch.Client, src *source.Source, paths *config.Paths, readmeReqs []detect.Requirement, scanErr error) bool {
	skillURL := src.GitHubRawURL(artifact.SkillFilename)
	content, err := client.FetchURL(skillURL)
	if err != nil {
		// No SKILL.md found - check if this is an npm package
		pkgURL := src.GitHubRawURL("package.json")
		if pkgContent, pkgErr := client.FetchURL(pkgURL); pkgErr == nil {
			showNpmPackageGuidance(pkgContent, src)
			return true
		}
		if scanErr != nil {
			exitWithError(fmt.Sprintf("failed to scan repo: %v", scanErr))
		}
		exitWithError("no artifacts found in repository")
	}

	art, parseErr := fetch.ParseSkill(content, skillURL)
	if parseErr != nil {
		exitWithError(parseErr.Error())
	}
	art.Source = src.String()
	installArtifactWithExtraReqs(art, paths, readmeReqs)
	return true
}

// installResult holds the results of installing artifacts
type installResult struct {
	installed     []string
	skipped       []skippedArtifact
	allReqs       []detect.Requirement
	skillContents []skillContent
}

type skippedArtifact struct {
	name   string
	reason string
}

type skillContent struct {
	name    string
	content string
}

// installFoundArtifacts installs all found artifacts and returns the results
func installFoundArtifacts(client *fetch.Client, src *source.Source, paths *config.Paths, artifacts []fetch.GitHubContent, readmeReqs []detect.Requirement) installResult {
	fmt.Println(ui.Success.Render(fmt.Sprintf("  Found %d artifact(s)", len(artifacts))))
	fmt.Println()

	var result installResult

	for _, item := range artifacts {
		url := item.DownloadURL
		if url == "" {
			url = src.GitHubRawURL(item.Path)
		}

		content, err := client.FetchURL(url)
		if err != nil {
			fmt.Println(ui.Warning.Render(fmt.Sprintf("  Skipping %s: %v", item.Name, err)))
			result.skipped = append(result.skipped, skippedArtifact{item.Name, fmt.Sprintf("fetch failed: %v", err)})
			continue
		}

		art, err := parseArtifact(content, item.Name, url)
		if err != nil {
			fmt.Println(ui.Warning.Render(fmt.Sprintf("  Skipping %s: %v", item.Name, err)))
			result.skipped = append(result.skipped, skippedArtifact{item.Name, fmt.Sprintf("parse failed: %v", err)})
			continue
		}

		// Discover skill includes if applicable
		includes := discoverSkillIncludes(client, src, item, art)

		art.Source = src.String()
		reqs := installArtifactQuietWithExtras(art, paths, includes, readmeReqs)
		result.installed = append(result.installed, art.Name)
		result.allReqs = detect.Merge(result.allReqs, reqs)

		if art.Type == artifact.TypeSkill {
			result.skillContents = append(result.skillContents, skillContent{art.Name, string(content)})
		}
	}

	return result
}

// discoverSkillIncludes finds additional files to include with a skill
func discoverSkillIncludes(client *fetch.Client, src *source.Source, item fetch.GitHubContent, art *artifact.Artifact) []fetch.IncludedFile {
	if art.Type != artifact.TypeSkill {
		return nil
	}

	skillDir := item.SkillDir
	if skillDir == "" && src.Path != "" {
		skillDir = src.Path
	}
	if skillDir == "" {
		return nil
	}

	// Build base API URL for discovery
	var baseAPIURL string
	if src.Host == "github.com" || src.Host == "" {
		baseAPIURL = fmt.Sprintf("https://api.github.com/repos/%s/%s/contents", src.Owner, src.Repo)
	} else {
		baseAPIURL = fmt.Sprintf("https://%s/api/v3/repos/%s/%s/contents", src.Host, src.Owner, src.Repo)
	}
	if src.Ref != "" {
		baseAPIURL += "?ref=" + src.Ref
	}

	includes, err := client.DiscoverSkillFiles(baseAPIURL, skillDir)
	if err != nil {
		fmt.Println(ui.Warning.Render(fmt.Sprintf("  Warning: couldn't fetch skill files for %s: %v", item.Name, err)))
	}
	return includes
}

// displayInstallSummary shows the final installation summary
func displayInstallSummary(result installResult, src *source.Source) {
	fmt.Println()
	if len(result.installed) > 0 {
		fmt.Println(ui.SuccessLine(fmt.Sprintf("Inscribed %d artifact(s)", len(result.installed))))
		for _, name := range result.installed {
			fmt.Println(ui.Muted.Render("    • " + name))
		}
	}

	if len(result.skipped) > 0 {
		fmt.Println()
		fmt.Println(ui.Warning.Render(fmt.Sprintf("  Skipped %d artifact(s):", len(result.skipped))))
		for _, s := range result.skipped {
			fmt.Println(ui.Muted.Render(fmt.Sprintf("    • %s: %s", s.name, s.reason)))
		}
	}

	if len(result.installed) == 0 {
		exitWithError("no artifacts were installed successfully")
	}

	if len(result.allReqs) > 0 {
		displayDetectedRequirements(src.String(), result.allReqs)
	}

	fmt.Println()
	fmt.Println(ui.Dim.Render("  Your tome grows stronger."))

	// Show usage info for installed skills
	for _, skill := range result.skillContents {
		usage := extractUsageSection(skill.content)
		if usage != "" {
			fmt.Println()
			fmt.Println(ui.Subtitle.Render(fmt.Sprintf("  Quick Start: %s", skill.name)))
			fmt.Println(ui.Divider(50))
			for _, line := range strings.Split(usage, "\n") {
				fmt.Println("  " + line)
			}
		}
	}

	fmt.Println(ui.PageFooter())
}

func learnSingleFile(client *fetch.Client, url, filename, source string, paths *config.Paths, extraReqs []detect.Requirement) {
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
	installArtifactWithExtraReqs(art, paths, extraReqs)
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
	learnSingleFile(client, src.URL, filename, src.Original, paths, nil)
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
	var skipped []struct {
		name   string
		reason string
	}
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
			skipped = append(skipped, struct {
				name   string
				reason string
			}{entry.Name(), fmt.Sprintf("read failed: %v", err)})
			continue
		}

		art, err := parseArtifact(content, entry.Name(), filePath)
		if err != nil {
			fmt.Println(ui.Warning.Render(fmt.Sprintf("  Skipping %s: %v", entry.Name(), err)))
			skipped = append(skipped, struct {
				name   string
				reason string
			}{entry.Name(), fmt.Sprintf("parse failed: %v", err)})
			continue
		}

		art.Source = src.Original
		installArtifactQuiet(art, paths)
		installed = append(installed, art.Name)
	}

	if len(installed) == 0 && len(skipped) == 0 {
		exitWithError("no artifacts found in directory")
	}

	if len(installed) == 0 && len(skipped) > 0 {
		// Found artifacts but all failed
		fmt.Println()
		fmt.Println(ui.Warning.Render(fmt.Sprintf("  Skipped %d artifact(s):", len(skipped))))
		for _, s := range skipped {
			fmt.Println(ui.Muted.Render(fmt.Sprintf("    • %s: %s", s.name, s.reason)))
		}
		exitWithError("no artifacts were installed successfully")
	}

	// Summary
	fmt.Println()
	fmt.Println(ui.SuccessLine(fmt.Sprintf("Inscribed %d artifact(s)", len(installed))))
	for _, name := range installed {
		fmt.Println(ui.Muted.Render("    • " + name))
	}

	// Report any skipped artifacts
	if len(skipped) > 0 {
		fmt.Println()
		fmt.Println(ui.Warning.Render(fmt.Sprintf("  Skipped %d artifact(s):", len(skipped))))
		for _, s := range skipped {
			fmt.Println(ui.Muted.Render(fmt.Sprintf("    • %s: %s", s.name, s.reason)))
		}
	}

	fmt.Println()
	fmt.Println(ui.Dim.Render("  Your tome grows stronger."))
	fmt.Println(ui.PageFooter())
}

func installArtifact(art *artifact.Artifact, paths *config.Paths) {
	installArtifactWithExtraReqs(art, paths, nil)
}

func installArtifactWithExtraReqs(art *artifact.Artifact, paths *config.Paths, extraReqs []detect.Requirement) {
	reqs := doInstallWithExtraReqs(art, paths, nil, extraReqs)

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

	// Display detected requirements
	displayDetectedRequirements(art.Name, reqs)

	fmt.Println()
	fmt.Println(ui.Dim.Render("  Your tome grows stronger."))
	fmt.Println(ui.PageFooter())
}

func installArtifactQuiet(art *artifact.Artifact, paths *config.Paths) {
	installArtifactQuietWithExtras(art, paths, nil, nil)
}

func installArtifactQuietWithIncludes(art *artifact.Artifact, paths *config.Paths, includes []fetch.IncludedFile) {
	installArtifactQuietWithExtras(art, paths, includes, nil)
}

func installArtifactQuietWithExtras(art *artifact.Artifact, paths *config.Paths, includes []fetch.IncludedFile, extraReqs []detect.Requirement) []detect.Requirement {
	reqs := doInstallWithExtraReqs(art, paths, includes, extraReqs)

	badge := getBadge(art.Type)
	name := art.Name
	if len(includes) > 0 {
		name = fmt.Sprintf("%s (+%d files)", art.Name, len(includes))
	}
	fmt.Printf("  %s %s\n", badge, ui.Highlight.Render(name))
	return reqs
}

func doInstall(art *artifact.Artifact, paths *config.Paths) []detect.Requirement {
	return doInstallWithExtraReqs(art, paths, nil, nil)
}

func doInstallWithExtraReqs(art *artifact.Artifact, paths *config.Paths, includes []fetch.IncludedFile, extraReqs []detect.Requirement) []detect.Requirement {
	reqs := doInstallWithIncludes(art, paths, includes)
	// Merge extra requirements (e.g., from README)
	if len(extraReqs) > 0 {
		reqs = detect.Merge(reqs, extraReqs)
		// Update the state with merged requirements
		state, err := config.LoadState(paths.StateFile)
		if err == nil {
			if installed := state.FindInstalled(art.Name); installed != nil {
				installed.Requirements = reqs
				config.SaveState(paths.StateFile, state)
			}
		}
	}
	return reqs
}

func doInstallWithIncludes(art *artifact.Artifact, paths *config.Paths, includes []fetch.IncludedFile) []detect.Requirement {
	installPath := getInstallPath(art, paths)
	installDir := filepath.Dir(installPath)

	// Create directory if needed
	if err := os.MkdirAll(installDir, 0755); err != nil {
		exitWithError(fmt.Sprintf("failed to create directory: %v", err))
	}

	// Write the main file
	if err := os.WriteFile(installPath, []byte(art.Content), 0644); err != nil {
		exitWithError(fmt.Sprintf("failed to write file: %v", err))
	}

	// Collect include paths for requirement detection
	var includePaths []string

	// Write included files (for skills)
	if art.Type == artifact.TypeSkill && len(includes) > 0 {
		skillDir := filepath.Dir(installPath)
		for _, inc := range includes {
			incPath := filepath.Join(skillDir, inc.Path)
			incDir := filepath.Dir(incPath)
			includePaths = append(includePaths, inc.Path)

			// Create subdirectory if needed
			if err := os.MkdirAll(incDir, 0755); err != nil {
				exitWithError(fmt.Sprintf("failed to create directory for %s: %v", inc.Path, err))
			}

			// Write the included file
			if err := os.WriteFile(incPath, inc.Content, 0644); err != nil {
				exitWithError(fmt.Sprintf("failed to write %s: %v", inc.Path, err))
			}

			// Make scripts executable
			if isScript(inc.Path, inc.Content) {
				os.Chmod(incPath, 0755)
			}
		}
	}

	// Detect requirements from content and includes
	contentReqs := detect.FromContent(art.Content)
	includeReqs := detect.FromIncludes(includePaths)
	allReqs := detect.Merge(contentReqs, includeReqs)

	// Update state
	state, err := config.LoadState(paths.StateFile)
	if err != nil {
		exitWithError(fmt.Sprintf("failed to load state: %v", err))
	}

	installed := artifact.InstalledArtifact{
		Artifact:     *art,
		LocalPath:    installPath,
		Requirements: allReqs,
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

	return allReqs
}

func getInstallPath(art *artifact.Artifact, paths *config.Paths) string {
	switch art.Type {
	case artifact.TypeSkill:
		// Skills go in a directory with SKILL.md
		safeDir := fetch.SanitizeFilename(art.Name)
		return filepath.Join(paths.SkillsDir, safeDir, artifact.SkillFilename)
	case artifact.TypeCommand:
		// Commands are just .md files
		safeName := fetch.SanitizeFilename(art.Name) + ".md"
		return filepath.Join(paths.CommandsDir, safeName)
	case artifact.TypeAgent:
		// Agents are .md files in agents/
		agentCfg := config.GetAgentConfig(paths.Agent)
		if agentCfg != nil && agentCfg.AgentsDir != "" {
			agentsDir := filepath.Join(paths.AgentDir, agentCfg.AgentsDir)
			safeName := fetch.SanitizeFilename(art.Name) + ".md"
			return filepath.Join(agentsDir, safeName)
		}
		// Fallback to commands if agents not supported
		safeName := fetch.SanitizeFilename(art.Name) + ".md"
		return filepath.Join(paths.CommandsDir, safeName)
	default:
		return filepath.Join(paths.CommandsDir, art.Filename)
	}
}

// learnPlugin handles installing a plugin and all its artifacts
func learnPlugin(client *fetch.Client, src *source.Source, apiURL string, paths *config.Paths) {
	fmt.Println(ui.PluginBadge + "  " + ui.Info.Render("Plugin detected"))
	fmt.Println(ui.Muted.Render(fmt.Sprintf("    %s/%s", src.Owner, src.Repo)))
	fmt.Println()

	// Fetch the complete plugin
	plugin, err := client.FetchPlugin(apiURL, src.String())
	if err != nil {
		exitWithError(fmt.Sprintf("failed to fetch plugin: %v", err))
	}

	// Display plugin info
	fmt.Println(ui.Highlight.Render("  " + plugin.Manifest.Name))
	if plugin.Manifest.Description != "" {
		fmt.Println(ui.Muted.Render("    " + plugin.Manifest.Description))
	}
	if plugin.Manifest.Author.Name != "" {
		fmt.Println(ui.Muted.Render(fmt.Sprintf("    by %s", plugin.Manifest.Author.Name)))
	}
	if plugin.Manifest.Version != "" {
		fmt.Println(ui.Muted.Render(fmt.Sprintf("    v%s", plugin.Manifest.Version)))
	}
	fmt.Println()

	// Count artifacts
	totalArtifacts := len(plugin.Skills) + len(plugin.Commands) + len(plugin.Agents) + len(plugin.Hooks)
	if totalArtifacts == 0 {
		fmt.Println(ui.Warning.Render("  No artifacts found in plugin"))
		return
	}

	fmt.Println(ui.Muted.Render(fmt.Sprintf("  Found %d artifact(s):", totalArtifacts)))
	if len(plugin.Skills) > 0 {
		fmt.Println(ui.Muted.Render(fmt.Sprintf("    • %d skill(s)", len(plugin.Skills))))
	}
	if len(plugin.Commands) > 0 {
		fmt.Println(ui.Muted.Render(fmt.Sprintf("    • %d command(s)", len(plugin.Commands))))
	}
	if len(plugin.Agents) > 0 {
		fmt.Println(ui.Muted.Render(fmt.Sprintf("    • %d agent(s)", len(plugin.Agents))))
	}
	if len(plugin.Hooks) > 0 {
		fmt.Println(ui.Muted.Render(fmt.Sprintf("    • %d hook(s)", len(plugin.Hooks))))
	}
	fmt.Println()

	// Install all artifacts
	var installed []string

	for _, skill := range plugin.Skills {
		skill.Source = src.String()
		installArtifactQuiet(&skill, paths)
		installed = append(installed, skill.Name)
	}

	for _, cmd := range plugin.Commands {
		cmd.Source = src.String()
		installArtifactQuiet(&cmd, paths)
		installed = append(installed, cmd.Name)
	}

	for _, agent := range plugin.Agents {
		agent.Source = src.String()
		installArtifactQuiet(&agent, paths)
		installed = append(installed, agent.Name)
	}

	// Install hooks to hooks directory
	if len(plugin.Hooks) > 0 {
		agentCfg := config.GetAgentConfig(paths.Agent)
		if agentCfg != nil && agentCfg.HooksDir != "" {
			hooksDir := filepath.Join(paths.AgentDir, agentCfg.HooksDir)
			if err := os.MkdirAll(hooksDir, 0755); err == nil {
				for _, hook := range plugin.Hooks {
					hookPath := filepath.Join(hooksDir, hook.Filename)
					if err := os.WriteFile(hookPath, []byte(hook.Content), 0755); err == nil {
						fmt.Printf("  %s %s\n", ui.HookBadge, ui.Highlight.Render(hook.Name))
						installed = append(installed, hook.Name)
					}
				}
				fmt.Println()
				fmt.Println(ui.Warning.Render("  Note: Add hooks to settings.json to enable them"))
				fmt.Println(ui.Muted.Render(fmt.Sprintf("    Installed to: %s", hooksDir)))
			}
		} else {
			fmt.Println(ui.Warning.Render("  Note: Hooks not supported for this agent"))
		}
	}

	// Summary
	fmt.Println()
	fmt.Println(ui.SuccessLine(fmt.Sprintf("Inscribed %d artifact(s) from plugin", len(installed))))
	for _, name := range installed {
		fmt.Println(ui.Muted.Render("    • " + name))
	}
	fmt.Println()
	fmt.Println(ui.Dim.Render("  Your tome grows stronger."))
	fmt.Println(ui.PageFooter())
}

// extractUsageSection extracts a "Quick Start", "Usage", or "Examples" section from markdown content.
// Returns the section content (without the header) or empty string if not found.
func extractUsageSection(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inSection := false
	sectionLevel := 0

	for _, line := range lines {
		// Check for section headers we care about
		if strings.HasPrefix(line, "#") {
			// Count the header level
			level := 0
			for _, c := range line {
				if c == '#' {
					level++
				} else {
					break
				}
			}

			headerText := strings.ToLower(strings.TrimSpace(strings.TrimLeft(line, "# ")))

			// If we're in a section and hit same or higher level header, stop
			if inSection && level <= sectionLevel {
				break
			}

			// Check if this is a section we want
			if strings.Contains(headerText, "quick start") ||
				strings.Contains(headerText, "usage") ||
				strings.Contains(headerText, "examples") ||
				strings.Contains(headerText, "getting started") {
				inSection = true
				sectionLevel = level
				continue // Skip the header line itself
			}
		}

		if inSection {
			result = append(result, line)
		}
	}

	// Trim leading/trailing empty lines and limit to ~15 lines
	text := strings.TrimSpace(strings.Join(result, "\n"))
	lines = strings.Split(text, "\n")
	if len(lines) > 15 {
		lines = append(lines[:15], "  ...")
	}
	return strings.Join(lines, "\n")
}

// isScript returns true if the file appears to be a script that should be executable.
// Detects by file extension (.sh, .bash, .zsh, .fish, .py, .rb, .pl) or shebang (#!).
func isScript(path string, content []byte) bool {
	// Check extension
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".sh", ".bash", ".zsh", ".fish", ".py", ".rb", ".pl":
		return true
	}

	// Check for shebang
	if len(content) >= 2 && content[0] == '#' && content[1] == '!' {
		return true
	}

	return false
}

// displayDetectedRequirements shows any detected setup requirements after install
func displayDetectedRequirements(name string, reqs []detect.Requirement) {
	if len(reqs) == 0 {
		return
	}

	fmt.Println()
	fmt.Println(ui.Warning.Render("  ⚠ Detected setup requirements:"))

	for _, req := range reqs {
		var icon, label string
		switch req.Type {
		case detect.TypeNPM:
			icon = "📦"
			pm := req.PackageManager
			if pm == "" {
				pm = "npm"
			}
			label = fmt.Sprintf("%s: %s", pm, req.Value)
		case detect.TypePip:
			icon = "🐍"
			pm := req.PackageManager
			if pm == "" {
				pm = "pip"
			}
			label = fmt.Sprintf("%s: %s", pm, req.Value)
		case detect.TypeBrew:
			icon = "🍺"
			label = fmt.Sprintf("brew: %s", req.Value)
		case detect.TypeCargo:
			icon = "🦀"
			label = fmt.Sprintf("cargo: %s", req.Value)
		case detect.TypeEnv:
			icon = "🔑"
			label = fmt.Sprintf("env: %s", req.Value)
		case detect.TypeRuntime:
			icon = "⚙️"
			label = fmt.Sprintf("runtime: %s", req.Value)
		case detect.TypeCommand:
			icon = "💻"
			label = fmt.Sprintf("command: %s", req.Value)
		default:
			icon = "•"
			label = req.Value
		}

		// Show source if from includes
		source := ""
		if strings.HasPrefix(req.Source, "include:") {
			source = ui.Dim.Render(fmt.Sprintf(" (from %s)", strings.TrimPrefix(req.Source, "include:")))
		} else if req.Line > 0 {
			source = ui.Dim.Render(fmt.Sprintf(" (line %d)", req.Line))
		}

		fmt.Printf("    %s %s%s\n", icon, label, source)
	}

	fmt.Println()
	fmt.Printf("  Run: %s\n", ui.Highlight.Render(fmt.Sprintf("tome doctor %s", name)))
}

// showNpmPackageGuidance displays helpful info when a repo is an npm package, not a tome collection
func showNpmPackageGuidance(pkgContent []byte, src *source.Source) {
	// Parse package.json to get package name
	var pkg struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(pkgContent, &pkg); err != nil || pkg.Name == "" {
		// Couldn't parse, use repo name as fallback
		pkg.Name = src.Repo
	}

	fmt.Println()
	fmt.Println(ui.Warning.Render("  This repository is an npm package, not a tome collection."))
	fmt.Println()

	if pkg.Description != "" {
		fmt.Println(ui.Muted.Render("  " + ui.Truncate(pkg.Description, 60)))
		fmt.Println()
	}

	fmt.Println(ui.Info.Render("  To install this package:"))
	fmt.Println()
	fmt.Printf("    %s\n", ui.Highlight.Render(fmt.Sprintf("bun add %s", pkg.Name)))
	fmt.Println(ui.Muted.Render("    or"))
	fmt.Printf("    %s\n", ui.Highlight.Render(fmt.Sprintf("npm install %s", pkg.Name)))
	fmt.Println()
	fmt.Println(ui.Dim.Render("  Tome is for skills, commands, and prompts (markdown artifacts)."))
	fmt.Println(ui.Dim.Render("  Use your package manager for npm/bun packages."))
	fmt.Println(ui.PageFooter())
}
