package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kennyg/tome/internal/config"
	"github.com/kennyg/tome/internal/ui"
)

var infoCmd = &cobra.Command{
	Use:     "study <name>",
	Aliases: []string{"examine", "info"},
	Short:   "Study an artifact in detail",
	Long: `Examine the details of an inscribed artifact.

Shows metadata, source, installation date, and contents.`,
	Args: cobra.ExactArgs(1),
	Run:  runInfo,
}

func runInfo(cmd *cobra.Command, args []string) {
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

	badge := getBadge(artifact.Type)

	fmt.Println(ui.Title.Render(artifact.Name))
	fmt.Println()
	fmt.Printf("%s %s\n", badge, ui.Muted.Render(string(artifact.Type)))
	fmt.Println()

	if artifact.Description != "" {
		fmt.Println(artifact.Description)
		fmt.Println()
	}

	fmt.Println(ui.Subtitle.Render("Details"))
	fmt.Println(ui.Divider(40))

	if artifact.Author != "" {
		fmt.Printf("  Author:    %s\n", artifact.Author)
	}
	if artifact.Version != "" {
		fmt.Printf("  Version:   %s\n", artifact.Version)
	}
	fmt.Printf("  Source:    %s\n", artifact.Source)
	fmt.Printf("  Path:      %s\n", artifact.LocalPath)

	if !artifact.InstalledAt.IsZero() {
		fmt.Printf("  Installed: %s\n", artifact.InstalledAt.Format("2006-01-02 15:04"))
	}
}
