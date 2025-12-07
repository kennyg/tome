# Tome - AI Agent Skill Manager

A CLI tool to customize your AI agent by installing skills, commands, and prompts from GitHub repositories.

## Inspiration

- **openskills** (github.com/numman-ali/openskills) - existing tool but only handles skills, not commands/prompts
- **RPG spellbooks** - WoW grimoires, collecting abilities and artifacts
- Gamifying the process of discovering and collecting AI agent capabilities

## Core Concept

A "spellbook" for your AI agent where you collect abilities:

```bash
tome search memory           # find in registry
tome learn bd-tracking       # install a skill
tome cast /deploy            # run a command
tome list                    # your collection
tome sync                    # update everything
```

## Artifact Types

1. **Skills** (passive abilities) - agent learns new capabilities
   - Location: `~/.claude/skills/` or `.claude/skills/`
   - Format: `SKILL.md` with YAML frontmatter

2. **Commands** (active spells) - slash commands you invoke
   - Location: `~/.claude/commands/` or `.claude/commands/`
   - Format: markdown files with prompts

3. **Prompts** (incantations) - reusable prompt templates
   - TBD on location/format

4. **Hooks** (enchantments) - event triggers
   - Location: `.claude/settings.json` hooks section

## Key Features

- Install from any GitHub repo (not just a central registry)
- Support direct URLs to individual files
- Support subpaths within repos (e.g., `owner/repo:path/to/skill`)
- Interactive selection like openskills
- Multi-agent support (Claude Code, Cursor, Windsurf, Aider)

## CLI Commands

```bash
tome search <query>          # search registry/github
tome learn <source>          # install skill
tome cast <command>          # run a command (or just use /command directly?)
tome list                    # show installed artifacts
tome list --skills           # filter by type
tome list --commands
tome sync                    # update all installed artifacts
tome remove <name>           # uninstall
tome info <name>             # show details about an artifact
```

## Install Source Formats

```bash
# GitHub repo (scans for all artifacts)
tome learn owner/repo

# GitHub repo with subpath
tome learn owner/repo:examples/my-skill

# Direct URL
tome learn https://raw.githubusercontent.com/.../SKILL.md

# Local file/directory
tome learn ./my-local-skill
```

## RPG Flavor Ideas (optional/future)

- Rarity tiers: common, rare, legendary
- Usage stats: "You've cast /deploy 47 times"
- Achievements: "Learned 10 skills"
- Fun messages: "Your tome grows stronger..."

## Tech Stack Options

- TypeScript (like openskills) - good npm ecosystem
- Go - single binary, fast
- Rust - single binary, very fast

## Prior Art

- openskills - https://github.com/numman-ali/openskills
- Claude Code skills spec from Anthropic
- beads skill example - https://github.com/steveyegge/beads/tree/main/examples/claude-code-skill

## Open Questions

- Should we maintain a central registry, or just rely on GitHub search?
- How to handle versioning/updates?
- Should `tome cast` actually run commands, or is that redundant with `/command`?
- Support for other agents beyond Claude Code?
