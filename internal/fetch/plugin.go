package fetch

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kennyg/tome/internal/artifact"
)

// IsPlugin checks if a GitHub repo/path is a plugin by looking for .claude-plugin/plugin.json
func (c *Client) IsPlugin(apiURL string) bool {
	// Check for .claude-plugin directory
	contents, err := c.ListGitHubContents(apiURL)
	if err != nil {
		return false
	}

	for _, item := range contents {
		if item.Type == "dir" && item.Name == ".claude-plugin" {
			// Check for plugin.json inside
			pluginDirURL := appendPath(apiURL, ".claude-plugin")
			pluginContents, err := c.ListGitHubContents(pluginDirURL)
			if err != nil {
				return false
			}
			for _, pItem := range pluginContents {
				if pItem.Type == "file" && pItem.Name == "plugin.json" {
					return true
				}
			}
		}
	}
	return false
}

// FetchPlugin fetches and parses a complete plugin from a GitHub repo
func (c *Client) FetchPlugin(apiURL string, source string) (*artifact.Plugin, error) {
	plugin := &artifact.Plugin{
		Source: source,
	}

	// Fetch plugin.json manifest
	manifestURL := appendPath(apiURL, ".claude-plugin")
	manifestContents, err := c.ListGitHubContents(manifestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to list .claude-plugin contents: %w", err)
	}

	var manifestDownloadURL string
	for _, item := range manifestContents {
		if item.Type == "file" && item.Name == "plugin.json" {
			manifestDownloadURL = item.DownloadURL
			break
		}
	}

	if manifestDownloadURL == "" {
		return nil, fmt.Errorf("plugin.json not found in .claude-plugin/")
	}

	manifestContent, err := c.FetchURL(manifestDownloadURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch plugin.json: %w", err)
	}

	if err := json.Unmarshal(manifestContent, &plugin.Manifest); err != nil {
		return nil, fmt.Errorf("failed to parse plugin.json: %w", err)
	}

	// List root contents to find artifact directories
	contents, err := c.ListGitHubContents(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugin contents: %w", err)
	}

	// Process each artifact directory
	for _, item := range contents {
		if item.Type != "dir" {
			continue
		}

		switch item.Name {
		case "skills":
			skills, err := c.fetchPluginSkills(apiURL)
			if err == nil {
				plugin.Skills = skills
			}
		case "commands":
			commands, err := c.fetchPluginCommands(apiURL)
			if err == nil {
				plugin.Commands = commands
			}
		case "agents":
			agents, err := c.fetchPluginAgents(apiURL)
			if err == nil {
				plugin.Agents = agents
			}
		case "hooks":
			hooks, err := c.fetchPluginHooks(apiURL)
			if err == nil {
				plugin.Hooks = hooks
			}
		}
	}

	return plugin, nil
}

// fetchPluginSkills fetches all skills from a plugin's skills/ directory
func (c *Client) fetchPluginSkills(apiURL string) ([]artifact.Artifact, error) {
	var skills []artifact.Artifact

	skillsURL := appendPath(apiURL, "skills")
	contents, err := c.ListGitHubContents(skillsURL)
	if err != nil {
		return nil, err
	}

	for _, item := range contents {
		if item.Type == "dir" {
			// Check for SKILL.md in subdirectory
			skillDirURL := appendPath(skillsURL, item.Name)
			skillContents, err := c.ListGitHubContents(skillDirURL)
			if err != nil {
				continue
			}

			for _, skillFile := range skillContents {
				if skillFile.Type == "file" && strings.ToUpper(skillFile.Name) == "SKILL.MD" {
					content, err := c.FetchURL(skillFile.DownloadURL)
					if err != nil {
						continue
					}

					art, err := ParseSkill(content, skillFile.DownloadURL)
					if err != nil {
						continue
					}

					skills = append(skills, *art)
				}
			}
		} else if item.Type == "file" && strings.ToUpper(item.Name) == "SKILL.MD" {
			// Flat skill at skills/SKILL.md
			content, err := c.FetchURL(item.DownloadURL)
			if err != nil {
				continue
			}

			art, err := ParseSkill(content, item.DownloadURL)
			if err != nil {
				continue
			}

			skills = append(skills, *art)
		}
	}

	return skills, nil
}

