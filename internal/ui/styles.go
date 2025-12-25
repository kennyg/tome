package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
)

// IsTTY indicates whether stdout is an interactive terminal.
// When false, UI functions produce plain text without colors or decorations.
var IsTTY = term.IsTerminal(os.Stdout.Fd())

// ═══════════════════════════════════════════════════════════════════════════════
// COLOR PALETTE - Ancient tome meets modern terminal
// ═══════════════════════════════════════════════════════════════════════════════

var (
	// Gradient colors for the logo and accents
	Gradient1 = lipgloss.Color("#FF6B6B") // Warm coral
	Gradient2 = lipgloss.Color("#C44569") // Deep rose
	Gradient3 = lipgloss.Color("#6C5CE7") // Royal purple
	Gradient4 = lipgloss.Color("#A29BFE") // Soft lavender

	// Primary palette - rich and warm
	Gold       = lipgloss.Color("#F4D03F") // Bright gold
	Amber      = lipgloss.Color("#E59866") // Warm amber
	Bronze     = lipgloss.Color("#CD6155") // Deep bronze
	Copper     = lipgloss.Color("#DC7633") // Copper accent
	Parchment  = lipgloss.Color("#FAE5D3") // Light parchment
	Sepia      = lipgloss.Color("#A67B5B") // Sepia tone
	DarkBrown  = lipgloss.Color("#5D4037") // Dark leather

	// Accent colors - magical elements
	Purple     = lipgloss.Color("#9B59B6") // Mystical purple
	Violet     = lipgloss.Color("#8E44AD") // Deep violet
	Blue       = lipgloss.Color("#5DADE2") // Arcane blue
	Cyan       = lipgloss.Color("#76D7C4") // Ethereal cyan
	Green      = lipgloss.Color("#58D68D") // Nature green
	Emerald    = lipgloss.Color("#27AE60") // Deep emerald
	Pink       = lipgloss.Color("#FF6B9D") // Enchanted pink
	Magenta    = lipgloss.Color("#E91E8C") // Vivid magenta

	// Neutrals
	White      = lipgloss.Color("#FDFEFE")
	LightGray  = lipgloss.Color("#D5D8DC")
	Gray       = lipgloss.Color("#AAB7B8")
	DarkGray   = lipgloss.Color("#5D6D7E")
	Charcoal   = lipgloss.Color("#2C3E50")
	Black      = lipgloss.Color("#1C2833")
)

// ═══════════════════════════════════════════════════════════════════════════════
// TEXT STYLES
// ═══════════════════════════════════════════════════════════════════════════════

var (
	// Title - gradient effect simulated with bold gold
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(Gold)

	// Subtitle for secondary headings
	Subtitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(Amber)

	// Success messages
	Success = lipgloss.NewStyle().
		Foreground(Green)

	// Error messages
	Error = lipgloss.NewStyle().
		Foreground(Pink).
		Bold(true)

	// Warning messages
	Warning = lipgloss.NewStyle().
		Foreground(Copper)

	// Info messages
	Info = lipgloss.NewStyle().
		Foreground(Blue)

	// Muted/secondary text
	Muted = lipgloss.NewStyle().
		Foreground(Gray)

	// Dim - even more subtle
	Dim = lipgloss.NewStyle().
		Foreground(DarkGray)

	// Highlight for important items
	Highlight = lipgloss.NewStyle().
		Foreground(Gold).
		Bold(true)

	// Link style
	Link = lipgloss.NewStyle().
		Foreground(Cyan).
		Underline(true)

	// Code/command style
	Code = lipgloss.NewStyle().
		Foreground(Magenta)
)

// ═══════════════════════════════════════════════════════════════════════════════
// PANEL STYLES - Card-like containers
// ═══════════════════════════════════════════════════════════════════════════════

