package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	artifactPkg "github.com/kennyg/tome/internal/artifact"
	"github.com/kennyg/tome/internal/config"
	"github.com/kennyg/tome/internal/ui"
)

var removeCmd = &cobra.Command{
	Use:     "forget <name>",
	Aliases: []string{"erase", "unlearn", "remove", "rm"},
	Short:   "Erase an inscription from the tome",
	Long: `Forget an artifact, erasing it from your tome.

Examples:
  tome forget my-skill
  tome erase deploy-command`,
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

	// Determine what to remove based on artifact type and path structure
	pathToRemove := artifact.LocalPath
	isDirectory := false
	isSymlink := false

	if artifact.Type == artifactPkg.TypeSkill {
		// Skills can be:
		// 1. Directory-based: skills/<name>/SKILL.md -> remove the directory
		// 2. Symlinks: skills/<name> -> target -> remove the symlink only
		// 3. Flat files: skills/<name>.md -> remove the file
		parentDir := filepath.Dir(artifact.LocalPath)
		filename := filepath.Base(artifact.LocalPath)

		// Check if the parent directory is the skill's own directory (not the main skills dir)
		if parentDir != paths.SkillsDir && strings.HasSuffix(filename, ".md") {
			// Check if parent is a symlink
			if info, err := os.Lstat(parentDir); err == nil {
				if info.Mode()&os.ModeSymlink != 0 {
					// It's a symlink - remove just the symlink, not the target
					pathToRemove = parentDir
					isSymlink = true
					// Note: os.Remove handles symlinks correctly (removes link, not target)
				} else if info.IsDir() {
					// It's a real directory - remove the whole thing
					pathToRemove = parentDir
					isDirectory = true
				}
			}
		}
	}

	// Show what we're actually removing
	if isSymlink {
		target, _ := os.Readlink(pathToRemove)
		fmt.Println(ui.Muted.Render(fmt.Sprintf("    Symlink: %s → %s", pathToRemove, target)))
	} else if isDirectory {
		fmt.Println(ui.Muted.Render(fmt.Sprintf("    Directory: %s", pathToRemove)))
	} else {
		fmt.Println(ui.Muted.Render(fmt.Sprintf("    File: %s", pathToRemove)))
	}
	fmt.Println()

	// Remove from disk
	var removeErr error
	if isDirectory {
		removeErr = os.RemoveAll(pathToRemove)
	} else {
		removeErr = os.Remove(pathToRemove)
	}

	if removeErr != nil && !os.IsNotExist(removeErr) {
		exitWithError(fmt.Sprintf("failed to remove %s: %v", pathToRemove, removeErr))
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
