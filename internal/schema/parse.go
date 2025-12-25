package schema

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseFrontmatter extracts YAML frontmatter from content.
// Returns the parsed frontmatter map, the body content, and any error.
func ParseFrontmatter(content []byte) (map[string]interface{}, string, error) {
	text := string(content)
	fm := make(map[string]interface{})

	// Check for frontmatter delimiter
	if !strings.HasPrefix(text, "---") {
		return fm, text, nil
	}

	// Find the closing delimiter
	rest := strings.TrimPrefix(text[3:], "\n")

	idx := strings.Index(rest, "\n---")
	if idx == -1 {
		return fm, text, nil
	}

	// Extract and parse YAML
	yamlContent := rest[:idx]
	body := strings.TrimPrefix(rest[idx+4:], "\n")

	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return nil, "", fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return fm, body, nil
}

// ParseFrontmatterTyped extracts YAML frontmatter into a typed struct.
// Returns the body content and any error.
func ParseFrontmatterTyped[T any](content []byte, target *T) (string, error) {
	text := string(content)

	// Check for frontmatter delimiter
	if !strings.HasPrefix(text, "---") {
		return text, nil
	}

	// Find the closing delimiter
	rest := strings.TrimPrefix(text[3:], "\n")

	idx := strings.Index(rest, "\n---")
	if idx == -1 {
		return text, nil
	}

	// Extract and parse YAML
	yamlContent := rest[:idx]
	body := strings.TrimPrefix(rest[idx+4:], "\n")

	if err := yaml.Unmarshal([]byte(yamlContent), target); err != nil {
		return "", fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return body, nil
}

// SerializeFrontmatter creates content with YAML frontmatter and body.
func SerializeFrontmatter(fm interface{}, body string) ([]byte, error) {
	yamlBytes, err := yaml.Marshal(fm)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize frontmatter: %w", err)
	}

	var result strings.Builder
	result.WriteString("---\n")
	result.Write(yamlBytes)
	result.WriteString("---\n")
	if body != "" {
		result.WriteString("\n")
		result.WriteString(body)
	}

	return []byte(result.String()), nil
}

// getString safely extracts a string from a map
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// getStringSlice safely extracts a string slice from a map
func getStringSlice(m map[string]interface{}, key string) []string {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case []string:
			return val
		case []interface{}:
			result := make([]string, 0, len(val))
			for _, item := range val {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
	}
	return nil
}
