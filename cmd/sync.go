package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/kennyg/tome/internal/config"
	"github.com/kennyg/tome/internal/fetch"
	"github.com/kennyg/tome/internal/source"
	"github.com/kennyg/tome/internal/ui"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Update all installed artifacts",
	Long: `Sync all installed artifacts with their sources.

Checks for updates and downloads newer versions if available.`,
	Run: runSync,
}

var syncDry bool

func init() {
	syncCmd.Flags().BoolVar(&syncDry, "dry-run", false, "Check for updates without applying them")
}

func runSync(cmd *cobra.Command, args []string) {
	paths, err := config.GetPaths()
	if err != nil {
		exitWithError(err.Error())
	}

	state, err := config.LoadState(paths.StateFile)
	if err != nil {
		exitWithError(err.Error())
	}

	if len(state.Installed) == 0 {
		fmt.Println()
		fmt.Println(ui.Muted.Render("  No artifacts installed. Nothing to sync."))
		fmt.Println()
		return
	}

	fmt.Println()
	fmt.Println(ui.Title.Render("  Syncing your tome..."))
	fmt.Println()

	client := fetch.NewClient()
	var updated, unchanged, failed int

	for i := range state.Installed {
		a := &state.Installed[i]
		badge := getBadge(a.Type)
		fmt.Printf("  %s %s ", badge, ui.Highlight.Render(a.Name))

		// Determine the URL to fetch
		var fetchURL string

		// Prefer stored source_url if available
		if a.SourceURL != "" {
			// Strip any token params from URL (they expire)
			fetchURL = stripTokenFromURL(a.SourceURL)
		} else {
			// Fall back to parsing source
			src, err := source.Parse(a.Source)
			if err != nil {
				fmt.Println(ui.Warning.Render("⚠ invalid source"))
				failed++
				continue
			}

			switch src.Type {
			case source.TypeGitHub:
				fetchURL = src.GitHubRawURL("")
			case source.TypeURL:
				fetchURL = src.URL
			case source.TypeLocal:
				// Skip local sources - they don't need syncing
				fmt.Println(ui.Muted.Render("↷ local"))
				unchanged++
				continue
			}
		}

		// Fetch current content
		content, err := client.FetchURL(fetchURL)
		if err != nil {
			fmt.Println(ui.Warning.Render("⚠ fetch failed"))
			failed++
			continue
		}

		// Compare with installed version using hash
		newHash := hashContent(content)
		oldHash := a.Hash
		if oldHash == "" {
			// Compute hash from local file for legacy entries
			localContent, err := os.ReadFile(a.LocalPath)
			if err == nil {
				oldHash = hashContent(localContent)
			}
		}

		if newHash == oldHash {
			fmt.Println(ui.Muted.Render("✓ up to date"))
			unchanged++
			continue
		}

		// Update available
		if syncDry {
			fmt.Println(ui.Info.Render("↑ update available"))
			updated++
			continue
		}

		// Apply update
		if err := os.WriteFile(a.LocalPath, content, 0644); err != nil {
			fmt.Println(ui.Warning.Render("⚠ write failed"))
			failed++
			continue
		}

		// Update state
		a.Hash = newHash
		a.UpdatedAt = time.Now()

		fmt.Println(ui.Success.Render("↑ updated"))
		updated++
	}

	// Save state if we made changes
	if updated > 0 && !syncDry {
		if err := config.SaveState(paths.StateFile, state); err != nil {
			fmt.Println()
			fmt.Println(ui.Warning.Render(fmt.Sprintf("  Warning: failed to save state: %v", err)))
		}
	}

	// Summary
	fmt.Println()
	fmt.Println(ui.Divider(50))
	fmt.Println()

	if syncDry {
		fmt.Println(ui.Info.Render("  Dry run complete."))
	} else if updated > 0 {
		fmt.Println(ui.Success.Render(fmt.Sprintf("  Updated %d artifact(s).", updated)))
	} else {
		fmt.Println(ui.Success.Render("  All artifacts are up to date."))
	}

	if failed > 0 {
		fmt.Println(ui.Warning.Render(fmt.Sprintf("  %d artifact(s) failed to sync.", failed)))
	}

	fmt.Println()
}

func hashContent(content []byte) string {
	h := sha256.Sum256(content)
	return hex.EncodeToString(h[:])
}

func stripTokenFromURL(url string) string {
	// Remove ?token=... from GitHub URLs (tokens expire)
	if idx := strings.Index(url, "?token="); idx != -1 {
		return url[:idx]
	}
	return url
}
