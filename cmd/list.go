package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/kennyg/tome/internal/artifact"
	"github.com/kennyg/tome/internal/config"
	"github.com/kennyg/tome/internal/ui"
)

var listCmd = &cobra.Command{
	Use:     "index",
	Aliases: []string{"contents", "list", "ls"},
	Short:   "View the tome's index",
	Long:    `Display all inscribed skills, commands, prompts, and hooks.`,
	Run:     runList,
}

var (
	listSkills   bool
	listCommands bool
	listPrompts  bool
	listHooks    bool
)

func init() {
	listCmd.Flags().BoolVar(&listSkills, "skills", false, "Show only skills")
	listCmd.Flags().BoolVar(&listCommands, "commands", false, "Show only commands")
	listCmd.Flags().BoolVar(&listPrompts, "prompts", false, "Show only prompts")
	listCmd.Flags().BoolVar(&listHooks, "hooks", false, "Show only hooks")
}

func runList(cmd *cobra.Command, args []string) {
	paths, err := config.GetPaths()
	if err != nil {
		exitWithError(err.Error())
	}

	state, err := config.LoadState(paths.StateFile)
	if err != nil {
		exitWithError(err.Error())
	}

	// Determine which types to show
	showAll := !listSkills && !listCommands && !listPrompts && !listHooks
	typeFilter := make(map[artifact.Type]bool)
	if showAll || listSkills {
		typeFilter[artifact.TypeSkill] = true
	}
	if showAll || listCommands {
		typeFilter[artifact.TypeCommand] = true
	}
	if showAll || listPrompts {
		typeFilter[artifact.TypePrompt] = true
	}
	if showAll || listHooks {
		typeFilter[artifact.TypeHook] = true
	}

	// Filter and display
	var filtered []artifact.InstalledArtifact
	for _, a := range state.Installed {
		if typeFilter[a.Type] {
			filtered = append(filtered, a)
		}
	}

	if len(filtered) == 0 {
		fmt.Print(ui.EmptyTome())
		return
	}

	// Header
	fmt.Println()
	fmt.Println(ui.SectionHeader("Your Tome", 56))
	fmt.Println()

	// Group by type
	byType := make(map[artifact.Type][]artifact.InstalledArtifact)
	for _, a := range filtered {
		byType[a.Type] = append(byType[a.Type], a)
	}

	// Display each type in a beautiful card layout
	for _, t := range []artifact.Type{artifact.TypeSkill, artifact.TypeCommand, artifact.TypePrompt, artifact.TypeHook} {
		artifacts := byType[t]
		if len(artifacts) == 0 {
			continue
		}

		badge := getBadge(t)
		count := lipgloss.NewStyle().Foreground(ui.DarkGray).Render(fmt.Sprintf("(%d)", len(artifacts)))
		fmt.Printf("  %s %s\n", badge, count)
		fmt.Println()

		for _, a := range artifacts {
			name := lipgloss.NewStyle().Foreground(ui.White).Bold(true).Render(a.Name)
			desc := ui.Truncate(a.Description, 50)
			descStyled := lipgloss.NewStyle().Foreground(ui.Gray).Render(desc)
			fmt.Printf("    %s\n", name)
			fmt.Printf("    %s\n", descStyled)
			fmt.Println()
		}
	}

	// Footer
	total := len(filtered)
	footer := lipgloss.NewStyle().Foreground(ui.DarkGray).Render(fmt.Sprintf("  %d artifact(s) inscribed", total))
	fmt.Println(footer)
	fmt.Println(ui.PageFooter())
}

func getBadge(t artifact.Type) string {
	switch t {
	case artifact.TypeSkill:
		return ui.SkillBadge
	case artifact.TypeCommand:
		return ui.CmdBadge
	case artifact.TypePrompt:
		return ui.PromptBadge
	case artifact.TypeHook:
		return ui.HookBadge
	default:
		return ""
	}
}
