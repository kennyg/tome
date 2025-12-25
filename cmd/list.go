package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/kennyg/tome/internal/artifact"
	"github.com/kennyg/tome/internal/config"
	"github.com/kennyg/tome/internal/detect"
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
	listShort    bool
)

func init() {
	listCmd.Flags().BoolVar(&listSkills, "skills", false, "Show only skills")
	listCmd.Flags().BoolVar(&listCommands, "commands", false, "Show only commands")
	listCmd.Flags().BoolVar(&listPrompts, "prompts", false, "Show only prompts")
	listCmd.Flags().BoolVar(&listHooks, "hooks", false, "Show only hooks")
	listCmd.Flags().BoolVar(&listShort, "short", false, "Truncate descriptions to one line")
}

// artifactWithLocation tracks an artifact and where it's from
type artifactWithLocation struct {
	artifact.InstalledArtifact
	Location string // "project" or "global"
	InEffect bool   // true if this is the active version
}

func runList(cmd *cobra.Command, args []string) {
	agent := config.DefaultAgent()

	// Collect artifacts from both locations
	var allArtifacts []artifactWithLocation
	seenNames := make(map[string]bool) // track which names we've seen (for in-effect logic)

	// First, load project-local artifacts (they take precedence)
	if config.IsAttuned(agent) {
		localPaths, err := config.GetLocalPaths(agent)
		if err == nil {
			localState, err := config.LoadState(localPaths.StateFile)
			if err == nil {
				for _, a := range localState.Installed {
					key := fmt.Sprintf("%s:%s", a.Type, a.Name)
					seenNames[key] = true
					allArtifacts = append(allArtifacts, artifactWithLocation{
						InstalledArtifact: a,
						Location:          "project",
						InEffect:          true, // local always in effect
					})
				}
			}
		}
	}

	// Then load global artifacts
	globalPaths, err := config.GetPathsForAgent(agent)
	if err != nil {
		exitWithError(err.Error())
	}

	globalState, err := config.LoadState(globalPaths.StateFile)
	if err != nil {
		exitWithError(err.Error())
	}

	for _, a := range globalState.Installed {
		key := fmt.Sprintf("%s:%s", a.Type, a.Name)
		inEffect := !seenNames[key] // only in effect if not shadowed by local
		allArtifacts = append(allArtifacts, artifactWithLocation{
			InstalledArtifact: a,
			Location:          "global",
			InEffect:          inEffect,
		})
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

	// Filter
	var filtered []artifactWithLocation
	for _, a := range allArtifacts {
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
	byType := make(map[artifact.Type][]artifactWithLocation)
	for _, a := range filtered {
		byType[a.Type] = append(byType[a.Type], a)
	}

	// Display each type
	for _, t := range []artifact.Type{artifact.TypeSkill, artifact.TypeCommand, artifact.TypePrompt, artifact.TypeHook} {
		artifacts := byType[t]
		if len(artifacts) == 0 {
			continue
		}

		badge := getBadge(t)
		count := lipgloss.NewStyle().Foreground(ui.DarkGray).Render(fmt.Sprintf("(%d)", len(artifacts)))
		fmt.Printf("  %s %s\n", badge, count)
		fmt.Println()

		descWidth := ui.DescriptionWidth() - 12 // account for location tag
		for _, a := range artifacts {
			// Check if setup is needed
			needsSetup := false
			if len(a.Requirements) > 0 && !a.SetupDone {
				results := detect.VerifyAll(a.Requirements)
				needsSetup = detect.HasUnsatisfied(results)
			}

			// Format location tag
			var locTag string
			if a.InEffect {
				locTag = lipgloss.NewStyle().Foreground(ui.Green).Render(fmt.Sprintf("[%s]", a.Location))
			} else {
				locTag = lipgloss.NewStyle().Foreground(ui.DarkGray).Render(fmt.Sprintf("[%s]", a.Location))
			}

			// Format setup indicator
			setupTag := ""
			if needsSetup {
				setupTag = " " + lipgloss.NewStyle().Foreground(ui.Amber).Render("[needs setup]")
			}

			// Format name (dim if not in effect)
			var name string
			if a.InEffect {
				name = lipgloss.NewStyle().Foreground(ui.White).Bold(true).Render(a.Name)
			} else {
				name = lipgloss.NewStyle().Foreground(ui.DarkGray).Render(a.Name)
			}

			fmt.Printf("    %s %s%s\n", name, locTag, setupTag)

			// Display description: wrap if --full, truncate otherwise
			descStyle := lipgloss.NewStyle().Foreground(ui.Gray)
			if !a.InEffect {
				descStyle = lipgloss.NewStyle().Foreground(ui.DarkGray)
			}

			if listShort {
				// Truncate to single line
				desc := ui.Truncate(a.Description, descWidth)
				fmt.Printf("    %s\n", descStyle.Render(desc))
			} else {
				// Wrap description to multiple lines (default)
				lines := ui.WrapText(a.Description, descWidth, "    ")
				for _, line := range lines {
					fmt.Printf("    %s\n", descStyle.Render(line))
				}
			}
			fmt.Println()
		}
	}

	// Footer with counts
	var projectInEffect, globalInEffect, shadowedCount int
	for _, a := range filtered {
		if a.InEffect {
			if a.Location == "project" {
				projectInEffect++
			} else {
				globalInEffect++
			}
		} else {
			shadowedCount++
		}
	}

	total := projectInEffect + globalInEffect
	var footer string
	if shadowedCount > 0 {
		footer = fmt.Sprintf("  %d in effect (%d project, %d global, %d shadowed)", total, projectInEffect, globalInEffect, shadowedCount)
	} else {
		footer = fmt.Sprintf("  %d in effect (%d project, %d global)", total, projectInEffect, globalInEffect)
	}
	fmt.Println(lipgloss.NewStyle().Foreground(ui.DarkGray).Render(footer))
	fmt.Println(ui.PageFooter())
}

func getBadge(t artifact.Type) string {
	switch t {
	case artifact.TypeSkill:
		return ui.SkillBadge()
	case artifact.TypeCommand:
		return ui.CmdBadge()
	case artifact.TypePrompt:
		return ui.PromptBadge()
	case artifact.TypeHook:
		return ui.HookBadge()
	case artifact.TypeAgent:
		return ui.AgentBadge()
	case artifact.TypePlugin:
		return ui.PluginBadge()
	default:
		return ""
	}
}