var (
	// Panel border style
	panelBorder = lipgloss.RoundedBorder()

	// Main panel - elegant rounded box
	Panel = lipgloss.NewStyle().
		Border(panelBorder).
		BorderForeground(DarkGray).
		Padding(0, 1)

	// Highlighted panel
	PanelHighlight = lipgloss.NewStyle().
		Border(panelBorder).
		BorderForeground(Gold).
		Padding(0, 1)

	// Success panel
	PanelSuccess = lipgloss.NewStyle().
		Border(panelBorder).
		BorderForeground(Green).
		Padding(0, 1)

	// Error panel
	PanelError = lipgloss.NewStyle().
		Border(panelBorder).
		BorderForeground(Pink).
		Padding(0, 1)
)

// ═══════════════════════════════════════════════════════════════════════════════
// BADGES - Type indicators with flair
// ═══════════════════════════════════════════════════════════════════════════════

var baseBadge = lipgloss.NewStyle().
	Padding(0, 1).
	Bold(true)

// Artifact type badge functions

// SkillBadge returns the skill type badge
func SkillBadge() string {
	if !IsTTY {
		return "[SKILL]"
	}
	return baseBadge.Background(Purple).Foreground(White).Render("✦ SKILL")
}

// CmdBadge returns the command type badge
func CmdBadge() string {
	if !IsTTY {
		return "[CMD]"
	}
	return baseBadge.Background(Blue).Foreground(White).Render("⌘ CMD")
}

// PromptBadge returns the prompt type badge
func PromptBadge() string {
	if !IsTTY {
		return "[PROMPT]"
	}
	return baseBadge.Background(Emerald).Foreground(White).Render("✎ PROMPT")
}

// HookBadge returns the hook type badge
func HookBadge() string {
	if !IsTTY {
		return "[HOOK]"
	}
	return baseBadge.Background(Copper).Foreground(White).Render("⚡ HOOK")
}

// AgentBadge returns the agent type badge
func AgentBadge() string {
	if !IsTTY {
		return "[AGENT]"
	}
	return baseBadge.Background(Magenta).Foreground(White).Render("◈ AGENT")
}

// PluginBadge returns the plugin type badge
func PluginBadge() string {
	if !IsTTY {
		return "[PLUGIN]"
	}
	return baseBadge.Background(Gold).Foreground(Black).Render("⬡ PLUGIN")
}

// Status badge functions

// StatusOK returns the success status badge
func StatusOK() string {
	if !IsTTY {
		return "[OK]"
	}
	return baseBadge.Background(Green).Foreground(White).Render("✓")
}

// StatusWarn returns the warning status badge
func StatusWarn() string {
	if !IsTTY {
		return "[!]"
	}
	return baseBadge.Background(Copper).Foreground(White).Render("!")
}

// StatusError returns the error status badge
func StatusError() string {
	if !IsTTY {
		return "[ERR]"
	}
	return baseBadge.Background(Pink).Foreground(White).Render("✗")
}

// StatusNew returns the new status badge
func StatusNew() string {
	if !IsTTY {
		return "[NEW]"
	}
	return baseBadge.Background(Cyan).Foreground(White).Render("NEW")
}

// StatusUpdate returns the update status badge
func StatusUpdate() string {
	if !IsTTY {
		return "[UPD]"
	}
	return baseBadge.Background(Gold).Foreground(Black).Render("UPD")
}

// ═══════════════════════════════════════════════════════════════════════════════
// LOGO - The centerpiece
// ═══════════════════════════════════════════════════════════════════════════════

// Logo returns the stunning Tome logo
func Logo() string {
	// Plain output for non-TTY environments
	if !IsTTY {
		return "\n  TOME - Your Grimoire of AI Skills\n"
	}

	// Book/tome ASCII art with gradient coloring
	lines := []struct {
		text  string
		color lipgloss.Color
	}{
		{"", Black},
		{"        ▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄", DarkGray},
		{"       ██▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀██", Bronze},
		{"       ██                            ██", Bronze},
		{"       ██  ▀█▀  ▄▀▀▄  █▄ ▄█  █▀▀     ██", Gold},
		{"       ██   █   █  █  █ ▀ █  █▀▀     ██", Amber},
		{"       ██   █   ▀▄▄▀  █   █  ▀▀▀     ██", Copper},
		{"       ██                            ██", Bronze},
		{"       ██       ─────────────        ██", DarkGray},
		{"       ██        ✦ Your Grimoire     ██", Purple},
		{"       ██                            ██", Bronze},
		{"       ██▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄██", Bronze},
		{"        ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀", DarkGray},
		{"", Black},
	}

	var result strings.Builder
	for _, line := range lines {
		styled := lipgloss.NewStyle().Foreground(line.color).Render(line.text)
		result.WriteString(styled)
		result.WriteString("\n")
	}

	return result.String()
}

