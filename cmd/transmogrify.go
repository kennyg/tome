package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kennyg/tome/internal/fetch"
	"github.com/kennyg/tome/internal/schema"
	"github.com/kennyg/tome/internal/source"
	"github.com/kennyg/tome/internal/ui"
)

var transmogrifyCmd = &cobra.Command{
	Use:     "transmogrify <source>",
	Aliases: []string{"convert", "morph", "xform"},
	Short:   "Convert artifacts between agent formats",
	Long: `Convert skill artifacts between different AI agent formats.

Supports conversion between:
  - Claude Code (skills/*/SKILL.md)
  - OpenCode (.opencode/skill/*/SKILL.md)
  - GitHub Copilot (agents/*.agent.md)
  - Cursor (.cursor/rules/*.md)

Sources can be:
  - Local file path
  - Local directory
  - GitHub repository (owner/repo)

Examples:
  tome transmogrify agents/CSharp.agent.md --to claude
  tome transmogrify ./copilot-skills/ --to claude --output ./converted/
  tome transmogrify github/awesome-copilot --to claude --dry-run`,
	Args: cobra.ExactArgs(1),
	Run:  runTransmogrify,
}

var (
	transmogrifyTo     string
	transmogrifyOutput string
	transmogrifyDryRun bool
	transmogrifyForce  bool
)

func init() {
	transmogrifyCmd.Flags().StringVar(&transmogrifyTo, "to", "", "Target format (claude, opencode, copilot, cursor)")
	transmogrifyCmd.Flags().StringVarP(&transmogrifyOutput, "output", "o", "", "Output directory (default: stdout for single file)")
	transmogrifyCmd.Flags().BoolVar(&transmogrifyDryRun, "dry-run", false, "Show what would be converted without doing it")
	transmogrifyCmd.Flags().BoolVarP(&transmogrifyForce, "force", "f", false, "Overwrite existing files")

	transmogrifyCmd.MarkFlagRequired("to")

	rootCmd.AddCommand(transmogrifyCmd)
}

func runTransmogrify(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(ui.SectionHeader("Transmogrify", 56))
	fmt.Println()

	// Validate target format
	targetFormat := schema.Format(transmogrifyTo)
	if !targetFormat.IsValid() {
		exitWithError(fmt.Sprintf("invalid target format: %s (valid: claude, opencode, copilot, cursor)", transmogrifyTo))
	}

	sourceArg := args[0]

	// Determine source type
	src, err := source.Parse(sourceArg)
	if err != nil {
		// Might be a local path
		if _, statErr := os.Stat(sourceArg); statErr == nil {
			transmogrifyLocal(sourceArg, targetFormat)
			return
		}
		exitWithError(err.Error())
	}

	switch src.Type {
	case source.TypeGitHub:
		transmogrifyGitHub(src, targetFormat)
	case source.TypeLocal:
		transmogrifyLocal(src.Path, targetFormat)
	default:
		exitWithError("unsupported source type")
	}
}

func transmogrifyLocal(path string, targetFormat schema.Format) {
	info, err := os.Stat(path)
	if err != nil {
		exitWithError(fmt.Sprintf("cannot access %s: %v", path, err))
	}

	if info.IsDir() {
		transmogrifyDirectory(path, targetFormat)
	} else {
		transmogrifyFile(path, targetFormat)
	}
}

