# ui_audit Tool

Analyzes an app for code quality violations.

## Usage

**MCP Tool:**
```
ui_audit(name: "app-name")
```

**HTTP API:**
```bash
.ui/mcp audit APP-NAME
```

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| name | string | Yes | App name to audit |

## Response

```json
{
  "app": "app-name",
  "violations": [
    {"type": "dead_method", "location": "app.lua", "detail": "Type:unusedMethod"}
  ],
  "warnings": [
    {"type": "external_method", "location": "app.lua", "detail": "Type:onProgress (called by Claude)"}
  ],
  "summary": {
    "total_methods": 25,
    "dead_methods": 1,
    "viewdef_violations": 0
  }
}
```

## Violation Types

### Lua Violations

| Type | Description |
|------|-------------|
| `dead_method` | Method defined but never called |
| `missing_reloading_guard` | Instance creation not wrapped in `if not session.reloading` |
| `global_name_mismatch` | Global variable doesn't match app directory name |

### Viewdef Violations

| Type | Description |
|------|-------------|
| `html_parse_error` | Malformed HTML in viewdef |
| `style_in_list_item` | `<style>` tag in list-item viewdef |
| `item_prefix` | Using `item.` prefix in list-item bindings |
| `ui_action_non_button` | `ui-action` on non-button element |
| `wrong_hidden_syntax` | Using `ui-class="hidden:..."` instead of `ui-class-hidden` |
| `ui_value_checkbox` | `ui-value` on sl-checkbox/sl-switch |
| `operator_in_path` | Operators in binding paths |

### Warnings

| Type | Description |
|------|-------------|
| `external_method` | Method called by Claude via ui_run, not from code |

## Example

```bash
# Audit the "app-console" app
.ui/mcp audit app-console
```
