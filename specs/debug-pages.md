# Debug Pages: Static HTML with Click-to-Copy

## Background

The MCP server has debug pages (`/variables` and `/state`) for inspecting runtime state. On the UI port, `/variables` already serves a static HTML file from `.ui/html/variables.html` that fetches JSON from the session API. On the MCP port, both `/variables` and `/state` still embed full HTML in Go code — this means changes require rebuilding the binary.

## Goals

1. **Remove old MCP port embedded HTML** for `/variables` and `/state`. The static `variables.html` on the UI port supersedes the old Shoelace tree version. Remove `handleVariables`, `handleState`, `renderVariableTree`, `renderVariableNode`, and related helpers from Go code. Keep `getDebugVariables()` in `resources.go` — the static variable browser fetches JSON from the UI server's `/{sessionId}/variables.json` route which depends on it.

2. **Redirect MCP port `/variables` and `/state`** to the UI port equivalents so existing bookmarks/scripts still work.

3. **Add click-to-copy** to the variable browser (`variables.html`). Clicking any cell except the Path column copies the full JSON for that variable's row to the clipboard. A brief toast confirms the copy.

## Variable Browser Click-to-Copy

- Clicking any cell other than the Path column copies the entire variable's JSON object to the clipboard
- A small toast notification appears briefly ("Copied") and fades away
- Path cells are excluded because they're used for tree expand/collapse interaction

## Origin

`variables.html` was originally an embedded Go string in ui-engine's server code. It has been extracted to a standalone file and now lives in frictionless at `install/html/variables.html`. The ui-engine still has the embedded version, which can be used as a reference for syncing if needed. The `state.html` is new and lives only in frictionless.

## Environment

- Language: Go (server), HTML/CSS/JS (debug pages)
- The `install/html/` directory is bundled into the Go binary at build time
- Files are installed to `.ui/html/` on `frictionless install` or auto-install