func transmogrifyFile(path string, targetFormat schema.Format) {
	fmt.Println(ui.InfoLine(fmt.Sprintf("Source: %s", path)))
	fmt.Println(ui.InfoLine(fmt.Sprintf("Target: %s", targetFormat)))
	fmt.Println()

	// Read file
	content, err := os.ReadFile(path)
	if err != nil {
		exitWithError(fmt.Sprintf("failed to read file: %v", err))
	}

	// Parse (auto-detect format)
	skill, err := schema.ParseAuto(content, path)
	if err != nil {
		exitWithError(fmt.Sprintf("failed to parse: %v", err))
	}

	fmt.Println(ui.Muted.Render(fmt.Sprintf("  Detected format: %s", skill.GetFormat())))
	fmt.Println(ui.Muted.Render(fmt.Sprintf("  Skill name: %s", skill.GetName())))
	fmt.Println()

	// Convert
	result, err := schema.ConvertWithInfo(skill, targetFormat)
	if err != nil {
		exitWithError(fmt.Sprintf("conversion failed: %v", err))
	}

	// Show warnings
	for _, w := range result.Warnings {
		fmt.Println(ui.WarningLine(w))
	}

	if transmogrifyDryRun {
		fmt.Println(ui.Muted.Render("  [dry-run] Would convert:"))
		fmt.Println(ui.Muted.Render(fmt.Sprintf("    %s → %s", result.SourceFormat, result.TargetFormat)))
		fmt.Println()
		fmt.Println(ui.SuccessLine("Dry run complete"))
		fmt.Println(ui.PageFooter())
		return
	}

	// Output
	if transmogrifyOutput == "" {
		// Print to stdout
		fmt.Println(ui.Muted.Render("  Output:"))
		fmt.Println()
		fmt.Println(string(result.Content))
	} else {
		// Write to file
		outDir := transmogrifyOutput
		if targetFormat == schema.FormatClaude || targetFormat == schema.FormatOpenCode {
			// Create skill directory structure
			outDir = filepath.Join(transmogrifyOutput, schema.OutputDirectory(skill, targetFormat))
		}

		if err := os.MkdirAll(outDir, 0755); err != nil {
			exitWithError(fmt.Sprintf("failed to create output directory: %v", err))
		}

		outPath := filepath.Join(outDir, schema.OutputFilename(skill, targetFormat))

		// Check if exists
		if !transmogrifyForce {
			if _, err := os.Stat(outPath); err == nil {
				exitWithError(fmt.Sprintf("output file exists: %s (use --force to overwrite)", outPath))
			}
		}

		if err := os.WriteFile(outPath, result.Content, 0644); err != nil {
			exitWithError(fmt.Sprintf("failed to write file: %v", err))
		}

		fmt.Println(ui.SuccessLine(fmt.Sprintf("Wrote %s", outPath)))
	}

	fmt.Println(ui.PageFooter())
}

func transmogrifyDirectory(path string, targetFormat schema.Format) {
	fmt.Println(ui.InfoLine(fmt.Sprintf("Source: %s/", path)))
	fmt.Println(ui.InfoLine(fmt.Sprintf("Target: %s", targetFormat)))
	fmt.Println()

	// Find all potential skill files
	var files []string
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		// Check for known patterns
		base := filepath.Base(p)
		if strings.EqualFold(base, "SKILL.md") ||
			strings.HasSuffix(base, ".agent.md") ||
			strings.HasSuffix(base, ".prompt.md") {
			files = append(files, p)
		}
		return nil
	})
	if err != nil {
		exitWithError(fmt.Sprintf("failed to scan directory: %v", err))
	}

	if len(files) == 0 {
		fmt.Println(ui.WarningLine("No convertible files found"))
		fmt.Println(ui.PageFooter())
		return
	}

	fmt.Println(ui.Muted.Render(fmt.Sprintf("  Found %d file(s)", len(files))))
	fmt.Println()

	var converted, failed int
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Println(ui.Warning.Render(fmt.Sprintf("  ! %s: %v", filepath.Base(file), err)))
			failed++
			continue
		}

		skill, err := schema.ParseAuto(content, file)
		if err != nil {
			fmt.Println(ui.Warning.Render(fmt.Sprintf("  ! %s: %v", filepath.Base(file), err)))
			failed++
			continue
		}

		result, err := schema.ConvertWithInfo(skill, targetFormat)
		if err != nil {
			fmt.Println(ui.Warning.Render(fmt.Sprintf("  ! %s: %v", filepath.Base(file), err)))
			failed++
			continue
		}

		relPath, _ := filepath.Rel(path, file)

		if transmogrifyDryRun {
			fmt.Printf("  %s %s → %s\n",
				ui.Success.Render("✓"),
				relPath,
				schema.OutputFilename(skill, targetFormat))
			converted++
			continue
		}

		if transmogrifyOutput != "" {
			outDir := filepath.Join(transmogrifyOutput, schema.OutputDirectory(skill, targetFormat))
			if err := os.MkdirAll(outDir, 0755); err != nil {
				fmt.Println(ui.Warning.Render(fmt.Sprintf("  ! %s: %v", skill.GetName(), err)))
				failed++
				continue
			}

			outPath := filepath.Join(outDir, schema.OutputFilename(skill, targetFormat))
			if err := os.WriteFile(outPath, result.Content, 0644); err != nil {
				fmt.Println(ui.Warning.Render(fmt.Sprintf("  ! %s: %v", skill.GetName(), err)))
				failed++
				continue
			}

			fmt.Printf("  %s %s → %s\n",
				ui.Success.Render("✓"),
				relPath,
				outPath)
		} else {
			fmt.Printf("  %s %s\n", ui.Success.Render("✓"), relPath)
		}
		converted++
	}

	fmt.Println()
	if transmogrifyDryRun {
		fmt.Println(ui.SuccessLine(fmt.Sprintf("Would convert %d file(s)", converted)))
	} else {
		fmt.Println(ui.SuccessLine(fmt.Sprintf("Converted %d file(s)", converted)))
	}
	if failed > 0 {
		fmt.Println(ui.WarningLine(fmt.Sprintf("%d file(s) failed", failed)))
	}
	fmt.Println(ui.PageFooter())
}