// fetchPluginCommands fetches all commands from a plugin's commands/ directory
func (c *Client) fetchPluginCommands(apiURL string) ([]artifact.Artifact, error) {
	var commands []artifact.Artifact

	commandsURL := appendPath(apiURL, "commands")
	contents, err := c.ListGitHubContents(commandsURL)
	if err != nil {
		return nil, err
	}

	for _, item := range contents {
		if item.Type == "file" && strings.HasSuffix(strings.ToLower(item.Name), ".md") {
			content, err := c.FetchURL(item.DownloadURL)
			if err != nil {
				continue
			}

			art, err := ParseCommand(content, item.Name, item.DownloadURL)
			if err != nil {
				continue
			}

			commands = append(commands, *art)
		}
	}

	return commands, nil
}

// fetchPluginAgents fetches all agents from a plugin's agents/ directory
func (c *Client) fetchPluginAgents(apiURL string) ([]artifact.Artifact, error) {
	var agents []artifact.Artifact

	agentsURL := appendPath(apiURL, "agents")
	contents, err := c.ListGitHubContents(agentsURL)
	if err != nil {
		return nil, err
	}

	for _, item := range contents {
		if item.Type == "file" && strings.HasSuffix(strings.ToLower(item.Name), ".md") {
			content, err := c.FetchURL(item.DownloadURL)
			if err != nil {
				continue
			}

			// Parse agent similar to command
			art, err := ParseAgent(content, item.Name, item.DownloadURL)
			if err != nil {
				continue
			}

			agents = append(agents, *art)
		}
	}

	return agents, nil
}

// fetchPluginHooks fetches hooks from a plugin's hooks/ directory
func (c *Client) fetchPluginHooks(apiURL string) ([]artifact.Artifact, error) {
	var hooks []artifact.Artifact

	hooksURL := appendPath(apiURL, "hooks")
	contents, err := c.ListGitHubContents(hooksURL)
	if err != nil {
		return nil, err
	}

	for _, item := range contents {
		// Look for hooks.json or individual hook files
		if item.Type == "file" && item.Name == "hooks.json" {
			content, err := c.FetchURL(item.DownloadURL)
			if err != nil {
				continue
			}

			// Parse hooks.json - it contains an array of hook definitions
			parsedHooks, err := ParseHooksJSON(content, item.DownloadURL)
			if err != nil {
				continue
			}

			hooks = append(hooks, parsedHooks...)
		}
	}

	return hooks, nil
}

// ParseAgent parses an agent markdown file
func ParseAgent(content []byte, filename string, sourceURL string) (*artifact.Artifact, error) {
	fm, body, err := parseFrontmatter(content)
	if err != nil {
		return nil, err
	}

	name := fm.Name
	if name == "" {
		name = strings.TrimSuffix(filename, ".md")
	}

	description := fm.Description
	if description == "" {
		description = extractDescriptionFromContent(body)
	}

	return &artifact.Artifact{
		Name:        name,
		Type:        artifact.TypeAgent,
		Description: description,
		Version:     fm.Version,
		Author:      fm.Author,
		SourceURL:   sourceURL,
		Content:     string(content),
		Filename:    filename,
	}, nil
}

// HookDefinition represents a hook in hooks.json
type HookDefinition struct {
	Matcher string   `json:"matcher"`
	Hooks   []string `json:"hooks"`
}

// ParseHooksJSON parses a hooks.json file
func ParseHooksJSON(content []byte, sourceURL string) ([]artifact.Artifact, error) {
	var hookDefs []HookDefinition
	if err := json.Unmarshal(content, &hookDefs); err != nil {
		return nil, fmt.Errorf("failed to parse hooks.json: %w", err)
	}

	var hooks []artifact.Artifact
	for _, def := range hookDefs {
		for i, hookCmd := range def.Hooks {
			hooks = append(hooks, artifact.Artifact{
				Name:        fmt.Sprintf("%s-hook-%d", def.Matcher, i+1),
				Type:        artifact.TypeHook,
				Description: fmt.Sprintf("Hook for %s", def.Matcher),
				Event:       def.Matcher,
				Command:     hookCmd,
				SourceURL:   sourceURL,
				Content:     string(content),
				Filename:    "hooks.json",
			})
		}
	}

	return hooks, nil
}
