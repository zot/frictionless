# MCPScript

**Requirements:** R54, R55, R56, R57, R58, R59, R60, R61, R62, R63, R64, R65, R66, R67, R68, R76, R77, R201

## Knows

- `dir`: Base directory (from script location)
- `port`: MCP server port (from `mcp-port` file)
- `prog`: Command name (first argument)

## Does

- **help**: Display usage with all commands
- **status**: GET `/api/ui_status`, return JSON
- **browser**: POST `/api/ui_open_browser`
- **display**: POST `/api/ui_display` with app name
- **run**: POST `/api/ui_run` with Lua code (guards with `FRICTIONLESS_MCP`)
- **event**: Long-poll `/wait`, track PID in `.eventpid`, kill previous watcher
- **state**: GET `/state`
- **variables**: GET `/variables`
- **progress**: Build Lua code for `mcp:appProgress()` + `addAgentThinking()`, POST to run
- **audit**: POST `/api/ui_audit` with app name
- **patterns**: Scan `patterns/*.md`, extract frontmatter description
- **theme**: Delegate to frictionless binary with `--dir`
- **checkpoint**: ensure fossil (download to `~/.claude/bin/fossil` for this platform when missing), then `exec` the frictionless binary's `checkpoint` with `--dir`, adding nothing to its output or exit status
- **linkapp**: Hand off to `frictionless linkapp`, as checkpoint does (see crc-AppLink.md)

## Collaborators

- **Checkpoint**: the Go package behind `frictionless checkpoint`
- **LinkappScript**: Handles app symlink management
- **FrictionlessBinary**: For theme and checkpoint commands
- **curl**: HTTP client for all API calls
- **jq**: JSON construction for POST bodies
