# AGENTS.md - Tome

AI agent guide for contributing to Tome, the AI Agent Skill Manager.

## Project Overview

Tome is a CLI tool for discovering, installing, and managing AI agent capabilities (skills, commands, prompts, hooks) from GitHub repositories. Think of it as a "spellbook" for your AI agent.

**Tech Stack:**
- **Language**: Go 1.23+
- **CLI Framework**: Cobra
- **TUI**: Charm (Bubble Tea, Lip Gloss)
- **Task Runner**: mise
- **Git Hooks**: hk
- **Issue Tracking**: bd (beads)

## Core Principles

### dot-config Specification

This project follows the [dot-config specification](https://dot-config.github.io/) for configuration file locations. This keeps project roots clean and follows XDG conventions.

**Config locations:**
- **User config**: `~/.config/tome/` (or `$XDG_CONFIG_HOME/tome/`)
- **Project config**: `.config/tome/` (in project root)

When adding new configuration, always place it under `.config/tome/` - never scatter config files in the project root.

## Project Structure

```
tome/
├── main.go                 # Entry point
├── go.mod / go.sum         # Go dependencies
├── mise.toml               # Tool versions and tasks
├── hk.pkl                  # Git hook configuration
├── cmd/                    # CLI commands
│   ├── root.go             # Main command + logo
│   ├── learn.go            # Install artifacts
│   ├── list.go             # Show installed
│   ├── search.go           # Search GitHub
│   ├── remove.go           # Uninstall
│   ├── sync.go             # Update all
│   └── info.go             # Artifact details
├── internal/
│   ├── ui/                 # Lip Gloss styles
│   ├── artifact/           # Types (Skill, Command, Prompt, Hook)
│   ├── source/             # GitHub/URL/local source parsing
│   └── config/             # Paths and state management
├── .config/tome/           # Project-level config (dot-config spec)
└── .beads/                 # Issue tracking database

# User config lives at: ~/.config/tome/
```

## Development Commands

```bash
# Build
mise run build              # Build to bin/tome

# Run
./bin/tome --help           # See all commands
mise run dev learn foo/bar  # Run in dev mode

# Test
mise run test               # Run all tests

# Format/Lint
mise run fmt                # Format code
mise run lint               # Run linter (requires golangci-lint)

# Install locally
mise run install            # Install to GOBIN
```

## Releasing

**Always use mise for releases:**

```bash
mise run release <version>  # e.g., mise run release 0.2.0
```

This will:
1. Create and push the git tag
2. Trigger the GitHub Actions release workflow
3. Build binaries for all platforms via GoReleaser

**Do NOT manually tag releases** - always use `mise run release`.

## Issue Tracking with bd (beads)

**IMPORTANT**: This project uses **bd (beads)** for ALL issue tracking. Do NOT use markdown TODOs, task lists, or other tracking methods.

### Why bd?

- Dependency-aware: Track blockers and relationships between issues
- Git-friendly: Auto-syncs to JSONL for version control
- Agent-optimized: JSON output, ready work detection, discovered-from links
- Prevents duplicate tracking systems and confusion

### Quick Start

**Check for ready work:**
```bash
bd ready
```

**Create new issues:**
```bash
bd create "Issue title" -t bug|feature|task -p 0-4
bd create "Issue title" -p 1 --deps discovered-from:tome-xxx
```

**Claim and update:**
```bash
bd update tome-xxx --status in_progress
bd update tome-xxx --priority 1
```

**Complete work:**
```bash
bd close tome-xxx --reason "Completed"
```

### Issue Types

- `bug` - Something broken
- `feature` - New functionality
- `task` - Work item (tests, docs, refactoring)
- `epic` - Large feature with subtasks
- `chore` - Maintenance (dependencies, tooling)

### Priorities

- `0` - Critical (security, data loss, broken builds)
- `1` - High (major features, important bugs)
- `2` - Medium (default, nice-to-have)
- `3` - Low (polish, optimization)
- `4` - Backlog (future ideas)

### Workflow for AI Agents

1. **Check ready work**: `bd ready` shows unblocked issues
2. **Claim your task**: `bd update <id> --status in_progress`
3. **Work on it**: Implement, test, document
4. **Discover new work?** Create linked issue:
   - `bd create "Found bug" -p 1 --deps discovered-from:<parent-id>`
5. **Complete**: `bd close <id> --reason "Done"`
6. **Commit together**: Always commit `.beads/` changes with code changes

### CLI Help

Run `bd <command> --help` to see all available flags for any command.

## Coding Guidelines

### Adding New Commands

1. Create `cmd/<name>.go`
2. Define the cobra.Command
3. Add to `rootCmd` in `cmd/root.go`
4. Use `internal/ui` styles for output

### Style Guide

- Use Lip Gloss styles from `internal/ui/styles.go` for all output
- Keep commands focused - delegate to internal packages
- Return errors up, handle display at command level
- Use `exitWithError()` for fatal errors

### Testing

- Write tests for internal packages
- Use table-driven tests
- Test source parsing edge cases

## Important Rules

- Use bd for ALL task tracking
- Link discovered work with `discovered-from` dependencies
- Check `bd ready` before asking "what should I work on?"
- Run `bd <cmd> --help` to discover available flags
- Do NOT create markdown TODO lists
- Do NOT use external issue trackers
