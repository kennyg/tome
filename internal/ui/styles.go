package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color palette - ancient tome/parchment theme
var (
	// Primary colors - aged parchment and ink
	Gold      = lipgloss.Color("#D4A574") // Aged gold/bronze
	Amber     = lipgloss.Color("#E8B866") // Warm amber
	Sepia     = lipgloss.Color("#A67B5B") // Sepia brown
	Ink       = lipgloss.Color("#2C1810") // Dark ink
	Parchment = lipgloss.Color("#F5E6D3") // Light parchment

	// Accent colors - magical elements
	Purple    = lipgloss.Color("#8B5CF6") // Mystical purple
	Pink      = lipgloss.Color("#EC4899") // Enchanted pink
	Blue      = lipgloss.Color("#60A5FA") // Arcane blue
	Green     = lipgloss.Color("#34D399") // Nature/success
	Yellow    = lipgloss.Color("#FBBF24") // Highlight gold
	Orange    = lipgloss.Color("#F97316") // Warning flame
	Gray      = lipgloss.Color("#9CA3AF")
	DarkGray  = lipgloss.Color("#6B7280")
	LightGray = lipgloss.Color("#D1D5DB")
)

// Text styles
var (
	// Title is the main title style - golden header
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(Gold).
		MarginBottom(1)

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
		Foreground(Orange)

	// Info messages
	Info = lipgloss.NewStyle().
		Foreground(Blue)

	// Muted/secondary text
	Muted = lipgloss.NewStyle().
		Foreground(Gray)

	// Highlight for important items - golden emphasis
	Highlight = lipgloss.NewStyle().
			Foreground(Yellow).
			Bold(true)
)

// Component styles
var (
	// Box for bordered containers - scroll-like appearance
	Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Sepia).
		Padding(1, 2)

	// ListItem for list entries
	ListItem = lipgloss.NewStyle().
			PaddingLeft(2)

	// SelectedItem for highlighted list items
	SelectedItem = lipgloss.NewStyle().
			Foreground(Gold).
			Bold(true).
			PaddingLeft(2)

	// Badge for type indicators
	Badge = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(Purple).
		Padding(0, 1).
		MarginRight(1)
)

// Artifact type badges with book-themed styling
var (
	SkillBadge  = Badge.Background(Purple).Render("✦ SKILL")
	CmdBadge    = Badge.Background(Blue).Render("⌘ CMD")
	PromptBadge = Badge.Background(Green).Render("✎ PROMPT")
	HookBadge   = Badge.Background(Orange).Render("⚡ HOOK")
)

// Logo returns the styled Tome logo - ancient book aesthetic
func Logo() string {
	// Book/tome ASCII art
	logo := `
       ┌─────────────────────┐
       │  ═══════════════    │
       │    ╔╦╗╔═╗╔╦╗╔═╗     │
       │     ║ ║ ║║║║║╣      │
       │     ╩ ╚═╝╩ ╩╚═╝     │
       │  ═══════════════    │
       │    ✦ Your Grimoire  │
       └─────────────────────┘`

	return lipgloss.NewStyle().
		Foreground(Gold).
		Bold(true).
		Render(logo)
}

// ScrollTop returns the top of a scroll decoration
func ScrollTop(width int) string {
	if width < 4 {
		width = 4
	}
	inner := strings.Repeat("═", width-2)
	return lipgloss.NewStyle().
		Foreground(Sepia).
		Render("╔" + inner + "╗")
}

// ScrollBottom returns the bottom of a scroll decoration
func ScrollBottom(width int) string {
	if width < 4 {
		width = 4
	}
	inner := strings.Repeat("═", width-2)
	return lipgloss.NewStyle().
		Foreground(Sepia).
		Render("╚" + inner + "╝")
}

// Divider returns a horizontal divider - scroll flourish style
func Divider(width int) string {
	return lipgloss.NewStyle().
		Foreground(Sepia).
		Render("  " + strings.Repeat("─", width-4) + "  ")
}

// ChapterDivider returns a fancy chapter break
func ChapterDivider(width int) string {
	side := (width - 5) / 2
	return lipgloss.NewStyle().
		Foreground(Gold).
		Render(strings.Repeat("─", side) + "  ◆  " + strings.Repeat("─", side))
}

// PageHeader returns a decorative page header
func PageHeader(title string) string {
	styled := lipgloss.NewStyle().
		Foreground(Gold).
		Bold(true).
		Render(title)
	return "  📜 " + styled
}
