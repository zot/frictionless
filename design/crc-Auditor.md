# Auditor

**Source Spec:** specs/ui-audit.md
**Requirements:** R23, R24, R25, R26, R27, R28, R29, R30, R31, R32, R33, R34, R35, R36, R37, R38

Analyzes frictionless apps for code quality violations.

## Knows

- frameworkMethods: Methods never flagged as dead (`new`, `mutate`)
- externalMethods: Methods called by Claude, flagged as warnings (`addAgentMessage`, `updateRequirements`, `onAppProgress`, `onAppUpdated`)
- buttonElements: Elements where `ui-action` is valid (`button`, `sl-button`, `sl-icon-button`)
- operatorChars: Characters invalid in binding paths (`!`, `=`, `&`, `|`, `+`, `-`)
- namespaceAttrs: Attributes excluded from path checks (`ui-namespace` - viewdef namespace identifier)
- factoryFunctions: Map of local function names that create prototype methods (detected dynamically)
- calledFactories: Set of factory functions called at outer scope (outside any function definition)

## Does

- **AuditApp(baseDir, appName)**: Orchestrates full audit, returns AuditResult
- **analyzeLua(content, appName)**: Extracts method definitions/calls, detects factory functions, checks reloading guard and global name
- **analyzeViewdef(path, content, isListItem)**: Parses HTML, walks DOM for violations
- **extractMethodDefs(content)**: Regex extracts `function Type:method(` patterns
- **extractLuaCalls(content)**: Regex extracts `:method(` call patterns
- **extractViewdefCalls(attrs)**: Extracts `method()` from ui-* attribute values
- **detectFactoryFunctions(content)**: Identifies local functions that assign to `proto[...]` parameter
- **detectOuterScopeCalls(content, factoryFunctions)**: Finds factory function calls at outer scope (not inside function bodies)
- **findDeadMethods(defs, luaCalls, viewdefCalls, calledFactories)**: Cross-references to find unused methods, excluding factory-created methods
- **findMissingMethods(defs, viewdefCalls)**: Finds viewdef calls that don't match any Lua method definition
- **checkReloadingGuard(content)**: Verifies instance creation is guarded
- **checkGlobalName(content, appName)**: Verifies global matches directory name
- **walkDOM(node, isListItem, violations)**: Recursively checks each node for violations

## Collaborators

- **html.Parser** (golang.org/x/net/html): Parses viewdef HTML into DOM tree
- **regexp**: Extracts patterns from Lua code
- **os/filepath**: Reads app files from disk

## Sequences

- seq-audit.md
