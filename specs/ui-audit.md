# ui_audit Tool Specification

**Language:** Go
**Environment:** frictionless MCP server

## Purpose

Provide an automated code quality checker that agents can use to audit frictionless apps for common violations. Background agents access it via HTTP API (curl), foreground agents via MCP tool.

## Motivation

The ui-builder skill manually checks for violations during the audit phase. Automating these checks allows:
- Consistent enforcement across all builds
- Background agents to self-audit without MCP tool access
- Faster feedback during iterative development

## Access Methods

### MCP Tool
Foreground agents call `ui_audit` with an app name.

### HTTP API
Background agents POST to `/api/ui_audit` with JSON body `{"name": "app-name"}`.

## Checks

### Lua Checks (app.lua)

**Dead methods**: Methods defined but never called from Lua code or viewdefs. Framework methods (`new`, `mutate`) are excluded. Methods intended for Claude to call via `ui_run` (like `addAgentMessage`, `onAppProgress`) are flagged as warnings, not violations. Methods created by factory functions called at the outer scope are not dead, since the factory call itself represents intentional method creation.

**Factory method pattern**: Factory functions are local functions that dynamically add methods to a prototype. When a factory function is called at the outer scope (not inside another function), any methods it creates are considered "used" and not flagged as dead. This pattern is common for generating similar methods:

```lua
local function makeCollapsible(proto, fieldName)
    proto["toggle" .. fieldName] = function(self) ... end
    proto[fieldName .. "Hidden"] = function(self) ... end
    proto[fieldName .. "Icon"] = function(self) ... end
end

makeCollapsible(AppInfo, "KnownIssues")  -- outer scope call
makeCollapsible(AppInfo, "FixedIssues")  -- outer scope call
```

The audit tool detects this pattern by:
1. Identifying local functions that assign to `proto[...]` (factory functions)
2. Tracking calls to these factory functions at the outer scope
3. Marking methods created by called factories as "used"

**Missing reloading guard**: Instance creation (`= Type:new()`) without being wrapped in `if not session.reloading then`. This causes duplicate instances on hot-reload.

**Global name mismatch**: The global variable name should match the app directory name (e.g., `apps` directory → `apps` global, not `appsApp`).

### Viewdef Checks (*.html files)

**Malformed HTML**: The HTML cannot be parsed. Indicates syntax errors that will break the UI.

**Style in list-item**: A `<style>` tag appears in a `*.list-item.html` file. Styles get duplicated for each list item and should be in the top-level viewdef only.

**item. prefix in list-item**: Bindings in list-item viewdefs use `item.name` instead of just `name`. The item IS the context in its own viewdef.

**ui-action on non-button**: The `ui-action` attribute only works on button elements (`button`, `sl-button`, `sl-icon-button`). Other elements should use `ui-event-click`.

**Wrong hidden syntax**: Using `ui-class="hidden:condition"` instead of `ui-class-hidden="condition"`.

**ui-value on checkbox**: Using `ui-value` on `sl-checkbox` or `sl-switch` renders the boolean as text. Should use `ui-attr-checked` instead.

**ui-value on badge**: Using `ui-value` on `sl-badge` is not supported. Use a `<span ui-value="..."></span>` inside the badge instead.

**Operators in paths**: Binding paths contain operators (`!`, `==`, `&&`, `||`, `+`, `-`). Paths don't support operators - use Lua methods instead. Excludes `ui-namespace` which is a viewdef namespace identifier (e.g., `list-item`), not a binding path.

**Non-empty method args**: Method calls in paths can only be `method()` or `method(_)`. Using `method(arg)` with any other content is not allowed - the underscore `_` is a special placeholder for the binding value.

**Invalid path syntax**: As a final validation, all ui-* attribute values (except `ui-namespace`) are checked against the path grammar. Valid path syntax is:

```
prefix   ::= ident | "[" ( ident | number ) "]" | ident "()"
suffix   ::= prefix | ident "(_)"
property ::= ident [ "=" text ]
path     ::= { prefix "." } suffix [ "?" property { "&" property } ]
```

Examples of valid paths: `name`, `getName()`, `parent.child`, `items[0].name`, `setValue(_)`, `items?wrapper=ViewList`, `search?keypress`
Examples of invalid paths: `getValue(x)`, `name[`, `foo..bar`

**Missing Lua method**: A viewdef binding references a method that doesn't exist in app.lua. For example, `ui-action="doSomething()"` where `doSomething` is not defined on any prototype. This catches typos and forgotten implementations.

## Output

JSON response with:
- `app`: The app name that was audited
- `violations`: Array of issues that must be fixed
- `warnings`: Array of potential issues (like external methods)
- `reminders`: Array of behavioral checks that cannot be automated (agent should verify manually)
- `summary`: Counts of total methods, dead methods, and viewdef violations

Each violation/warning includes:
- `type`: The violation type identifier
- `location`: File path and optionally line number
- `detail`: Human-readable description

## Example Response

```json
{
  "app": "my-app",
  "violations": [
    {"type": "dead_method", "location": "app.lua", "detail": "MyType:unusedMethod"}
  ],
  "warnings": [
    {"type": "external_method", "location": "app.lua", "detail": "MyApp:onProgress (called by Claude)"}
  ],
  "reminders": [
    "Check for missing `min-height: 0` on scrollable flex children",
    "Check that Cancel buttons revert changes",
    "Check for slow function bindings that need caching (e.g., methods calling io.popen/os.execute)"
  ],
  "summary": {
    "total_methods": 25,
    "dead_methods": 1,
    "viewdef_violations": 0
  }
}
```

## Integration

### ui-builder Skill

The ui-builder skill's audit phase (step 6) should call `ui_audit` instead of manually checking for violations. The skill should:

1. Call `ui_audit(name)` after writing files
2. If violations are returned, fix them before marking complete
3. Report warnings to the user but don't block completion

This replaces the manual violation checklist in the skill with automated enforcement.

## Reminders

Reminders are behavioral checks that cannot be automated. The audit tool includes them in every response to prompt agents to verify manually:

- **min-height: 0**: Scrollable flex children need `min-height: 0` to allow shrinking below content size
- **Cancel buttons revert changes**: Cancel buttons should restore original state, not just close dialogs
- **Slow function bindings**: Methods called from viewdef bindings that use `io.popen`, `os.execute`, or similar slow operations, or that build large lists should cache results to avoid UI lag
