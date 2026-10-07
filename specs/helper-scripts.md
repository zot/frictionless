# Helper Scripts Specification

**Language:** Bash
**Environment:** Unix-like systems (Linux, macOS)

## Overview

Two shell scripts installed to `{base_dir}/` that provide CLI access to MCP server functionality. These scripts are the primary interface for Claude agents to interact with the UI system.

## Script: `mcp`

The main CLI interface for all MCP operations. Reads server port from `{base_dir}/mcp-port` file.

### Commands

| Command | Description |
|---------|-------------|
| `mcp --help` | Show usage help |
| `mcp status` | Get server status (JSON) |
| `mcp browser` | Open browser to UI session |
| `mcp display APP` | Display APP in the browser |
| `mcp run 'lua code'` | Execute Lua code in session |
| `mcp event` | Wait for next UI event (120s timeout) |
| `mcp state` | Get current session state (JSON) |
| `mcp variables` | Get current variable values |
| `mcp progress APP PERCENT STAGE` | Report build progress |
| `mcp linkapp add\|remove APP` | Activate or deactivate an app |
| `mcp checkpoint CMD APP [MSG]` | Manage app checkpoints |
| `mcp audit APP` | Run code quality audit |
| `mcp patterns` | List available patterns |
| `mcp theme list\|classes\|audit` | Theme management |

### Checkpoint Subcommands

| Command | Description |
|---------|-------------|
| `checkpoint save APP [MSG]` | Save current state (no-op if unchanged) |
| `checkpoint list APP` | List checkpoint history |
| `checkpoint rollback APP [N]` | Rollback to Nth checkpoint |
| `checkpoint diff APP [N]` | Show diff from checkpoint N |
| `checkpoint clear APP` | Clear checkpoints (reset to baseline) |
| `checkpoint baseline APP` | Set current state as baseline |
| `checkpoint count APP` | Return checkpoint count |
| `checkpoint update APP [MSG]` | Save update checkpoint (survives baseline) |
| `checkpoint local APP [MSG]` | Save local checkpoint (survives baseline) |

### Event Command Behavior

The `event` command:
1. Kills any previous event watcher (tracked via `.eventpid` file)
2. Long-polls `/wait` endpoint with 120s timeout
3. Returns JSON array of events on success
4. Returns empty output on timeout (exit 0)
5. Exits non-zero on server error

### Fossil Integration

Checkpoints use fossil SCM. The script:
1. Auto-downloads fossil to `~/.claude/bin/fossil` if missing
2. Detects platform (Linux x86_64, macOS arm64/x86_64)
3. Downloads appropriate binary from fossil-scm.org
4. Initializes per-app repository in `{app_dir}/checkpoint.fossil`

### MCP Guard

When `FRICTIONLESS_MCP=1` environment variable is set, the `run` command exits immediately. This prevents infinite recursion when Lua code calls shell commands.

## Script: `linkapp`

Activates and deactivates ui-engine apps. An app is active when
`{base_dir}/lua/{app}` links to its directory; that one link is the whole
on/off bit. ui-engine reads the app's entry file, modules and viewdefs
through it (see ui-engine's app-discovery spec).

### Commands

| Command | Description |
|---------|-------------|
| `linkapp add APP` | Activate an app |
| `linkapp remove APP` | Deactivate an app |
| `linkapp list` | List active apps |

### Directory Structure

```
{base_dir}/
├── apps/
│   └── {app}/
│       ├── app.lua           # Entry file, loaded as lua/{app}/app.lua
│       ├── json.lua          # Optional module files: require("{app}.json")
│       └── viewdefs/*.html
└── lua/
    └── {app} -> ../apps/{app}   # Activation link
```

### Link Behavior

**add:**
- Creates `{base_dir}/lua/` if missing
- Links the app directory to `lua/{app}`
- Removes the app's old-scheme links if present: `lua/{app}.lua`, and any
  `viewdefs/` symlink pointing into the app's `viewdefs/`

**remove:**
- Removes the `lua/{app}` link
- Removes the same old-scheme links as `add`

**list:**
- Scans `lua/` for directory links and reports app names
