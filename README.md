# ui-mcp

**Version: 0.6.0**

**An app ecosystem for Claude.** Build your own Claude apps or download them:

- **Dashboards** — surface information at a glance
- **Command frontends** — tame complex UNIX tools with forms and buttons
- **Workflow tools** — common Claude usage patterns as clickable actions
- **Life beyond code** — expense tracking, habit building, project planning
- **Prototype production apps** — build functional wireframes at a fraction of the tokens

Build and modify apps while they run. No restarts, no rebuilds, no wait.

## How It Works

UI-MCP uses [ui-engine](https://github.com/zot/ui-engine) to eliminate complexity that eats tokens:

- **No API layer** — no endpoints, no serialization, no DTOs
- **No frontend code** — just HTML templates with declarative bindings
- **No sync wiring** — change backend data, UI updates automatically—no code to detect or push changes

Claude writes your app logic and skips everything else. See [Architecture](doc/OVERVIEW.md) for details.

## Usage

Once installed, Claude Code automatically starts ui-mcp when needed. The server uses `.ui` as the default working directory.

### Building UIs

Ask Claude to build a UI:

```
/ui-builder make a contacts app with search and inline editing
```

Or display an existing app:

```
/ui show contacts
```

### Standalone Mode

Run ui-mcp independently for development or testing:

```bash
ui-mcp serve --port 8000
ui-mcp serve --port 8000 --dir /path/to/ui-dir
```

The `--dir` option specifies the working directory for Lua scripts, viewdefs, and apps. Defaults to `.ui`.

### Bundling

Create custom binaries with your site embedded:

```bash
ui-mcp bundle site/ -o my-ui-dir   # Create bundled binary
ui-mcp ls                          # List bundled files
ui-mcp cat index.html              # Show file contents
ui-mcp cp '*.lua' scripts/         # Copy matching files
ui-mcp extract output/             # Extract all bundled files to current directory
```

## Documentation (in .ui by default)

- **[Platform Reference](resources/reference.md)** — Architecture, tools, and quick start guide
- **[Viewdef Syntax](resources/viewdefs.md)** — HTML template bindings (`ui-*` attributes)
- **[Lua API](resources/lua.md)** — Class patterns and globals
- **[Agent Workflow](resources/mcp.md)** — Best practices for AI agents

## Installation

Paste this into Claude Code to install:

```
Install using github zot/ui-mcp readme
```

To install manually:

```bash
# Download (replace OS/ARCH: linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64)
mkdir -p {home}/.claude/bin
curl -L https://github.com/zot/ui-mcp/releases/latest/download/ui-mcp-linux-amd64 -o {home}/.claude/bin/ui-mcp
chmod +x {home}/.claude/bin/ui-mcp

# Add to Claude Code
claude mcp add ui-mcp -- {home}/.claude/bin/ui-mcp} mcp

# Initialize the project
{home}/.claude/bin/ui-mcp} install
```

## Future Directions

### App Permissions

Lua apps have filesystem access via `io.open`, which allows reading files without Claude Code permission prompts. While convenient, this could be misused to explore directories the user hasn't explicitly shared.

**Planned guardrails:**
- Apps declare required permissions in a manifest (e.g., `app.json`)
- Permissions scoped to specific paths: `.claude/`, project root, etc.
- User approves permissions on first run
- Sandbox enforcement in the Lua environment
