package mcp

// CRC: crc-Auditor.md | Seq: seq-audit.md

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// AuditResult contains the results of auditing an app
type AuditResult struct {
	App        string       `json:"app"`
	Violations []Violation  `json:"violations"`
	Warnings   []Violation  `json:"warnings"`
	Summary    AuditSummary `json:"summary"`
}

// Violation represents a single audit finding
type Violation struct {
	Type     string `json:"type"`
	Location string `json:"location"`
	Detail   string `json:"detail"`
}

// AuditSummary provides counts of findings
type AuditSummary struct {
	TotalMethods       int `json:"total_methods"`
	DeadMethods        int `json:"dead_methods"`
	ViewdefViolations  int `json:"viewdef_violations"`
}

// Known method lists
var (
	// Framework methods are never flagged as dead
	frameworkMethods = map[string]bool{
		"new":    true,
		"mutate": true,
	}

	// External methods are flagged as warnings, not violations
	externalMethods = map[string]bool{
		"addAgentMessage":    true,
		"updateRequirements": true,
		"onAppProgress":      true,
		"onAppUpdated":       true,
	}

	// Button elements where ui-action is valid
	buttonElements = map[string]bool{
		"button":         true,
		"sl-button":      true,
		"sl-icon-button": true,
	}

	// Built-in viewdef functions that don't need Lua definitions
	builtinViewdefFunctions = map[string]bool{
		"wrapper": true, // ViewList wrapper parameter
	}
)

// Regex patterns for Lua analysis
var (
	// Matches: function TypeName:methodName(
	methodDefPattern = regexp.MustCompile(`function\s+(\w+):(\w+)\s*\(`)

	// Matches: :methodName(
	methodCallPattern = regexp.MustCompile(`:(\w+)\s*\(`)

	// Matches: methodName() or methodName(_) in attribute values
	viewdefCallPattern = regexp.MustCompile(`(\w+)\((_)?\)`)

	// Matches: globalName = TypeName:new(
	globalAssignPattern = regexp.MustCompile(`^(\w+)\s*=\s*\w+:new\(`)

	// Matches: if not session.reloading then
	reloadingGuardPattern = regexp.MustCompile(`if\s+not\s+session\.reloading\s+then`)

	// Matches operators in paths
	// Note: we strip () contents before checking, so no need for lookahead
	operatorPattern = regexp.MustCompile(`[!&|]|==|~=|\+|-`)

	// Matches method calls with non-empty args (not () or (_))
	// Captures: method name, args content
	nonEmptyArgsPattern = regexp.MustCompile(`(\w+)\(([^)]+)\)`)

	// Path syntax validation regex
	// prefix: ident | bracket | ident()
	// suffix: prefix | ident(_)
	// property: ident(=text)?
	// path: (prefix.)* suffix (?property(&property)*)?
	pathSyntaxPattern = regexp.MustCompile(
		`^(?:(?:[a-zA-Z_]\w*(?:\(\))?|\[(?:[a-zA-Z_]\w*|\d+)\])\.)*` + // prefixes with dots
			`(?:[a-zA-Z_]\w*(?:\(\)|(?:\(_\)))?|\[(?:[a-zA-Z_]\w*|\d+)\])` + // suffix
			`(?:\?[a-zA-Z_]\w*(?:=[^&]*)?(?:&[a-zA-Z_]\w*(?:=[^&]*)?)*)?$`) // optional properties
)

