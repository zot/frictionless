package mcp

// CRC: crc-ThemeManager.md | Seq: seq-theme-inject.md, seq-theme-list.md
// Theme management: parsing CSS themes, listing, auditing, and index.html injection

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ThemeFrontmatter represents theme metadata parsed from CSS comments
type ThemeFrontmatter struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Classes     []ThemeClass `json:"classes"`
}

// ThemeClass represents a semantic CSS class defined in a theme
type ThemeClass struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Usage       string   `json:"usage"`
	Elements    []string `json:"elements"`
}

// ThemeListResult is returned by the list action
type ThemeListResult struct {
	Themes  []ThemeInfo `json:"themes"`
	Current string      `json:"current"`
}

// ThemeInfo contains basic theme information
type ThemeInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AccentColor string `json:"accent_color,omitempty"`
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

// Patterns for parsing CSS theme files
var (
	// Matches @theme name in CSS comment
	themeNamePattern = regexp.MustCompile(`@theme\s+(\S+)`)
	// Matches @description ... (until next @ or end of comment)
	themeDescPattern = regexp.MustCompile(`@description\s+([^\n@]+)`)
	// Matches @class blocks - stops at next @class or end of indented section
	classBlockPattern = regexp.MustCompile(`(?m)@class\s+(\S+)\s*\n((?:\s+@(?:description|usage|elements)[^\n]*\n?)*)`)
	// Matches individual class attributes
	classDescPattern    = regexp.MustCompile(`@description\s+(.+)`)
	classUsagePattern   = regexp.MustCompile(`@usage\s+(.+)`)
	classElementPattern = regexp.MustCompile(`@elements\s+(.+)`)
	// cssClassPattern matches class="..." or class='...' in HTML
	cssClassPattern = regexp.MustCompile(`class=["']([^"']+)["']`)
	// frictionlessBlockPattern matches the injected block in index.html
	frictionlessBlockPattern = regexp.MustCompile(`(?s)<!--\s*#frictionless\s*-->.*?<!--\s*/frictionless\s*-->[\r\n]*`)
	// accentColorPattern matches --term-accent: value in CSS
	accentColorPattern = regexp.MustCompile(`--term-accent:\s*([^;]+);`)
)

// Default theme name
const defaultThemeName = "lcars"

// ParseThemeCSS extracts metadata from CSS comment block
func ParseThemeCSS(content []byte) (*ThemeFrontmatter, error) {
	text := string(content)

	fm := &ThemeFrontmatter{}

	// Extract theme name
	if match := themeNamePattern.FindStringSubmatch(text); match != nil {
		fm.Name = match[1]
	}

	// Extract description
	if match := themeDescPattern.FindStringSubmatch(text); match != nil {
		fm.Description = strings.TrimSpace(match[1])
	}

	// Extract class definitions
	classMatches := classBlockPattern.FindAllStringSubmatch(text, -1)
	for _, match := range classMatches {
		className := match[1]
		classBody := match[2]

		tc := ThemeClass{Name: className}

		if descMatch := classDescPattern.FindStringSubmatch(classBody); descMatch != nil {
			tc.Description = strings.TrimSpace(descMatch[1])
		}
		if usageMatch := classUsagePattern.FindStringSubmatch(classBody); usageMatch != nil {
			tc.Usage = strings.TrimSpace(usageMatch[1])
		}
		if elemMatch := classElementPattern.FindStringSubmatch(classBody); elemMatch != nil {
			elements := strings.Split(elemMatch[1], ",")
			for _, e := range elements {
				tc.Elements = append(tc.Elements, strings.TrimSpace(e))
			}
		}

		fm.Classes = append(fm.Classes, tc)
	}

	return fm, nil
}

// ListThemes returns all theme CSS files in the themes directory (excludes base.css)
func ListThemes(baseDir string) ([]string, error) {
	themesDir := filepath.Join(baseDir, "html", "themes")
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
		if strings.HasSuffix(name, ".css") && name != "base.css" {
			// Remove .css extension
			themes = append(themes, strings.TrimSuffix(name, ".css"))
		}
	}

	sort.Strings(themes)
	return themes, nil
}

// ListThemesWithInfo returns themes with their metadata
func ListThemesWithInfo(baseDir string) (*ThemeListResult, error) {
	themes, err := ListThemes(baseDir)
	if err != nil {
		return nil, err
	}

	result := &ThemeListResult{
		Themes:  make([]ThemeInfo, 0, len(themes)),
		Current: GetCurrentTheme(baseDir),
	}

	for _, theme := range themes {
		fm, err := GetThemeClasses(baseDir, theme)
		info := ThemeInfo{Name: theme}
		if err == nil && fm != nil {
			info.Description = fm.Description
		}
		// Extract accent color from theme CSS
		info.AccentColor = GetThemeAccentColor(baseDir, theme)
		result.Themes = append(result.Themes, info)
	}

	return result, nil
}

// GetThemeAccentColor extracts --term-accent from a theme CSS file
func GetThemeAccentColor(baseDir, theme string) string {
	themePath := filepath.Join(baseDir, "html", "themes", theme+".css")
	content, err := os.ReadFile(themePath)
	if err != nil {
		return ""
	}
	if match := accentColorPattern.FindSubmatch(content); match != nil {
		return strings.TrimSpace(string(match[1]))
	}
	return ""
}

