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
)

var rootCmd = &cobra.Command{
	Use:   "tome",
	Short: "AI Agent Skill Manager",
	Long: ui.Logo() + `

  Your spellbook for AI agent capabilities.
  Discover, install, and manage skills, commands, and prompts.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(attuneCmd)
	rootCmd.AddCommand(learnCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("tome %s\n", Version)
	},
}

// exitWithError prints an error and exits
func exitWithError(msg string) {
	fmt.Fprintln(os.Stderr, ui.Error.Render("Error: "+msg))
	os.Exit(1)
}
