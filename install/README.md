# frictionless

**Version: 0.8.0**

An MCP server that enables AI agents to build interactive UIs for rich two-way communication with users.

## Benefits

- **Prototyping** — Agent and user collaborate on UI wireframes for production apps
- **Testing** — Create mock UIs for testing workflows
- **Stateful interaction** — Go beyond text-only conversations:
  - Collect structured input (forms, selections, ratings, file picks)
  - Present data with layout (lists, tables, comparisons)
  - Multi-step workflows (wizards, confirmations, progress tracking)
  - Real-time feedback loops (editing, previewing, validation)
- **Claude Apps** — Persistent UIs for interacting with Claude:
  - Launch panels with buttons for design, implement, analyze gaps
  - Project dashboards showing available commands, agents, skills
  - Status displays for background tasks and build progress

## Installation

Paste this into Claude Code to install:

```
Install from github zot/frictionless readme
```

To install manually:

```bash
# Download (replace OS/ARCH: linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64)
mkdir -p {home}/.claude/bin
curl -L https://github.com/zot/frictionless/releases/latest/download/frictionless-linux-amd64 -o {home}/.claude/bin/frictionless
chmod +x {home}/.claude/bin/frictionless

# Add to Claude Code
claude mcp add frictionless -- {home}/.claude/bin/frictionless} mcp
```

## Usage

Once installed, use `/ui` to start the frictionless server. The server uses `.ui` as the default working directory.

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

Run frictionless independently for development or testing:

```bash
frictionless serve --port 8000
frictionless serve --port 8000 --dir /path/to/ui-dir
```

The `--dir` option specifies the working directory for Lua scripts, viewdefs, and apps. Defaults to `.ui`.

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

## Future Directions

### App Permissions

Lua apps have filesystem access via `io.open`, which allows reading files without Claude Code permission prompts. While convenient, this could be misused to explore directories the user hasn't explicitly shared.

**Planned guardrails:**
- Apps declare required permissions in a manifest (e.g., `app.json`)
- Permissions scoped to specific paths: `.claude/`, project root, etc.
- User approves permissions on first run
- Sandbox enforcement in the Lua environment