// LogoCompact returns a smaller logo for headers
func LogoCompact() string {
	if !IsTTY {
		return "TOME"
	}
	book := lipgloss.NewStyle().Foreground(Bronze).Render("📖")
	name := lipgloss.NewStyle().Foreground(Gold).Bold(true).Render("TOME")
	return fmt.Sprintf(" %s %s ", book, name)
}

// ═══════════════════════════════════════════════════════════════════════════════
// DECORATIVE ELEMENTS
// ═══════════════════════════════════════════════════════════════════════════════

// Divider returns a horizontal divider
func Divider(width int) string {
	return lipgloss.NewStyle().
		Foreground(DarkGray).
		Render(strings.Repeat("─", width))
}

// DoubleDivider returns a fancy double-line divider
func DoubleDivider(width int) string {
	return lipgloss.NewStyle().
		Foreground(Bronze).
		Render(strings.Repeat("═", width))
}

// Flourish returns a decorative flourish
func Flourish() string {
	return lipgloss.NewStyle().
		Foreground(Gold).
		Render("  ─── ✦ ───  ")
}

// SectionHeader creates a decorated section header
func SectionHeader(title string, _ int) string {
	// Plain output for non-TTY environments
	if !IsTTY {
		return fmt.Sprintf("=== %s ===", title)
	}

	// Use terminal width, capped at 80
	width := TerminalWidth()
	if width > 80 {
		width = 80
	}

	titleStyled := lipgloss.NewStyle().
		Foreground(Gold).
		Bold(true).
		Render(title)

	titleLen := lipgloss.Width(title)
	padLeft := (width - titleLen - 6) / 2
	padRight := width - titleLen - 6 - padLeft

	left := lipgloss.NewStyle().Foreground(DarkGray).Render(strings.Repeat("─", padLeft) + "┤ ")
	right := lipgloss.NewStyle().Foreground(DarkGray).Render(" ├" + strings.Repeat("─", padRight))

	return left + titleStyled + right
}

// ═══════════════════════════════════════════════════════════════════════════════
// CARDS - For displaying artifacts
// ═══════════════════════════════════════════════════════════════════════════════

// ArtifactCard creates a beautiful card for an artifact
func ArtifactCard(badge, name, description string, width int) string {
	nameStyled := lipgloss.NewStyle().
		Foreground(White).
		Bold(true).
		Render(name)

	// Truncate description if needed
	maxDescLen := width - 6
	if len(description) > maxDescLen {
		description = description[:maxDescLen-3] + "..."
	}

	descStyled := lipgloss.NewStyle().
		Foreground(Gray).
		Render(description)

	content := fmt.Sprintf("%s  %s\n   %s", badge, nameStyled, descStyled)

	return content
}

// ═══════════════════════════════════════════════════════════════════════════════
// TABLES - For structured data
// ═══════════════════════════════════════════════════════════════════════════════

// TableHeader creates a styled table header
func TableHeader(columns ...string) string {
	var cells []string
	for _, col := range columns {
		cells = append(cells, lipgloss.NewStyle().
			Foreground(Gold).
			Bold(true).
			Render(col))
	}
	return strings.Join(cells, "  ")
}

// TableRow creates a styled table row
func TableRow(columns ...string) string {
	var cells []string
	for i, col := range columns {
		style := lipgloss.NewStyle().Foreground(White)
		if i > 0 {
			style = style.Foreground(Gray)
		}
		cells = append(cells, style.Render(col))
	}
	return strings.Join(cells, "  ")
}

