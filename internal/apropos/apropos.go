package apropos

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/kennyg/tome/internal/artifact"
)

// Index holds the apropos index data
type Index struct {
	Generated time.Time `yaml:"generated"`
	Skills    []Skill   `yaml:"skills"`
}

// Skill represents an indexed skill
type Skill struct {
	Name        string   `yaml:"name"`
	Path        string   `yaml:"path"`
	Description string   `yaml:"description"`
	Keywords    []string `yaml:"keywords"`
	ModTime     int64    `yaml:"mod_time"`
}

// Frontmatter represents the YAML frontmatter of a SKILL.md
type Frontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

const IndexFileName = ".apropos"

// common stopwords to filter out when extracting keywords
var stopwords = map[string]bool{
	"a": true, "an": true, "the": true, "and": true, "or": true, "but": true,
	"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
	"with": true, "by": true, "from": true, "as": true, "is": true, "was": true,
	"are": true, "were": true, "been": true, "be": true, "have": true, "has": true,
	"had": true, "do": true, "does": true, "did": true, "will": true, "would": true,
	"could": true, "should": true, "may": true, "might": true, "must": true,
	"this": true, "that": true, "these": true, "those": true, "it": true,
	"its": true, "when": true, "claude": true, "needs": true, "use": true,
	"using": true, "used": true, "can": true, "any": true, "other": true,
}

// LoadIndex loads the index from a skills directory
func LoadIndex(skillsDir string) (*Index, error) {
	indexPath := filepath.Join(skillsDir, IndexFileName)
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var index Index
	if err := yaml.Unmarshal(data, &index); err != nil {
		return nil, err
	}

	return &index, nil
}

// SaveIndex saves the index to a skills directory
func SaveIndex(skillsDir string, index *Index) error {
	indexPath := filepath.Join(skillsDir, IndexFileName)
	data, err := yaml.Marshal(index)
	if err != nil {
		return err
	}

	header := "# Apropos index - auto-generated, do not edit\n"
	return os.WriteFile(indexPath, append([]byte(header), data...), 0644)
}

// IsStale checks if the index is stale (any SKILL.md newer than index)
func IsStale(skillsDir string, index *Index) (bool, error) {
	if index == nil {
		return true, nil
	}

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return true, err
	}

	// Build a map of indexed skills by path for quick lookup
	indexed := make(map[string]int64)
	for _, s := range index.Skills {
		indexed[s.Path] = s.ModTime
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		skillPath := filepath.Join(skillsDir, entry.Name())
		skillMdPath := filepath.Join(skillPath, artifact.SkillFilename)

		info, err := os.Stat(skillMdPath)
		if err != nil {
			continue // skip skills without SKILL.md
		}

		modTime := info.ModTime().Unix()

		// Check if this skill is in the index and up to date
		if indexedTime, ok := indexed[skillPath]; !ok || indexedTime != modTime {
			return true, nil
		}
		delete(indexed, skillPath)
	}

	// If there are skills in the index that don't exist anymore, it's stale
	if len(indexed) > 0 {
		return true, nil
	}

	return false, nil
}

// BuildIndex scans skills directories and builds a fresh index
func BuildIndex(skillsDirs []string) (*Index, error) {
	index := &Index{
		Generated: time.Now(),
		Skills:    []Skill{},
	}

	for _, dir := range skillsDirs {
		skills, err := scanSkillsDir(dir)
		if err != nil {
			continue // skip dirs that don't exist
		}
		index.Skills = append(index.Skills, skills...)
	}

	return index, nil
}

func scanSkillsDir(skillsDir string) ([]Skill, error) {
	var skills []Skill

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		skillPath := filepath.Join(skillsDir, entry.Name())
		skill, err := parseSkill(skillPath)
		if err != nil {
			continue // skip invalid skills
		}
		skills = append(skills, *skill)
	}

	return skills, nil
}

func parseSkill(skillPath string) (*Skill, error) {
	skillMdPath := filepath.Join(skillPath, artifact.SkillFilename)

	info, err := os.Stat(skillMdPath)
	if err != nil {
		return nil, err
	}

	frontmatter, err := parseFrontmatter(skillMdPath)
	if err != nil {
		return nil, err
	}

	keywords := extractKeywords(frontmatter.Description)

	return &Skill{
		Name:        frontmatter.Name,
		Path:        skillPath,
		Description: frontmatter.Description,
		Keywords:    keywords,
		ModTime:     info.ModTime().Unix(),
	}, nil
}

func parseFrontmatter(path string) (*Frontmatter, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Look for opening ---
	if !scanner.Scan() {
		return nil, os.ErrNotExist
	}
	if strings.TrimSpace(scanner.Text()) != "---" {
		return nil, os.ErrNotExist
	}

	// Collect YAML until closing ---
	var yamlLines []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			break
		}
		yamlLines = append(yamlLines, line)
	}

	var fm Frontmatter
	yamlContent := strings.Join(yamlLines, "\n")
	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return nil, err
	}

	return &fm, nil
}

func extractKeywords(description string) []string {
	// Normalize: lowercase, remove punctuation
	re := regexp.MustCompile(`[^a-zA-Z0-9\s]`)
	normalized := re.ReplaceAllString(strings.ToLower(description), " ")

	words := strings.Fields(normalized)

	// Dedupe and filter
	seen := make(map[string]bool)
	var keywords []string

	for _, word := range words {
		if len(word) < 3 {
			continue
		}
		if stopwords[word] {
			continue
		}
		if seen[word] {
			continue
		}
		seen[word] = true
		keywords = append(keywords, word)
	}

	return keywords
}

// SearchResult represents a search match
type SearchResult struct {
	Skill Skill
	Score int // higher is better
}

// Search searches the index for skills matching the query
func Search(index *Index, query string) []SearchResult {
	if index == nil || len(index.Skills) == 0 {
		return nil
	}

	queryWords := strings.Fields(strings.ToLower(query))
	var results []SearchResult

	for _, skill := range index.Skills {
		score := scoreMatch(skill, queryWords)
		if score > 0 {
			results = append(results, SearchResult{
				Skill: skill,
				Score: score,
			})
		}
	}

	// Sort by score descending
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}

func scoreMatch(skill Skill, queryWords []string) int {
	score := 0
	nameLower := strings.ToLower(skill.Name)
	descLower := strings.ToLower(skill.Description)

	for _, qw := range queryWords {
		// Exact name match is highest value
		if nameLower == qw {
			score += 100
		} else if strings.Contains(nameLower, qw) {
			score += 50
		}

		// Description contains query word
		if strings.Contains(descLower, qw) {
			score += 10
		}

		// Keyword match
		for _, kw := range skill.Keywords {
			if kw == qw {
				score += 20
			} else if strings.Contains(kw, qw) {
				score += 5
			}
		}
	}

	return score
}

// List returns all skills in the index
func List(index *Index) []Skill {
	if index == nil {
		return nil
	}
	return index.Skills
}
