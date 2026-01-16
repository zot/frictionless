# ui_audit Tool Specification

**Language:** Go
**Environment:** ui-mcp MCP server

## Purpose

Provide an automated code quality checker that agents can use to audit ui-mcp apps for common violations. Background agents access it via HTTP API (curl), foreground agents via MCP tool.

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

**Dead methods**: Methods defined but never called from Lua code or viewdefs. Framework methods (`new`, `mutate`) are excluded. Methods intended for Claude to call via `ui_run` (like `addAgentMessage`, `onAppProgress`) are flagged as warnings, not violations.

**Missing reloading guard**: Instance creation (`= Type:new()`) without being wrapped in `if not session.reloading then`. This causes duplicate instances on hot-reload.

**Global name mismatch**: The global variable name should match the app directory name (e.g., `apps` directory → `apps` global, not `appsApp`).

### Viewdef Checks (*.html files)

**Malformed HTML**: The HTML cannot be parsed. Indicates syntax errors that will break the UI.

**Style in list-item**: A `<style>` tag appears in a `*.list-item.html` file. Styles get duplicated for each list item and should be in the top-level viewdef only.

**item. prefix in list-item**: Bindings in list-item viewdefs use `item.name` instead of just `name`. The item IS the context in its own viewdef.

**ui-action on non-button**: The `ui-action` attribute only works on button elements (`button`, `sl-button`, `sl-icon-button`). Other elements should use `ui-event-click`.

**Wrong hidden syntax**: Using `ui-class="hidden:condition"` instead of `ui-class-hidden="condition"`.

**ui-value on checkbox**: Using `ui-value` on `sl-checkbox` or `sl-switch` renders the boolean as text. Should use `ui-attr-checked` instead.

**Operators in paths**: Binding paths contain operators (`!`, `==`, `&&`, `||`, `+`, `-`). Paths don't support operators - use Lua methods instead.

## Output

JSON response with:
- `app`: The app name that was audited
- `violations`: Array of issues that must be fixed
- `warnings`: Array of potential issues (like external methods)
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
