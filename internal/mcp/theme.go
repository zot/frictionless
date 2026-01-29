package mcp

// Theme management: parsing, listing, and auditing theme class usage

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// ThemeFrontmatter represents the YAML frontmatter in a theme markdown file
type ThemeFrontmatter struct {
	Name        string       `yaml:"name" json:"name"`
	Description string       `yaml:"description" json:"description"`
	Classes     []ThemeClass `yaml:"classes" json:"classes"`
}

// ThemeClass represents a semantic CSS class defined in a theme
type ThemeClass struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	Usage       string   `yaml:"usage" json:"usage"`
	Elements    []string `yaml:"elements" json:"elements"`
}

// ThemeListResult is returned by the list action
type ThemeListResult struct {
	Themes  []string `json:"themes"`
	Current string   `json:"current"`
}

// ThemeClassesResult is returned by the classes action
type ThemeClassesResult struct {
	Theme   string       `json:"theme"`
	Classes []ThemeClass `json:"classes"`
}

// ClassUsage tracks where a CSS class is used
type ClassUsage struct {
	Class string `json:"class"`
	File  string `json:"file"`
	Line  int    `json:"line,omitempty"`
}

// ThemeAuditSummary provides counts for theme auditing
type ThemeAuditSummary struct {
	Total        int `json:"total"`
	Documented   int `json:"documented"`
	Undocumented int `json:"undocumented"`
}

// ThemeAuditResult contains results of auditing an app's theme usage
type ThemeAuditResult struct {
	App                 string            `json:"app"`
	Theme               string            `json:"theme"`
	UndocumentedClasses []ClassUsage      `json:"undocumented_classes"`
	UnusedThemeClasses  []string          `json:"unused_theme_classes"`
	Summary             ThemeAuditSummary `json:"summary"`
}

// frontmatterPattern matches YAML frontmatter between --- markers
var frontmatterPattern = regexp.MustCompile(`(?s)^---\n(.+?)\n---`)

// cssClassPattern matches class="..." or class='...' in HTML
var cssClassPattern = regexp.MustCompile(`class=["']([^"']+)["']`)

// ParseThemeFrontmatter extracts YAML frontmatter from markdown content
func ParseThemeFrontmatter(content []byte) (*ThemeFrontmatter, error) {
	match := frontmatterPattern.FindSubmatch(content)
	if match == nil {
		return nil, fmt.Errorf("no frontmatter found")
	}

	var fm ThemeFrontmatter
	if err := yaml.Unmarshal(match[1], &fm); err != nil {
		return nil, fmt.Errorf("parsing frontmatter: %w", err)
	}

	return &fm, nil
}

// ListThemes returns all theme markdown files in the themes directory
func ListThemes(baseDir string) ([]string, error) {
	themesDir := filepath.Join(baseDir, "themes")
	entries, err := os.ReadDir(themesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var themes []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".md") && name != "theme.md" {
			// Remove .md extension
			themes = append(themes, strings.TrimSuffix(name, ".md"))
		}
	}

	return themes, nil
}

// GetCurrentTheme resolves the theme.md symlink to get the active theme name
func GetCurrentTheme(baseDir string) string {
	themePath := filepath.Join(baseDir, "themes", "theme.md")
	target, err := os.Readlink(themePath)
	if err != nil {
		return ""
	}
	// target is like "lcars.md", extract "lcars"
	return strings.TrimSuffix(filepath.Base(target), ".md")
}

// GetThemeClasses parses a theme file and returns its documented classes
func GetThemeClasses(baseDir, theme string) (*ThemeFrontmatter, error) {
	themePath := filepath.Join(baseDir, "themes", theme+".md")
	content, err := os.ReadFile(themePath)
	if err != nil {
		return nil, fmt.Errorf("reading theme file: %w", err)
	}

	return ParseThemeFrontmatter(content)
}

// AuditAppTheme compares an app's CSS class usage against documented theme classes
func AuditAppTheme(baseDir, appName, theme string) (*ThemeAuditResult, error) {
	// Get theme classes
	themeFM, err := GetThemeClasses(baseDir, theme)
	if err != nil {
		return nil, fmt.Errorf("getting theme classes: %w", err)
	}

	// Build map of documented classes
	documentedClasses := make(map[string]bool)
	for _, c := range themeFM.Classes {
		documentedClasses[c.Name] = true
	}

	// Collect all CSS classes used in the app's viewdefs
	appPath := filepath.Join(baseDir, "apps", appName)
	viewdefsPath := filepath.Join(appPath, "viewdefs")

	usedClasses := make(map[string]ClassUsage)
	themeClassesUsed := make(map[string]bool)

	entries, err := os.ReadDir(viewdefsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("app %s has no viewdefs directory", appName)
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".html") {
			continue
		}

		filePath := filepath.Join(viewdefsPath, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		// Extract all class names from the file
		lines := strings.Split(string(content), "\n")
		for lineNum, line := range lines {
			matches := cssClassPattern.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				// Split class attribute value by whitespace
				classes := strings.Fields(match[1])
				for _, class := range classes {
					// Skip classes that are clearly dynamic or utility
					if strings.HasPrefix(class, "sl-") ||
						strings.HasPrefix(class, "ui-") ||
						class == "hidden" ||
						class == "" {
						continue
					}

					// Track usage
					if _, seen := usedClasses[class]; !seen {
						usedClasses[class] = ClassUsage{
							Class: class,
							File:  entry.Name(),
							Line:  lineNum + 1,
						}
					}

					// Track if it's a theme class
					if documentedClasses[class] {
						themeClassesUsed[class] = true
					}
				}
			}
		}
	}

	// Build result
	result := &ThemeAuditResult{
		App:                 appName,
		Theme:               theme,
		UndocumentedClasses: []ClassUsage{},
		UnusedThemeClasses:  []string{},
	}

	// Find undocumented classes (used but not in theme)
	for class, usage := range usedClasses {
		if !documentedClasses[class] {
			result.UndocumentedClasses = append(result.UndocumentedClasses, usage)
		}
	}

	// Find unused theme classes (in theme but not used by this app)
	for _, c := range themeFM.Classes {
		if !themeClassesUsed[c.Name] {
			result.UnusedThemeClasses = append(result.UnusedThemeClasses, c.Name)
		}
	}

	// Summary
	result.Summary.Total = len(usedClasses)
	result.Summary.Documented = result.Summary.Total - len(result.UndocumentedClasses)
	result.Summary.Undocumented = len(result.UndocumentedClasses)

	return result, nil
}
