# Tome

**Write once, use everywhere.**

Tome is a format converter and skill manager for AI coding agents. Write your prompts, skills, and MCP configs once—then convert them to work with Claude Code, OpenCode, GitHub Copilot, Cursor, and more.

Stop maintaining parallel config files. Start using one source of truth.

## The Problem

You use multiple AI coding agents. Each has its own format:

| Agent | Skills | Commands | MCP Config |
|-------|--------|----------|------------|
| Claude Code | `.claude/skills/*/SKILL.md` | `.claude/commands/*.md` | `.mcp.json` |
| OpenCode | `.opencode/skill/*/SKILL.md` | `.opencode/commands/*.md` | `opencode.json` |
| Copilot | `agents/*.agent.md` | `.github/prompts/*.prompt.md` | — |
| Cursor | `.cursor/rules/*.md` | — | `.cursor/mcp.json` |

Maintaining the same knowledge across all of them? That's busywork.

## The Solution

```bash
# Convert a Copilot agent to Claude format
tome transmogrify agents/CSharp.agent.md --to claude

# Convert an entire directory
tome transmogrify ./copilot-skills/ --to opencode --output ./converted/

# Convert MCP configs between formats
tome transmogrify .mcp.json --to opencode
tome transmogrify opencode.json --to claude

# Preview before converting
tome transmogrify ./skills/ --to cursor --dry-run
```

## Installation

### Homebrew (macOS/Linux)

```bash
brew install kennyg/tap/tome
```

### Go Install

```bash
go install github.com/kennyg/tome@latest
```

## Quick Start

### Converting Between Formats

```bash
# You have Copilot skills, want them in Claude
tome transmogrify agents/ --to claude

# You have Claude skills, want them in Cursor
tome transmogrify .claude/skills/ --to cursor

# Convert a single file
tome transmogrify my-skill.agent.md --to opencode
```

### Installing Skills from GitHub

```bash
# Install a skill collection
tome learn kennyg/yegges-tips

# See what you've installed
tome index

# Keep everything updated
tome renew
```

### Discovering Skills

```bash
# Search GitHub for skills
tome seek "react patterns"

# Preview before installing
tome peek owner/repo
```

## Commands

### Transmogrify (Convert)

The core feature. Convert skills, commands, and MCP configs between agent formats.

```bash
tome transmogrify <source> --to <format>
```

**Supported formats:** `claude`, `opencode`, `copilot`, `cursor`

**Supported artifact types:**
- Skills (SKILL.md, .agent.md, rules)
- Commands (.prompt.md, commands/*.md)
- MCP configs (.mcp.json, opencode.json)

*Aliases: `convert`, `morph`, `xform`*

### Learn (Install)

```bash
tome learn owner/repo           # Install from GitHub
tome learn owner/repo@branch    # Specific branch
```

*Aliases: `inscribe`, `add`, `install`*

### Index (List)

```bash
tome index                      # List all installed artifacts
tome index --agent claude       # Filter by agent
```

*Aliases: `list`, `ls`*

### Seek (Search)

```bash
tome seek "typescript testing"  # Find skills on GitHub
```

*Aliases: `search`, `find`*

### Renew (Update)

```bash
tome renew                      # Sync all installed skills
```

*Aliases: `sync`, `update`*

### Forget (Remove)

```bash
tome forget owner/repo          # Uninstall artifacts
```

*Aliases: `remove`, `rm`*

## Agent Support

| Agent | Skills | Commands | MCP | Status |
|-------|--------|----------|-----|--------|
| Claude Code | ✅ | ✅ | ✅ | Full support |
| OpenCode | ✅ | ✅ | ✅ | Full support |
| GitHub Copilot | ✅ | ✅ | — | Full support |
| Cursor | ✅ | — | ✅ | Full support |
| Windsurf | ✅ | — | — | Partial |

## GitHub Token (Optional)

For higher rate limits (5000 vs 60 requests/hour) or private repos:

```bash
export GITHUB_TOKEN=ghp_your_token_here
```

Tome auto-discovers tokens from `GITHUB_TOKEN`, `GH_TOKEN`, or gh CLI config.

## Philosophy

AI coding agents are converging on similar concepts—skills, prompts, MCP—but with different file formats. That fragmentation is temporary friction, not fundamental.

Tome bridges the gap. Write your knowledge once, convert it everywhere.

---

**Built with Go.** MIT License.
