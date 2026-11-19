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
| `mcp linkapp add\|remove\|list [APP]` | Activate, deactivate or list apps |
| `mcp checkpoint CMD APP [MSG]` | Manage app checkpoints |
| `mcp audit APP` | Run code quality audit |
| `mcp patterns` | List available patterns |
| `mcp theme list\|classes\|audit` | Theme management |

### Checkpoint Command

`mcp checkpoint CMD APP [ARG]` makes sure fossil is installed (see Fossil Integration),
then hands off to `frictionless checkpoint CMD APP [ARG]`, which does the work; the
commands and their behaviour are in `specs/checkpoint.md`. The script adds nothing to
the binary's output or exit status.

### Event Command Behavior

The `event` command:
1. Kills any previous event watcher (tracked via `.eventpid` file)
2. Long-polls `/wait` endpoint with 120s timeout
3. Returns JSON array of events on success
4. Returns empty output on timeout (exit 0)
5. Exits non-zero on server error

### Fossil Integration

Checkpoints use fossil SCM. Before handing a checkpoint command to the binary, the
script:
1. Auto-downloads fossil to `~/.claude/bin/fossil` if missing
2. Detects platform (Linux x86_64, macOS arm64/x86_64)
3. Downloads appropriate binary from fossil-scm.org

### MCP Guard

When `FRICTIONLESS_MCP=1` environment variable is set, the `run` command exits immediately. This prevents infinite recursion when Lua code calls shell commands.

### Linkapp Command

`mcp linkapp add|remove|list [APP]` hands off to `frictionless linkapp`, which does the
work: activation, deactivation and listing, on Unix and on Windows, are in
`specs/applink.md`. There is no separate `linkapp` script.