// AuditApp performs a full audit of an app
// CRC: crc-Auditor.md
func AuditApp(baseDir, appName string) (*AuditResult, error) {
	appPath := filepath.Join(baseDir, "apps", appName)

	result := &AuditResult{
		App:        appName,
		Violations: []Violation{},
		Warnings:   []Violation{},
	}

	// Scan all .lua files in the app directory
	methodDefs := make(map[string]string)
	luaCalls := make(map[string]bool)
	foundAppLua := false

	luaFiles, err := filepath.Glob(filepath.Join(appPath, "*.lua"))
	if err != nil {
		return nil, fmt.Errorf("scanning lua files: %w", err)
	}

	for _, luaFile := range luaFiles {
		content, err := os.ReadFile(luaFile)
		if err != nil {
			continue
		}

		filename := filepath.Base(luaFile)
		isAppLua := filename == "app.lua"
		if isAppLua {
			foundAppLua = true
		}

		// Only check guards/global in app.lua, but extract defs/calls from all files
		var checkResult *AuditResult
		if isAppLua {
			checkResult = result
		}

		defs, calls := analyzeLua(string(content), appName, checkResult)

		// Merge definitions and calls
		for name, fullName := range defs {
			methodDefs[name] = fullName
		}
		for call := range calls {
			luaCalls[call] = true
		}
	}

	if !foundAppLua {
		result.Violations = append(result.Violations, Violation{
			Type:     "missing_app_lua",
			Location: "app.lua",
			Detail:   "app.lua not found",
		})
		return result, nil
	}

	// Collect viewdef method calls
	viewdefCalls := make(map[string]bool)

	// Read and analyze viewdefs
	viewdefsPath := filepath.Join(appPath, "viewdefs")
	entries, err := os.ReadDir(viewdefsPath)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".html") {
				continue
			}

			filePath := filepath.Join(viewdefsPath, entry.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}

			isListItem := strings.HasSuffix(entry.Name(), ".list-item.html")
			calls := analyzeViewdef(entry.Name(), string(content), isListItem, result)

			for call := range calls {
				viewdefCalls[call] = true
			}
		}
	}

	// Find dead methods
	findDeadMethods(methodDefs, luaCalls, viewdefCalls, result)

	// Find missing methods (viewdef calls that don't exist in Lua)
	findMissingMethods(methodDefs, viewdefCalls, result)

	// Calculate summary
	result.Summary.TotalMethods = len(methodDefs)
	result.Summary.DeadMethods = 0
	for _, v := range result.Violations {
		if v.Type == "dead_method" {
			result.Summary.DeadMethods++
		} else {
			result.Summary.ViewdefViolations++
		}
	}

	return result, nil
}

// analyzeLua extracts method definitions and calls, checks guards and global name
// CRC: crc-Auditor.md
func analyzeLua(content, appName string, result *AuditResult) (defs map[string]string, calls map[string]bool) {
	defs = make(map[string]string)  // method name -> "Type:method"
	calls = make(map[string]bool)

	// Process line by line to distinguish definitions from calls
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		// Check if this line is a function definition
		if defMatch := methodDefPattern.FindStringSubmatch(line); defMatch != nil {
			typeName := defMatch[1]
			methodName := defMatch[2]
			fullName := typeName + ":" + methodName
			defs[methodName] = fullName
			// Don't extract calls from definition lines
			continue
		}

		// Extract method calls from non-definition lines
		for _, match := range methodCallPattern.FindAllStringSubmatch(line, -1) {
			calls[match[1]] = true
		}
	}

	// Check reloading guard
	checkReloadingGuard(content, result)

	// Check global name
	checkGlobalName(content, appName, result)

	return defs, calls
}

// checkReloadingGuard verifies instance creation is wrapped in reloading check
// CRC: crc-Auditor.md
func checkReloadingGuard(content string, result *AuditResult) {
	// Look for global = Type:new( patterns
	lines := strings.Split(content, "\n")
	inReloadingBlock := false
	braceDepth := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track if we're inside a reloading guard block
		if reloadingGuardPattern.MatchString(trimmed) {
			inReloadingBlock = true
			braceDepth = 1
		}

		if inReloadingBlock {
			braceDepth += strings.Count(trimmed, "then") + strings.Count(trimmed, "do") +
				strings.Count(trimmed, "function")
			braceDepth -= strings.Count(trimmed, "end")
			if braceDepth <= 0 {
				inReloadingBlock = false
			}
		}

		// Check for global assignments outside of reloading guard
		if globalAssignPattern.MatchString(trimmed) && !inReloadingBlock {
			// Ignore if it's inside a function (indented)
			if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
				result.Violations = append(result.Violations, Violation{
					Type:     "missing_reloading_guard",
					Location: fmt.Sprintf("app.lua:%d", i+1),
					Detail:   fmt.Sprintf("Instance creation not guarded: %s", strings.TrimSpace(line)),
				})
			}
		}
	}
}

