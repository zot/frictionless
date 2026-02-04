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
| `mcp linkapp add\|remove APP` | Manage app symlinks |
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

Manages symlinks for ui-engine apps. Creates links from `{base_dir}/lua/` and `{base_dir}/viewdefs/` to app source files.

### Commands

| Command | Description |
|---------|-------------|
| `linkapp add APP` | Create symlinks for app |
| `linkapp remove APP` | Remove symlinks for app |
| `linkapp list` | List linked apps |

### Directory Structure

```
{base_dir}/
├── apps/
│   └── {app}/
│       ├── app.lua
│       ├── json.lua          # Optional module files
│       └── viewdefs/*.html
├── lua/
│   ├── {app}.lua -> ../apps/{app}/app.lua   # Main app file for ui-engine
│   └── {app} -> ../apps/{app}               # Directory link for require("{app}.module")
└── viewdefs/
    └── {Type}.*.html -> ../apps/{app}/viewdefs/{Type}.*.html
```

### Link Behavior

**add:**
- Creates `{base_dir}/lua/` and `{base_dir}/viewdefs/` if missing
- Links `app.lua` to `lua/{app}.lua` (main app file for ui-engine)
- Links app directory to `lua/{app}` (enables `require("{app}.module")`)
- Links all viewdef HTML files individually

**remove:**
- Removes `lua/{app}.lua` symlink
- Removes `lua/{app}` directory symlink
- Scans viewdefs/ for symlinks pointing to app's viewdefs/ and removes them

**list:**
- Scans lua/ for `.lua` symlinks and reports app names
