# Auditor

**Source Spec:** specs/ui-audit.md

Analyzes ui-mcp apps for code quality violations.

## Knows

- frameworkMethods: Methods never flagged as dead (`new`, `mutate`)
- externalMethods: Methods called by Claude, flagged as warnings (`addAgentMessage`, `updateRequirements`, `onAppProgress`, `onAppUpdated`)
- buttonElements: Elements where `ui-action` is valid (`button`, `sl-button`, `sl-icon-button`)
- operatorChars: Characters invalid in binding paths (`!`, `=`, `&`, `|`, `+`, `-`)

## Does

- **AuditApp(baseDir, appName)**: Orchestrates full audit, returns AuditResult
- **analyzeLua(content, appName)**: Extracts method definitions/calls, checks reloading guard and global name
- **analyzeViewdef(path, content, isListItem)**: Parses HTML, walks DOM for violations
- **extractMethodDefs(content)**: Regex extracts `function Type:method(` patterns
- **extractLuaCalls(content)**: Regex extracts `:method(` call patterns
- **extractViewdefCalls(attrs)**: Extracts `method()` from ui-* attribute values
- **findDeadMethods(defs, luaCalls, viewdefCalls)**: Cross-references to find unused methods
- **checkReloadingGuard(content)**: Verifies instance creation is guarded
- **checkGlobalName(content, appName)**: Verifies global matches directory name
- **walkDOM(node, isListItem, violations)**: Recursively checks each node for violations

## Collaborators

- **html.Parser** (golang.org/x/net/html): Parses viewdef HTML into DOM tree
- **regexp**: Extracts patterns from Lua code
- **os/filepath**: Reads app files from disk

## Sequences

- seq-audit.md
