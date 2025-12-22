package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kennyg/tome/internal/config"
	"github.com/kennyg/tome/internal/detect"
	"github.com/kennyg/tome/internal/ui"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor [name]",
	Short: "Check setup requirements for artifacts",
	Long: `Verify that detected setup requirements are satisfied.

If no artifact name is given, checks all artifacts with requirements.
If a name is given, checks only that artifact.

Examples:
  tome doctor                    # Check all artifacts
  tome doctor open-orchestra     # Check specific artifact`,
	Args: cobra.MaximumNArgs(1),
	Run:  runDoctor,
}

func runDoctor(cmd *cobra.Command, args []string) {
	paths, err := config.GetPaths()
	if err != nil {
		exitWithError(err.Error())
	}

	state, err := config.LoadState(paths.StateFile)
	if err != nil {
		exitWithError(err.Error())
	}

	fmt.Println()
	fmt.Println(ui.SectionHeader("Diagnosing", 56))
	fmt.Println()

	if len(args) == 1 {
		// Check specific artifact
		name := args[0]
		artifact := state.FindInstalled(name)
		if artifact == nil {
			exitWithError(fmt.Sprintf("artifact '%s' not found", name))
		}

		if len(artifact.Requirements) == 0 {
			fmt.Printf("  %s %s\n", ui.Success.Render("✓"), artifact.Name)
			fmt.Println(ui.Muted.Render("    No setup requirements detected"))
			fmt.Println(ui.PageFooter())
			return
		}

		checkArtifact(artifact.Name, artifact.Requirements, true)
	} else {
		// Check all artifacts with requirements
		hasAny := false
		for _, artifact := range state.Installed {
			if len(artifact.Requirements) > 0 {
				hasAny = true
				checkArtifact(artifact.Name, artifact.Requirements, false)
				fmt.Println()
			}
		}

		if !hasAny {
			fmt.Println(ui.Muted.Render("  No artifacts with setup requirements found"))
		}
	}

	fmt.Println(ui.PageFooter())
}

func checkArtifact(name string, reqs []detect.Requirement, verbose bool) {
	results := detect.VerifyAll(reqs)
	allSatisfied := !detect.HasUnsatisfied(results)

	if allSatisfied {
		fmt.Printf("  %s %s\n", ui.Success.Render("✓"), name)
		if verbose {
			fmt.Println(ui.Muted.Render("    All requirements satisfied"))
		}
	} else {
		fmt.Printf("  %s %s\n", ui.Error.Render("✗"), name)
	}

	for _, r := range results {
		if r.Satisfied {
			if verbose {
				fmt.Printf("    %s %s: %s\n",
					ui.Success.Render("✓"),
					r.Requirement.Type,
					r.Requirement.Value)
			}
		} else {
			fmt.Printf("    %s %s: %s\n",
				ui.Error.Render("✗"),
				r.Requirement.Type,
				r.Requirement.Value)
			if r.Message != "" {
				// Indent multi-line messages
				fmt.Println(ui.Muted.Render("      " + r.Message))
			}
		}
	}
}
