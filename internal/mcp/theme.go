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

	"github.com/fsnotify/fsnotify"
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

// CRC: crc-ThemeManager.md | Seq: seq-theme-audit.md
// GetAllThemeClasses scans all theme CSS files and returns a deduplicated union of all @class entries.
func GetAllThemeClasses(baseDir string) ([]ThemeClass, error) {
	themes, err := ListThemes(baseDir)
	if err != nil {
		return nil, fmt.Errorf("listing themes: %w", err)
	}

	seen := make(map[string]bool)
	var classes []ThemeClass

	for _, theme := range themes {
		fm, err := GetThemeClasses(baseDir, theme)
		if err != nil {
			continue // skip themes that fail to parse
		}
		for _, c := range fm.Classes {
			if !seen[c.Name] {
				seen[c.Name] = true
				classes = append(classes, c)
			}
		}
	}

	sort.Slice(classes, func(i, j int) bool {
		return classes[i].Name < classes[j].Name
	})
	return classes, nil
}

// ResolveThemeClasses returns the theme label and class list for a given theme name.
// If theme is empty, it returns the deduplicated union of all themes with the label "(all)".
func ResolveThemeClasses(baseDir, theme string) (string, []ThemeClass, error) {
	if theme == "" {
		classes, err := GetAllThemeClasses(baseDir)
		if err != nil {
			return "", nil, fmt.Errorf("getting all theme classes: %w", err)
		}
		return "(all)", classes, nil
	}
	fm, err := GetThemeClasses(baseDir, theme)
	if err != nil {
		return "", nil, fmt.Errorf("getting theme classes: %w", err)
	}
	return theme, fm.Classes, nil
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

	block := GenerateThemeBlock(baseDir, themes, GetCurrentTheme(baseDir))

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

// HasThemeBlock checks if index.html already contains the frictionless theme block.
// Seq: seq-theme-inject.md
func HasThemeBlock(baseDir string) bool {
	indexPath := filepath.Join(baseDir, "html", "index.html")
	content, err := os.ReadFile(indexPath)
	if err != nil {
		return false
	}
	return frictionlessBlockPattern.Match(content)
}

// WatchIndexHTML watches index.html for external writes and re-injects the theme block if missing.
// Seq: seq-theme-inject.md
func WatchIndexHTML(baseDir string, logFn func(level int, format string, args ...interface{})) (func(), error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("creating watcher: %w", err)
	}

	indexPath := filepath.Join(baseDir, "html", "index.html")
	themesDir := filepath.Join(baseDir, "html", "themes")

	// Watch the html directory for index.html changes
	watchDir := filepath.Join(baseDir, "html")
	if err := watcher.Add(watchDir); err != nil {
		watcher.Close()
		return nil, fmt.Errorf("watching %s: %w", watchDir, err)
	}

	// Watch the themes directory for CSS changes (cache busting)
	if _, err := os.Stat(themesDir); err == nil {
		if err := watcher.Add(themesDir); err != nil {
			watcher.Close()
			return nil, fmt.Errorf("watching %s: %w", themesDir, err)
		}
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
					continue
				}
				if event.Name == indexPath {
					// index.html changed — re-inject if theme block is missing
					if HasThemeBlock(baseDir) {
						continue
					}
					logFn(2, "index.html changed without theme block, re-injecting")
					if err := InjectThemeBlock(baseDir); err != nil {
						logFn(0, "Warning: failed to re-inject theme block: %v", err)
					}
				} else if strings.HasSuffix(event.Name, ".css") && strings.HasPrefix(event.Name, themesDir) {
					// Theme CSS changed — re-inject to update cache-busting timestamps
					logFn(2, "theme CSS changed: %s, updating cache-busting", filepath.Base(event.Name))
					if err := InjectThemeBlock(baseDir); err != nil {
						logFn(0, "Warning: failed to re-inject theme block: %v", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logFn(0, "Warning: index.html watcher error: %v", err)
			}
		}
	}()

	return func() { watcher.Close() }, nil
}

// GenerateThemeBlock creates the HTML block to inject into index.html
func GenerateThemeBlock(baseDir string, themes []string, defaultTheme string) string {
	var sb strings.Builder

	sb.WriteString("  <!-- #frictionless -->\n")

	// Theme restore script - runs before CSS loads
	sb.WriteString("  <script>\n")
	sb.WriteString(fmt.Sprintf("    document.documentElement.className = 'theme-' + (localStorage.getItem('theme') || '%s');\n", defaultTheme))
	sb.WriteString("  </script>\n")

	themesDir := filepath.Join(baseDir, "html", "themes")

	// Base CSS always first (with cache busting)
	baseCB := cssModTime(filepath.Join(themesDir, "base.css"))
	sb.WriteString(fmt.Sprintf("  <link rel=\"stylesheet\" href=\"/themes/base.css%s\">\n", baseCB))

	// Theme CSS files (with cache busting)
	for _, theme := range themes {
		cb := cssModTime(filepath.Join(themesDir, theme+".css"))
		sb.WriteString(fmt.Sprintf("  <link rel=\"stylesheet\" href=\"/themes/%s.css%s\">\n", theme, cb))
	}

	// Favicon placeholder - set dynamically by each app's DEFAULT viewdef script
	sb.WriteString("  <link rel=\"icon\" id=\"app-favicon\" href=\"data:,\">\n")

	sb.WriteString("  <!-- /frictionless -->\n")

	return sb.String()
}

// cssModTime returns a cache-busting query string based on the file's modification time.
// Returns empty string if the file cannot be stat'd.
func cssModTime(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("?v=%d", info.ModTime().Unix())
}

// isSkippedClass returns true for classes that should be excluded from auditing
func isSkippedClass(class string) bool {
	if class == "" || class == "hidden" {
		return true
	}
	return strings.HasPrefix(class, "sl-") || strings.HasPrefix(class, "ui-")
}

// AuditAppTheme compares an app's CSS class usage against documented theme classes.
// If theme is empty, it audits against the union of all themes.
func AuditAppTheme(baseDir, appName, theme string) (*ThemeAuditResult, error) {
	themeName, classes, err := ResolveThemeClasses(baseDir, theme)
	if err != nil {
		return nil, err
	}
	return AuditAppWithClasses(baseDir, appName, themeName, classes)
}

// AuditAppWithClasses compares an app's CSS class usage against a provided class list.
func AuditAppWithClasses(baseDir, appName, themeName string, classes []ThemeClass) (*ThemeAuditResult, error) {
	// Build set of documented classes
	documentedClasses := make(map[string]bool, len(classes))
	for _, c := range classes {
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
		Theme:               themeName,
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
	for _, c := range classes {
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
