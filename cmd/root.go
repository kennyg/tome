package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kennyg/tome/internal/ui"
)

var (
	// Version is set at build time
	Version = "dev"

	// plainOutput forces plain text output without colors/decorations
	plainOutput bool
)

var rootCmd = &cobra.Command{
	Use:     "tome",
	Short:   "AI Agent Skill Manager",
	Version: Version,
	Long: ui.Logo() + `

  Your spellbook for AI agent capabilities.
  Discover, install, and manage skills, commands, and prompts.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if plainOutput {
			ui.IsTTY = false
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVar(&plainOutput, "plain", false, "Force plain text output (no colors/decorations)")

	// Subcommands
	rootCmd.AddCommand(aproposCmd)
	rootCmd.AddCommand(attuneCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(learnCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(buildCmd)
}

var versionCmd = &cobra.Command{
	Use:     "edition",
	Aliases: []string{"version", "ver"},
	Short:   "Show the tome's edition",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("tome %s\n", Version)
	},
}

// exitWithError prints an error and exits
func exitWithError(msg string) {
	fmt.Fprintln(os.Stderr, ui.Error.Render("Error: "+msg))
	os.Exit(1)
}