// ═══════════════════════════════════════════════════════════════════════════════
// PROGRESS INDICATORS
// ═══════════════════════════════════════════════════════════════════════════════

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner returns a spinner frame (for animated use)
func Spinner(frame int) string {
	return lipgloss.NewStyle().
		Foreground(Purple).
		Render(spinnerFrames[frame%len(spinnerFrames)])
}

// StaticSpinner returns a static loading indicator
func StaticSpinner() string {
	return lipgloss.NewStyle().
		Foreground(Purple).
		Render("◐")
}

// ProgressDots returns animated-style dots
func ProgressDots(count int) string {
	dots := strings.Repeat("●", count) + strings.Repeat("○", 3-count)
	return lipgloss.NewStyle().
		Foreground(Purple).
		Render(dots)
}

// ═══════════════════════════════════════════════════════════════════════════════
// STATUS LINE COMPONENTS
// ═══════════════════════════════════════════════════════════════════════════════

// StatusLine creates a status line with icon and message
func StatusLine(icon, message string, color lipgloss.Color) string {
	if !IsTTY {
		return fmt.Sprintf("  %s %s", icon, message)
	}
	iconStyled := lipgloss.NewStyle().Foreground(color).Render(icon)
	msgStyled := lipgloss.NewStyle().Foreground(color).Render(message)
	return fmt.Sprintf("  %s %s", iconStyled, msgStyled)
}

// SuccessLine creates a success status line
func SuccessLine(message string) string {
	if !IsTTY {
		return fmt.Sprintf("  OK: %s", message)
	}
	return StatusLine("✓", message, Green)
}

// ErrorLine creates an error status line
func ErrorLine(message string) string {
	if !IsTTY {
		return fmt.Sprintf("  ERROR: %s", message)
	}
	return StatusLine("✗", message, Pink)
}

// WarningLine creates a warning status line
func WarningLine(message string) string {
	if !IsTTY {
		return fmt.Sprintf("  WARN: %s", message)
	}
	return StatusLine("!", message, Copper)
}

// InfoLine creates an info status line
func InfoLine(message string) string {
	if !IsTTY {
		return fmt.Sprintf("  %s", message)
	}
	return StatusLine("→", message, Blue)
}

// ═══════════════════════════════════════════════════════════════════════════════
// EMPTY STATES
// ═══════════════════════════════════════════════════════════════════════════════

// EmptyTome returns a friendly empty state
func EmptyTome() string {
	if !IsTTY {
		return "\n  (empty)\n\n  Your tome awaits its first inscription...\n  Use `tome learn <source>` to begin.\n"
	}

	book := lipgloss.NewStyle().Foreground(DarkGray).Render(`
      ┌─────────────┐
      │             │
      │   (empty)   │
      │             │
      └─────────────┘`)

	message := lipgloss.NewStyle().Foreground(Gray).Render("Your tome awaits its first inscription...")
	hint := lipgloss.NewStyle().Foreground(Cyan).Render("tome learn <source>")

	return fmt.Sprintf("%s\n\n  %s\n  Use %s to begin.\n", book, message, hint)
}

// NoResults returns a friendly no-results state
func NoResults(query string) string {
	if !IsTTY {
		return fmt.Sprintf("\n  No artifacts found for \"%s\"\n  Try broader search terms\n", query)
	}

	crystal := lipgloss.NewStyle().Foreground(DarkGray).Render("🔮")
	message := lipgloss.NewStyle().Foreground(Gray).Render(fmt.Sprintf("No artifacts found for \"%s\"", query))
	hint := lipgloss.NewStyle().Foreground(Cyan).Render("Try broader search terms")

	return fmt.Sprintf("\n  %s %s\n  %s\n", crystal, message, hint)
}

// ═══════════════════════════════════════════════════════════════════════════════
// HELPER FUNCTIONS
// ═══════════════════════════════════════════════════════════════════════════════

// Pad adds padding to text
func Pad(text string, left int) string {
	return strings.Repeat(" ", left) + text
}

