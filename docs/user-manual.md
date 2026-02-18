# Frictionless User Manual

A dynamic personal software ecosystem for Claude. Build your own Claude apps or download them.

<!-- **Traceability:** design/design.md, install/resources/reference.md, install/resources/mcp.md -->

## Getting Started

### Installation

Paste this into Claude Code:

```
Install using github zot/frictionless readme
```

Or install manually:

```bash
# Download (replace OS/ARCH: linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64)
mkdir -p ~/.claude/bin
curl -L https://github.com/zot/frictionless/releases/latest/download/frictionless-linux-amd64 -o ~/.claude/bin/frictionless
chmod +x ~/.claude/bin/frictionless

# Add to Claude Code
claude mcp add frictionless -- ~/.claude/bin/frictionless mcp

# Initialize the project
~/.claude/bin/frictionless install
```

### Building Your First App the Frictionless Way

```
/ui show
```

and use the [Frictionless console](../install/resources/intro.md).

## Building Your First App the Hard Way With the CLI (hit tab if you can't spell "thorough")

```
/ui-thorough make a `contacts` app with search and inline editing
/ui show contacts
```

## Core Concepts

### What is Frictionless?

Frictionless eliminates the complexity that eats tokens when building UIs:

- **No API layer** — no endpoints, no serialization, no DTOs
- **No frontend code** — just HTML templates with declarative bindings
- **No sync wiring** — change backend data, UI updates automatically

Claude writes your app logic and skips everything else.

### Directory Structure

```
.ui/
├── apps/           # App source of truth
│   └── <app>/          # Each app has its own directory
│       ├── app.lua         # Lua classes and logic
│       ├── viewdefs/       # HTML templates
│       ├── design.md       # UI layout spec
│       └── README.md       # Events and methods
├── lua/            # Symlinks to apps/<app>/*.lua
├── viewdefs/       # Symlinks to apps/<app>/viewdefs/*
├── patterns/       # Reusable UI patterns
├── conventions/    # Layout, terminology, preferences
└── library/        # Proven implementations
```

### Hot-Loading

Edit files and see changes instantly — no server restart or manual refresh needed:

- **Lua files** (`.lua`) — Code re-executes, browser updates automatically
- **Viewdef files** (`.html`) — Templates reload, components re-render

## Using Apps

### Displaying Apps

Claude uses the `/ui` skill to display apps:

```
/ui show contacts
```

This loads the app and opens the browser.

### Interacting with Apps

Apps communicate with Claude through events:

1. You click a button or enter data in the UI
2. The app calls `mcp.pushState(event)` in Lua
3. Claude receives the event and responds
4. Claude updates the UI by modifying Lua state

### Event-Driven Communication

```
┌─────────┐    click/type    ┌─────────┐   mcp.pushState()   ┌─────────┐
│  User   │ ◄──────────────► │   UI    │ ──────────────────► │  Claude │
└─────────┘   Lua responds   └─────────┘                     └────┬────┘
     ▲                                                            │
     │                       ┌─────────┐      ui_run()            │
     └───────────────────────│ Browser │ ◄────────────────────────┘
           sees changes      └─────────┘   updates state
```

**Important:** The event loop must be running for Claude to hear your app. Type `/ui events` to start it. Without this, button clicks are silently ignored.

## Standalone Mode

Run frictionless independently for development or testing:

```bash
frictionless serve --port 8000
frictionless serve --port 8000 --dir /path/to/ui-dir
```

The `--dir` option specifies the working directory for Lua scripts, viewdefs, and apps. Defaults to `.ui`.

## Bundling Custom Binaries

Create custom binaries with your site embedded:

```bash
frictionless bundle site/ -o my-ui-dir   # Create bundled binary
frictionless ls                          # List bundled files
frictionless cat index.html              # Show file contents
frictionless cp '*.lua' scripts/         # Copy matching files
frictionless extract output/             # Extract all bundled files
```

## Use Cases

### Quality of Life

- **Claude life** — UIs for common Claude tasks
- **UNIX life** — UIs for UNIX tools

### Life Beyond Code

- Expense tracking
- Habit building
- Project planning

### Dashboards

Surface information at a glance with real-time updates.

### Prototype Production Apps

Build functional wireframes at a fraction of the tokens.

## Documentation

Documentation is installed to `.ui/resources/` by default:

- **[Platform Reference](resources/reference.md)** — Architecture, tools, quick start
- **[Viewdef Syntax](resources/viewdefs.md)** — HTML template bindings
- **[Lua API](resources/lua.md)** — Class patterns and globals
- **[Agent Workflow](resources/mcp.md)** — Best practices for AI agents

## Troubleshooting

### Check Logs

Read `.ui/log/lua.log` for Lua errors during development.

### Browser Not Opening

If the browser doesn't open automatically, navigate to the URL shown in `ui_status()` output.

### Hot-Loading Not Working

Ensure your Lua code follows the hot-loading pattern:

```lua
-- Guard instance creation
if not session.reloading then
    myApp = MyApp:new()
end
```

## Future Directions

### App Permissions

Lua apps have filesystem access via `io.open`. Planned guardrails include:

- Apps declare required permissions in a manifest
- Permissions scoped to specific paths
- User approves permissions on first run
- Sandbox enforcement in the Lua environment
