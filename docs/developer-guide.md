# Frictionless Developer Guide

A guide for developers building apps with Frictionless.

<!-- **Traceability:** design/design.md, design/crc-*.md, design/seq-*.md, install/resources/*.md -->

## Architecture Overview

Frictionless uses a **Server-Side UI** architecture:

```
┌─────────────────────────────────────────────────────────────┐
│  AI Agent (Claude)                                          │
│                                                             │
│  Decides WHEN to use UI for user communication              │
│  Instructs UI Agent on WHAT to build                        │
│  Receives notifications from user interactions              │
└─────────────────────┬───────────────────────────────────────┘
                      │ MCP tools + resources
                      ▼
┌─────────────────────────────────────────────────────────────┐
│  UI MCP Server (Go)                                         │
│                                                             │
│  Lifecycle:          auto-starts, ui_configure (optional)   │
│  Code execution:     ui_run(lua_code)                       │
│  Browser:            ui_open_browser()                      │
│                                                             │
│  Resources:          ui://reference, ui://lua, ui://mcp     │
│  State:              ui://state                             │
└─────────────────────┬───────────────────────────────────────┘
                      │ HTTP + WebSocket
                      ▼
┌─────────────────────────────────────────────────────────────┐
│  Browser UI                                                 │
│                                                             │
│  Renders viewdefs, binds to Lua state                       │
│  User interactions → mcp.pushState() → AI Agent             │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### MCPServer (crc-MCPServer.md)

The main server component that:

- Manages HTTP/WebSocket connections
- Handles session lifecycle
- Executes Lua code in session context
- Provides the `/wait` endpoint for event polling

### MCPTool (crc-MCPTool.md)

MCP tool implementations:

| Tool | Purpose |
|------|---------|
| `ui_status` | Get server state, URL, and connection count |
| `ui_run` | Execute Lua code in session context |
| `ui_display` | Load and display an app by name |
| `ui_open_browser` | Open browser to session (with conserve mode) |
| `ui_configure` | Reconfigure with different base directory |
| `ui_install` | Install/update bundled skills and resources |
| `ui_audit` | Check app code for common issues |

### MCPResource (crc-MCPResource.md)

MCP resource providers:

| Resource | Content |
|----------|---------|
| `ui://reference` | Platform reference guide |
| `ui://lua` | Lua API and patterns |
| `ui://viewdefs` | Viewdef syntax reference |
| `ui://mcp` | Agent workflow guide |
| `ui://state` | Live session state (JSON) |

### Auditor (crc-Auditor.md)

Code quality checker that detects:

- Dead methods (defined but never called)
- Missing `session.reloading` guard
- Global variable naming mismatches
- Viewdef issues (`<style>` in list-items, `item.` prefix errors, etc.)

## Lua Programming

### Defining Classes

Use `session:prototype()` for hot-loadable classes:

```lua
-- 1. Declare app prototype (serves as namespace)
MyApp = session:prototype("MyApp", {
    items = EMPTY,
    name = ""
})

-- 2. Nested prototypes use dotted names
MyApp.Item = session:prototype("MyApp.Item", { name = "" })
local Item = MyApp.Item  -- local shortcut

function MyApp:new(instance)
    instance = session:create(MyApp, instance)
    instance.items = instance.items or {}
    return instance
end

-- 3. Guard instance creation (idempotent)
if not session.reloading then
    myApp = MyApp:new()
end
```

### Global Objects

| Object | Purpose |
|--------|---------|
| `Object` | Base prototype with `tostring()` method |
| `session` | Session services: `prototype()`, `create()`, `reloading` |
| `mcp` | Agent communication: `pushState()`, `status()`, `display()` |
| `EMPTY` | Marker for optional fields that start nil |

### The `mcp` Global

```lua
-- Properties
mcp.type          -- Always "MCP"
mcp.value         -- Current app value

-- Methods
mcp.pushState(event)           -- Queue event for agent
mcp:pollingEvents()            -- Check if agent is polling
mcp:app(appName)               -- Get app by name
mcp:display(appName)           -- Display app in browser
mcp:status()                   -- Get server status
mcp:appProgress(name, pct, stage)  -- Report build progress
mcp:appUpdated(name)           -- Trigger dashboard rescan
```

### Schema Migrations

Use `mutate()` when adding or changing fields:

```lua
MyApp = session:prototype("MyApp", {
    items = EMPTY,
    newField = EMPTY  -- NEW field
})

function MyApp:mutate()
    if self.newField == nil then
        self.newField = {}
    end
end
```

### Change Detection

Changes to Lua tables are automatically detected and pushed to the browser:

```lua
function MyApp:clear()
    self.name = ""       -- Automatically synced
    self.items = {}      -- Automatically synced
end
```

## Viewdef Development

### Template Structure

```html
<template>
  <style>
    .my-app { padding: 1rem; }
    .hidden { display: none !important; }
  </style>
  <div class="my-app">
    <sl-input ui-value="name" label="Name"></sl-input>
    <sl-button ui-action="save()">Save</sl-button>
  </div>
</template>
```

### Binding Attributes

| Attribute | Purpose | Example |
|:----------|:--------|:--------|
| `ui-value` | Bind value/text | `<sl-input ui-value="name">` |
| `ui-action` | Button click handler | `<sl-button ui-action="save()">` |
| `ui-event-click` | Click on any element | `<div ui-event-click="toggle()">` |
| `ui-event-*` | Any event | `<sl-select ui-event-sl-change="onSelect()">` |
| `ui-view` | Render child object | `<div ui-view="selected">` |
| `ui-attr-*` | HTML attribute | `<sl-alert ui-attr-open="hasError">` |
| `ui-class-*` | CSS class toggle | `<div ui-class-active="isActive">` |
| `ui-style-*` | CSS style | `<div ui-style-color="textColor">` |

### Lists

Use `ui-view` with `wrapper=lua.ViewList`:

```html
<div ui-view="items?wrapper=lua.ViewList"></div>
```

Create a list-item viewdef for your type:

```html
<!-- MyApp.Item.list-item.html -->
<template>
  <div class="item">
    <span ui-value="name"></span>
    <sl-button ui-action="delete()">Delete</sl-button>
  </div>
</template>
```

### Path Syntax

```
property           → self.property
nested.path        → self.nested.path
method()           → self:method()
items[0]           → self.items[1] (Lua is 1-indexed)
```

**No operators in paths.** Use Lua methods for negation:

```lua
function MyApp:notEditing() return not self.editing end
```

## Server Lifecycle

```
STARTUP ──auto-configure──► RUNNING ◄──ui_configure──┐
                               │                      │
                               └──────────────────────┘
                                  (reconfigure)
```

### Lua Loading Sequence

1. ui-engine loads `main.lua` (mcp global does NOT exist yet)
2. `setupMCPGlobal()` creates the `mcp` global with core methods
3. `loadMCPLua()` loads `{base_dir}/lua/mcp.lua` if it exists
4. `loadAppInitFiles()` scans `{base_dir}/apps/*/` and loads each `init.lua`

### HTTP Endpoints

Debug and inspect runtime state:

- `GET /wait` — Long-poll for `mcp.pushState()` events
- `GET /variables` — Interactive variable tree view
- `GET /state` — Current session state JSON

Tool API (for spawned agents):

- `GET /api/ui_status` — Get server status
- `POST /api/ui_run` — Execute Lua code
- `POST /api/ui_display` — Load and display an app
- `POST /api/ui_audit` — Audit app for code quality

## Build System

### Development

```bash
go build                      # Quick build (no bundled files)
make build                    # Full build with bundled files
./build/frictionless mcp      # Run MCP server
```

### Testing

```bash
go test ./...                 # Run all tests
make build && go test ./...   # Tests requiring bundled files
```

### Release

```bash
make release                  # Build for all platforms
```

Output in `release/` directory:
- `frictionless-linux-amd64`
- `frictionless-linux-arm64`
- `frictionless-darwin-amd64`
- `frictionless-darwin-arm64`
- `frictionless-windows-amd64.exe`

### Versioning

Source of truth: `README.md` (`**Version: X.Y.Z**`)

To create a release:
1. Update version in `README.md` and `install/README.md`
2. Commit: `git commit -am "Release vX.Y.Z"`
3. Tag: `git tag vX.Y.Z`
4. Build: `make release`
5. Push: `git push && git push --tags`
6. Create GitHub release

## Best Practices

### App Structure

```
.ui/apps/<app>/
├── app.lua              # Lua classes and logic
├── viewdefs/            # HTML templates
│   ├── MyApp.DEFAULT.html
│   └── MyApp.Item.list-item.html
├── requirements.md      # Human-readable requirements
├── design.md            # UI layout spec
└── README.md            # Events, state, methods
```

### Preventing Drift

During iterative changes, features can accidentally disappear:

1. **Before modifying** — Read `design.md`
2. **Update spec first** — Modify layout/components in spec
3. **Then update code** — Change viewdef and Lua to match
4. **Verify** — Ensure implementation matches spec

### Hot-Loading Atomic Writes

When adding fields with `mutate()`, both must arrive in a single hot-load:

```bash
# 1. Copy to temp file (not watched)
cp app.lua app.lua.tmp

# 2. Make ALL changes to temp file
# 3. Atomic replace
mv app.lua.tmp app.lua
```

### Viewport Fitting

Apps should fit within the viewport without page scroll:

```css
html, body {
  margin: 0;
  padding: 0;
  overflow: hidden;
}
.my-app {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}
.scrollable-area {
  flex: 1;
  min-height: 0;     /* CRITICAL */
  overflow-y: auto;
}
```

## Sequence Diagrams

Key sequences are documented in `design/seq-*.md`:

- **seq-mcp-lifecycle.md** — Server startup and shutdown
- **seq-mcp-create-session.md** — Browser connection
- **seq-mcp-run.md** — Lua code execution
- **seq-mcp-receive-event.md** — Event handling
- **seq-mcp-state-wait.md** — Event polling
- **seq-audit.md** — Code auditing

## Gaps and Known Issues

See `design/design.md` Gaps section for:

- Test coverage gaps
- Documentation needs
- Technical debt
