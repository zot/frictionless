# Frictionless

**Version: 0.6.0**

**A dynamic app ecosystem for Claude.** Build your own Claude apps or download them:

Build and modify apps while they run. No recompiles. No restarts.

- **Quality of life** — tame complex tasks with forms and buttons
  - **Claude life** — UIs for common Claude tasks
  - **UNIX life** — UIs for UNIX tools
- **Life beyond code** — expense tracking, habit building, project planning
- **Dashboards** — surface information at a glance
- **Prototype production apps** — build functional wireframes at a fraction of the tokens

## How It Works

Frictionless uses [ui-engine](https://github.com/zot/ui-engine) to eliminate complexity that eats tokens:

- **No API layer** — no endpoints, no serialization, no DTOs
- **No frontend code** — just HTML templates with declarative bindings
- **No sync wiring** — change backend data, UI updates automatically—no code to detect or push changes

Claude writes your app logic and skips everything else. See [overview](doc/OVERVIEW.md) for details.

## Usage

Once installed, Claude Code automatically starts frictionless when needed. The server uses `.ui` as the default working directory.

### Building UIs

Ask Claude to build a UI:

```
/ui-builder make a contacts app with search and inline editing
```

Or display an existing app:

```
/ui show
```

### Standalone Mode

Run frictionless independently for development or testing:

```bash
frictionless serve --port 8000 --mcp-port 8001
frictionless serve --port 8000 --mcp-port 8001 --dir /path/to/ui-dir
```

The `--dir` option specifies the working directory for Lua scripts, viewdefs, and apps. Defaults to `.ui`.
The `--mcp-port` is only needed if you want to connect it to Claude.

### Bundling

Create custom binaries with your site embedded:

```bash
frictionless bundle site/ -o my-ui-dir   # Create bundled binary
frictionless ls                          # List bundled files
frictionless cat index.html              # Show file contents
frictionless cp '*.lua' scripts/         # Copy matching files
frictionless extract output/             # Extract all bundled files to current directory
```

## Documentation (in .ui by default)

- **[Platform Reference](resources/reference.md)** — Architecture, tools, and quick start guide
- **[Viewdef Syntax](resources/viewdefs.md)** — HTML template bindings (`ui-*` attributes)
- **[Lua API](resources/lua.md)** — Class patterns and globals
- **[Agent Workflow](resources/mcp.md)** — Best practices for AI agents

## Installation

Paste this into Claude Code to install:

```
Install using github zot/frictionless readme
```

To install manually:

```bash
# Download (replace OS/ARCH: linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64)
mkdir -p {home}/.claude/bin
curl -L https://github.com/zot/frictionless/releases/latest/download/frictionless-linux-amd64 -o {home}/.claude/bin/frictionless
chmod +x {home}/.claude/bin/frictionless

# Add to Claude Code
claude mcp add frictionless -- {home}/.claude/bin/frictionless} mcp

# Initialize the project
{home}/.claude/bin/frictionless} install
```

## Future Directions

### App Permissions

Lua apps have filesystem access via `io.open`, which allows reading files without Claude Code permission prompts. While convenient, this could be misused to explore directories the user hasn't explicitly shared.

**Planned guardrails:**
- Apps declare required permissions in a manifest (e.g., `app.json`)
- Permissions scoped to specific paths: `.claude/`, project root, etc.
- User approves permissions on first run
- Sandbox enforcement in the Lua environment
