# Tome

**Your AI coding agents just got superpowers.**

Tome is the skill manager for AI coding agents. Install prompts, commands, and workflows from across GitHub in seconds. Build collections. Share knowledge. Make your AI assistant actually know what you need it to know.

Stop copy-pasting prompts. Start wielding a grimoire of battle-tested skills.

## Why Tome?

Your AI coding agent is brilliant, but it doesn't know your stack, your patterns, or your team's conventions. Until now, you've been pasting the same context into every conversation. There's a better way.

**Tome lets you:**
- **Install skills instantly** from any GitHub repo with one command
- **Discover proven prompts** from the community with powerful search
- **Keep everything synced** as collections evolve and improve
- **Build your own grimoire** of team knowledge and best practices
- **Support every agent** - Claude Code, Cursor, Windsurf, and more

Think of it as package management for AI agent knowledge. `npm install` for prompts.

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

Install your first skill collection:

```bash
# Install Yegge's famous programming tips as agent prompts
tome learn kennyg/yegges-tips

# Search for React best practices
tome seek "react patterns"

# See what you've learned
tome index

# Keep everything up to date
tome renew
```

That's it. Your AI agent now has access to curated, versioned knowledge.

## Commands

Tome uses a grimoire-inspired command vocabulary (with practical aliases):

### Learn New Skills

```bash
tome learn owner/repo           # Install from GitHub
tome learn owner/repo@branch    # Install specific branch
tome learn owner/repo --path custom/location
```

*Aliases: `inscribe`, `add`, `install`*

### Browse Your Collection

```bash
tome index                      # List all installed artifacts
tome index --agent claude       # Filter by agent
```

*Aliases: `list`, `ls`*

### Search for Skills

```bash
tome seek "typescript testing"  # Find skills on GitHub
tome seek cursor --stars 100    # Filter by popularity
```

*Aliases: `search`, `find`*

### Preview Before Installing

```bash
tome peek owner/repo            # See what's in a collection
tome peek owner/repo:path       # Preview a specific skill
```

*Aliases: `preview`, `inspect`*

### Inspect Details

```bash
tome study owner/repo           # View artifact information
```

*Aliases: `info`, `examine`*

### Remove Skills

```bash
tome forget owner/repo          # Uninstall artifacts
```

*Aliases: `remove`, `rm`*

### Update Everything

```bash
tome renew                      # Sync all installed skills
tome renew owner/repo           # Update specific collection
```

*Aliases: `sync`, `update`*

### Create Your Own Collection

```bash
tome conjure my-team-skills     # Initialize new collection
cd my-team-skills
# Add your .md files and prompts
tome bind                       # Validate the collection
```

*Conjure aliases: `init`*
*Bind aliases: `build`, `validate`*

## Creating Collections

Sharing knowledge with your team (or the world) is simple:

1. **Initialize a collection:**
   ```bash
   tome conjure my-awesome-skills
   cd my-awesome-skills
   ```

2. **Add your skills:**
   Create markdown files with prompts, commands, or documentation. Organize them however you like.

3. **Define the manifest:**
   Edit `tome.yaml` to describe your collection:
   ```yaml
   name: my-awesome-skills
   description: Our team's AI agent skills
   author: your-name
   agents:
     - claude
     - cursor
   ```

4. **Validate:**
   ```bash
   tome bind
   ```

5. **Share:**
   Push to GitHub and anyone can install with `tome learn yourusername/my-awesome-skills`

## Example Collections

- [kennyg/yegges-tips](https://github.com/kennyg/yegges-tips) - Steve Yegge's programming wisdom as agent prompts

## How It Works

Tome manages a library of "artifacts" - prompts, commands, and knowledge - stored in your `~/.tome` directory (or custom locations). Each artifact is linked to a GitHub repository, making it easy to version, share, and update collective intelligence.

When you `tome learn` a collection, Tome:
1. Fetches the repo and reads its `tome.yaml` manifest
2. Installs artifacts to the appropriate locations for each AI agent
3. Tracks metadata so you can update, remove, or inspect later

Your AI agents automatically pick up these new skills in their context.

## Agent Support

Tome works with all major AI coding agents:
- **Claude Code** - Anthropic's official CLI
- **Cursor** - The AI-first code editor
- **Windsurf** - Codeium's agent environment
- And more...

Each agent has its own conventions for prompts and commands. Tome handles the details.

## Philosophy

**Knowledge should be shared, versioned, and reusable.**

Every team develops patterns. Every developer discovers tricks. Every project has domain context that would make AI agents more useful. But this knowledge lives in Slack messages, forgotten docs, and copy-paste buffers.

Tome makes it first-class. Package it. Version it. Share it. Install it with one command.

The grimoire awaits.

---

**Built with Go.** Inspired by every package manager that made development better.

## License

MIT