// GetCurrentTheme returns the default theme (from config or hardcoded default)
func GetCurrentTheme(baseDir string) string {
	// TODO: Read from config file when implemented
	// For now, return default
	return defaultThemeName
}

// GetThemeClasses parses a theme CSS file and returns its documented classes
func GetThemeClasses(baseDir, theme string) (*ThemeFrontmatter, error) {
	themePath := filepath.Join(baseDir, "html", "themes", theme+".css")
	content, err := os.ReadFile(themePath)
	if err != nil {
		return nil, fmt.Errorf("reading theme file: %w", err)
	}

	return ParseThemeCSS(content)
}

// InjectThemeBlock updates index.html with the frictionless theme block
func InjectThemeBlock(baseDir string) error {
	indexPath := filepath.Join(baseDir, "html", "index.html")

	content, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("reading index.html: %w", err)
	}

	themes, err := ListThemes(baseDir)
	if err != nil {
		return fmt.Errorf("listing themes: %w", err)
	}

	block := GenerateThemeBlock(themes, GetCurrentTheme(baseDir))

	// Remove existing frictionless block if present
	html := frictionlessBlockPattern.ReplaceAllString(string(content), "")

	// Find <head> tag (case-insensitive)
	headIndex := strings.Index(strings.ToLower(html), "<head>")
	if headIndex == -1 {
		return fmt.Errorf("no <head> tag found in index.html")
	}

	// Insert after <head> tag, preserving any trailing newline
	insertPos := headIndex + len("<head>")
	if insertPos < len(html) && html[insertPos] == '\n' {
		insertPos++
	}

	newHTML := html[:insertPos] + "\n" + block + html[insertPos:]

	return os.WriteFile(indexPath, []byte(newHTML), 0644)
}

// GenerateThemeBlock creates the HTML block to inject into index.html
func GenerateThemeBlock(themes []string, defaultTheme string) string {
	var sb strings.Builder

	sb.WriteString("  <!-- #frictionless -->\n")

	// Theme restore script - runs before CSS loads
	sb.WriteString("  <script>\n")
	sb.WriteString(fmt.Sprintf("    document.documentElement.className = 'theme-' + (localStorage.getItem('theme') || '%s');\n", defaultTheme))
	sb.WriteString("  </script>\n")

	// Base CSS always first
	sb.WriteString("  <link rel=\"stylesheet\" href=\"/themes/base.css\">\n")

	// Theme CSS files
	for _, theme := range themes {
		sb.WriteString(fmt.Sprintf("  <link rel=\"stylesheet\" href=\"/themes/%s.css\">\n", theme))
	}

	sb.WriteString("  <!-- /frictionless -->\n")

	return sb.String()
}

// isSkippedClass returns true for classes that should be excluded from auditing
func isSkippedClass(class string) bool {
	if class == "" || class == "hidden" {
		return true
	}
	return strings.HasPrefix(class, "sl-") || strings.HasPrefix(class, "ui-")
}

// AuditAppTheme compares an app's CSS class usage against documented theme classes
func AuditAppTheme(baseDir, appName, theme string) (*ThemeAuditResult, error) {
	themeFM, err := GetThemeClasses(baseDir, theme)
	if err != nil {
		return nil, fmt.Errorf("getting theme classes: %w", err)
	}

	// Build set of documented classes
	documentedClasses := make(map[string]bool, len(themeFM.Classes))
	for _, c := range themeFM.Classes {
		documentedClasses[c.Name] = true
	}

	viewdefsPath := filepath.Join(baseDir, "apps", appName, "viewdefs")
	entries, err := os.ReadDir(viewdefsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("app %s has no viewdefs directory", appName)
		}
		return nil, err
	}

	usedClasses := make(map[string]ClassUsage)
	themeClassesUsed := make(map[string]bool)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".html") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(viewdefsPath, entry.Name()))
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		for lineNum, line := range lines {
			for _, match := range cssClassPattern.FindAllStringSubmatch(line, -1) {
				for _, class := range strings.Fields(match[1]) {
					if isSkippedClass(class) {
						continue
					}

					if _, seen := usedClasses[class]; !seen {
						usedClasses[class] = ClassUsage{
							Class: class,
							File:  entry.Name(),
							Line:  lineNum + 1,
						}
					}

					if documentedClasses[class] {
						themeClassesUsed[class] = true
					}
				}
			}
		}
	}

	result := &ThemeAuditResult{
		App:                 appName,
		Theme:               theme,
		UndocumentedClasses: make([]ClassUsage, 0),
		UnusedThemeClasses:  make([]string, 0),
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

	// Sort results for deterministic output
	sort.Slice(result.UndocumentedClasses, func(i, j int) bool {
		return result.UndocumentedClasses[i].Class < result.UndocumentedClasses[j].Class
	})
	sort.Strings(result.UnusedThemeClasses)

	result.Summary.Total = len(usedClasses)
	result.Summary.Undocumented = len(result.UndocumentedClasses)
	result.Summary.Documented = result.Summary.Total - result.Summary.Undocumented

	return result, nil
}
