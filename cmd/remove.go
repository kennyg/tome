package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	artifactPkg "github.com/kennyg/tome/internal/artifact"
	"github.com/kennyg/tome/internal/config"
	"github.com/kennyg/tome/internal/ui"
)

var removeCmd = &cobra.Command{
	Use:     "remove <name>",
	Aliases: []string{"rm", "unlearn", "forget"},
	Short:   "Remove an installed artifact",
	Long: `Remove (uninstall) an artifact from your tome.

Examples:
  tome remove my-skill
  tome forget deploy-command`,
	Args: cobra.ExactArgs(1),
	Run:  runRemove,
}

func runRemove(cmd *cobra.Command, args []string) {
	name := args[0]

	paths, err := config.GetPaths()
	if err != nil {
		exitWithError(err.Error())
	}

	state, err := config.LoadState(paths.StateFile)
	if err != nil {
		exitWithError(err.Error())
	}

	artifact := state.FindInstalled(name)
	if artifact == nil {
		exitWithError(fmt.Sprintf("artifact '%s' not found", name))
	}

	fmt.Println()
	fmt.Println(ui.Title.Render("  Removing " + name))
	fmt.Println()

	badge := getBadge(artifact.Type)
	fmt.Printf("  %s %s\n", badge, ui.Highlight.Render(artifact.Name))
	fmt.Println(ui.Muted.Render(fmt.Sprintf("    Path: %s", artifact.LocalPath)))
	fmt.Println()

	// Remove the file from disk
	if err := os.Remove(artifact.LocalPath); err != nil && !os.IsNotExist(err) {
		exitWithError(fmt.Sprintf("failed to remove file: %v", err))
	}

	// For skills, also try to remove the parent directory if empty
	if artifact.Type == artifactPkg.TypeSkill {
		parentDir := filepath.Dir(artifact.LocalPath)
		// Only remove if it's a skill-specific directory (not the main skills dir)
		if parentDir != paths.SkillsDir {
			_ = os.Remove(parentDir) // Ignore error - dir may not be empty
		}
	}

	// Update state
	state.RemoveInstalled(artifact.Name, artifact.Type)
	if err := config.SaveState(paths.StateFile, state); err != nil {
		exitWithError(fmt.Sprintf("failed to update state: %v", err))
	}

	fmt.Println(ui.Success.Render("  Removed successfully."))
	fmt.Println()
	fmt.Println(ui.Muted.Render("  Your tome has been lightened."))
	fmt.Println()
}