// Truncate truncates text to max length with ellipsis
func Truncate(text string, max int) string {
	if len(text) <= max {
		return text
	}
	return text[:max-3] + "..."
}

// WrapText wraps text to fit within maxWidth, returning multiple lines.
// Each line is indented with the given prefix (typically spaces).
func WrapText(text string, maxWidth int, indent string) []string {
	if maxWidth <= 0 {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		if currentLine.Len() == 0 {
			currentLine.WriteString(word)
		} else if currentLine.Len()+1+len(word) <= maxWidth {
			currentLine.WriteString(" ")
			currentLine.WriteString(word)
		} else {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(word)
		}
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}

// Render applies a lipgloss style to text, returning plain text in non-TTY environments.
// Use this wrapper when you want TTY-aware styling.
func Render(style lipgloss.Style, text string) string {
	if !IsTTY {
		return text
	}
	return style.Render(text)
}

// Convenience functions for TTY-aware rendering of common styles

// RenderMuted renders text in muted style (TTY-aware)
func RenderMuted(text string) string {
	return Render(Muted, text)
}

// RenderDim renders text in dim style (TTY-aware)
func RenderDim(text string) string {
	return Render(Dim, text)
}

// RenderHighlight renders text in highlight style (TTY-aware)
func RenderHighlight(text string) string {
	return Render(Highlight, text)
}

// RenderSuccess renders text in success style (TTY-aware)
func RenderSuccess(text string) string {
	return Render(Success, text)
}

// RenderError renders text in error style (TTY-aware)
func RenderError(text string) string {
	return Render(Error, text)
}

// RenderWarning renders text in warning style (TTY-aware)
func RenderWarning(text string) string {
	return Render(Warning, text)
}

// RenderInfo renders text in info style (TTY-aware)
func RenderInfo(text string) string {
	return Render(Info, text)
}

// TerminalWidth returns the current terminal width, defaulting to 80 if unknown
func TerminalWidth() int {
	w, _, err := term.GetSize(os.Stdout.Fd())
	if err != nil || w <= 0 {
		return 80
	}
	return w
}

// DescriptionWidth returns the recommended width for descriptions based on terminal size
func DescriptionWidth() int {
	w := TerminalWidth()
	// Account for indentation (4 chars) and some margin
	desc := w - 8
	if desc < 40 {
		return 40
	}
	return desc
}

// Box wraps content in a styled box
func Box(content string, width int, borderColor lipgloss.Color) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width).
		Padding(0, 1).
		Render(content)
}

// GradientText applies a gradient-like effect to text (alternating colors)
func GradientText(text string, colors ...lipgloss.Color) string {
	if len(colors) == 0 {
		return text
	}

	var result strings.Builder
	runes := []rune(text)
	for i, r := range runes {
		color := colors[i%len(colors)]
		styled := lipgloss.NewStyle().Foreground(color).Render(string(r))
		result.WriteString(styled)
	}
	return result.String()
}

// ═══════════════════════════════════════════════════════════════════════════════
// PAGE TEMPLATES
// ═══════════════════════════════════════════════════════════════════════════════

// PageHeader creates a consistent page header
func PageHeader(title string) string {
	icon := lipgloss.NewStyle().Foreground(Gold).Render("📜")
	titleStyled := lipgloss.NewStyle().Foreground(Gold).Bold(true).Render(title)
	return fmt.Sprintf("\n  %s %s\n", icon, titleStyled)
}

// PageFooter creates a consistent page footer matching the header width
func PageFooter() string {
	// Plain output for non-TTY environments
	if !IsTTY {
		return "\n"
	}

	width := TerminalWidth()
	if width > 80 {
		width = 80
	}
	padSide := (width - 5) / 2 // 5 = " ✦ " with spaces
	left := strings.Repeat("─", padSide)
	right := strings.Repeat("─", width-padSide-5)
	line := lipgloss.NewStyle().Foreground(DarkGray).Render(left + " ✦ " + right)
	return "\n" + line + "\n"
}