// checkGlobalName verifies the global variable matches the app directory name
// CRC: crc-Auditor.md
func checkGlobalName(content, appName string, result *AuditResult) {
	// Convert app name to expected global (kebab-case to camelCase)
	expected := kebabToCamel(appName)

	// Look for the global assignment
	matches := globalAssignPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		globalName := match[1]
		// Check if it matches expected (case-insensitive for the first char)
		if !strings.EqualFold(globalName, expected) && globalName != appName {
			result.Violations = append(result.Violations, Violation{
				Type:     "global_name_mismatch",
				Location: "app.lua",
				Detail:   fmt.Sprintf("Global '%s' should be '%s' (matching directory)", globalName, expected),
			})
			break // Only report once
		}
	}
}

// kebabToCamel converts kebab-case to camelCase
func kebabToCamel(s string) string {
	parts := strings.Split(s, "-")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// analyzeViewdef parses HTML and checks for violations
// CRC: crc-Auditor.md
func analyzeViewdef(filename, content string, isListItem bool, result *AuditResult) map[string]bool {
	calls := make(map[string]bool)

	// Parse HTML
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		result.Violations = append(result.Violations, Violation{
			Type:     "html_parse_error",
			Location: fmt.Sprintf("viewdefs/%s", filename),
			Detail:   fmt.Sprintf("HTML parse error: %s", err.Error()),
		})
		return calls
	}

	// Walk DOM
	walkDOM(doc, filename, isListItem, result, calls)

	return calls
}

// walkDOM recursively checks each node for violations
// CRC: crc-Auditor.md
func walkDOM(n *html.Node, filename string, isListItem bool, result *AuditResult, calls map[string]bool) {
	if n.Type == html.ElementNode {
		tagName := n.Data

		// Check for style tag in list-item
		if isListItem && tagName == "style" {
			result.Violations = append(result.Violations, Violation{
				Type:     "style_in_list_item",
				Location: fmt.Sprintf("viewdefs/%s", filename),
				Detail:   "<style> block found in list-item viewdef (put styles in top-level viewdef)",
			})
		}

		// Check attributes
		for _, attr := range n.Attr {
			// Check ui-action on non-button
			if attr.Key == "ui-action" && !buttonElements[tagName] {
				result.Violations = append(result.Violations, Violation{
					Type:     "ui_action_non_button",
					Location: fmt.Sprintf("viewdefs/%s", filename),
					Detail:   fmt.Sprintf("ui-action on <%s> (use ui-event-click for non-buttons)", tagName),
				})
			}

			// Check wrong hidden syntax
			if attr.Key == "ui-class" && strings.Contains(attr.Val, "hidden:") {
				result.Violations = append(result.Violations, Violation{
					Type:     "wrong_hidden_syntax",
					Location: fmt.Sprintf("viewdefs/%s", filename),
					Detail:   fmt.Sprintf("Use ui-class-hidden instead of ui-class=\"hidden:...\""),
				})
			}

			// Check ui-value on checkbox/switch
			if attr.Key == "ui-value" && (tagName == "sl-checkbox" || tagName == "sl-switch") {
				result.Violations = append(result.Violations, Violation{
					Type:     "ui_value_checkbox",
					Location: fmt.Sprintf("viewdefs/%s", filename),
					Detail:   fmt.Sprintf("ui-value on <%s> renders boolean as text (use ui-attr-checked)", tagName),
				})
			}

			// Check ui-value on sl-badge
			if attr.Key == "ui-value" && tagName == "sl-badge" {
				result.Violations = append(result.Violations, Violation{
					Type:     "ui_value_badge",
					Location: fmt.Sprintf("viewdefs/%s", filename),
					Detail:   "ui-value on <sl-badge> not supported; use <span ui-value=\"...\"></span> inside the badge",
				})
			}

			// Check for ui-* attributes
			if strings.HasPrefix(attr.Key, "ui-") {
				// Check for item. prefix in list-item
				if isListItem && strings.Contains(attr.Val, "item.") {
					result.Violations = append(result.Violations, Violation{
						Type:     "item_prefix",
						Location: fmt.Sprintf("viewdefs/%s", filename),
						Detail:   fmt.Sprintf("Remove 'item.' prefix - item IS the context in list-item viewdefs"),
					})
				}

				// Check for operators in paths
				// Skip ui-namespace which is a viewdef namespace identifier, not a binding path
				// Skip checking inside () since those are method calls
				if attr.Key != "ui-namespace" {
					pathPart := attr.Val
					if idx := strings.Index(pathPart, "?"); idx != -1 {
						pathPart = pathPart[:idx] // Remove query params
					}
					// Remove method call parentheses content for operator check
					cleanPath := regexp.MustCompile(`\([^)]*\)`).ReplaceAllString(pathPart, "()")
					if operatorPattern.MatchString(cleanPath) {
						result.Violations = append(result.Violations, Violation{
							Type:     "operator_in_path",
							Location: fmt.Sprintf("viewdefs/%s", filename),
							Detail:   fmt.Sprintf("Operators in path '%s' (use Lua methods instead)", attr.Val),
						})
					}

					// Check for non-empty method args (only () or (_) allowed)
					for _, match := range nonEmptyArgsPattern.FindAllStringSubmatch(attr.Val, -1) {
						args := match[2]
						if args != "_" {
							result.Violations = append(result.Violations, Violation{
								Type:     "non_empty_method_args",
								Location: fmt.Sprintf("viewdefs/%s", filename),
								Detail:   fmt.Sprintf("Method '%s(%s)' has invalid args; only () or (_) allowed", match[1], args),
							})
						}
					}

					// Validate path syntax as final check
					if !pathSyntaxPattern.MatchString(attr.Val) {
						result.Violations = append(result.Violations, Violation{
							Type:     "invalid_path_syntax",
							Location: fmt.Sprintf("viewdefs/%s", filename),
							Detail:   fmt.Sprintf("Invalid path syntax: '%s'", attr.Val),
						})
					}
				}

				// Extract method calls
				for _, match := range viewdefCallPattern.FindAllStringSubmatch(attr.Val, -1) {
					calls[match[1]] = true
				}
			}
		}
	}

	// Recurse to children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkDOM(c, filename, isListItem, result, calls)
	}
}