func transmogrifyGitHub(src *source.Source, targetFormat schema.Format) {
	fmt.Println(ui.InfoLine(fmt.Sprintf("Source: %s", src.String())))
	fmt.Println(ui.InfoLine(fmt.Sprintf("Target: %s", targetFormat)))
	fmt.Println()

	client := fetch.NewClient()
	apiURL := src.GitHubAPIURL()

	fmt.Println(ui.Muted.Render("  Scanning repository..."))

	artifacts, err := client.FindArtifacts(apiURL)
	if err != nil {
		exitWithError(fmt.Sprintf("failed to scan repository: %v", err))
	}

	if len(artifacts) == 0 {
		fmt.Println(ui.WarningLine("No convertible artifacts found"))
		fmt.Println(ui.PageFooter())
		return
	}

	fmt.Println(ui.Muted.Render(fmt.Sprintf("  Found %d artifact(s)", len(artifacts))))
	fmt.Println()

	var converted, failed int
	for _, item := range artifacts {
		url := item.DownloadURL
		if url == "" {
			url = src.GitHubRawURL(item.Path)
		}

		content, err := client.FetchURL(url)
		if err != nil {
			fmt.Println(ui.Warning.Render(fmt.Sprintf("  ! %s: %v", item.Name, err)))
			failed++
			continue
		}

		skill, err := schema.ParseAuto(content, item.Name)
		if err != nil {
			fmt.Println(ui.Warning.Render(fmt.Sprintf("  ! %s: %v", item.Name, err)))
			failed++
			continue
		}

		result, err := schema.ConvertWithInfo(skill, targetFormat)
		if err != nil {
			fmt.Println(ui.Warning.Render(fmt.Sprintf("  ! %s: %v", item.Name, err)))
			failed++
			continue
		}

		if transmogrifyDryRun {
			fmt.Printf("  %s %s (%s → %s)\n",
				ui.Success.Render("✓"),
				skill.GetName(),
				result.SourceFormat,
				result.TargetFormat)
			converted++
			continue
		}

		if transmogrifyOutput != "" {
			outDir := filepath.Join(transmogrifyOutput, schema.OutputDirectory(skill, targetFormat))
			if err := os.MkdirAll(outDir, 0755); err != nil {
				fmt.Println(ui.Warning.Render(fmt.Sprintf("  ! %s: %v", skill.GetName(), err)))
				failed++
				continue
			}

			outPath := filepath.Join(outDir, schema.OutputFilename(skill, targetFormat))
			if err := os.WriteFile(outPath, result.Content, 0644); err != nil {
				fmt.Println(ui.Warning.Render(fmt.Sprintf("  ! %s: %v", skill.GetName(), err)))
				failed++
				continue
			}

			fmt.Printf("  %s %s → %s\n",
				ui.Success.Render("✓"),
				skill.GetName(),
				outPath)
		} else {
			// Just print the converted content
			fmt.Printf("  %s %s\n", ui.Success.Render("✓"), skill.GetName())
		}
		converted++
	}

	fmt.Println()
	if transmogrifyDryRun {
		fmt.Println(ui.SuccessLine(fmt.Sprintf("Would convert %d artifact(s)", converted)))
	} else {
		fmt.Println(ui.SuccessLine(fmt.Sprintf("Converted %d artifact(s)", converted)))
	}
	if failed > 0 {
		fmt.Println(ui.WarningLine(fmt.Sprintf("%d artifact(s) failed", failed)))
	}
	fmt.Println(ui.PageFooter())
}