// findDeadMethods identifies methods defined but never called
// CRC: crc-Auditor.md
func findDeadMethods(defs map[string]string, luaCalls, viewdefCalls map[string]bool, result *AuditResult) {
	for methodName, fullName := range defs {
		// Skip framework methods
		if frameworkMethods[methodName] {
			continue
		}

		// Skip mcp:* methods (MCP extension points called externally by Claude)
		if strings.HasPrefix(fullName, "mcp:") {
			continue
		}

		// Check if called from Lua or viewdefs
		if luaCalls[methodName] || viewdefCalls[methodName] {
			continue
		}

		// Check if it's an external method (warning, not violation)
		if externalMethods[methodName] {
			result.Warnings = append(result.Warnings, Violation{
				Type:     "external_method",
				Location: "app.lua",
				Detail:   fmt.Sprintf("%s (called by Claude via ui_run)", fullName),
			})
			continue
		}

		// Dead method
		result.Violations = append(result.Violations, Violation{
			Type:     "dead_method",
			Location: "app.lua",
			Detail:   fullName,
		})
	}
}

// findMissingMethods identifies viewdef calls that don't match any Lua method definition
// CRC: crc-Auditor.md
func findMissingMethods(defs map[string]string, viewdefCalls map[string]bool, result *AuditResult) {
	for callName := range viewdefCalls {
		if builtinViewdefFunctions[callName] {
			continue
		}
		if _, exists := defs[callName]; !exists {
			result.Violations = append(result.Violations, Violation{
				Type:     "missing_method",
				Location: "viewdefs",
				Detail:   fmt.Sprintf("Method '%s()' called in viewdef but not defined in Lua", callName),
			})
		}
	}
}
