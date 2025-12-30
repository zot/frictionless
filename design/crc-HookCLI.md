# HookCLI

**Source Spec:** prompt-ui.md

**Implementation:** cmd/ui-mcp/main.go (runHooks, installHook, uninstallHook, hookStatus)

## Responsibilities

### Knows
- hookScriptPath: Target path for hook script (.claude/hooks/permission-ui.sh)
- settingsPath: Claude settings file (.claude/settings.json)
- hookScript: Hook script content (embedded as constant in main.go)

### Does
- runHooks: Dispatch to install/uninstall/status subcommands
- installHook: Create hook script, make executable, update settings.json with PermissionRequest hook
- uninstallHook: Remove hook from settings.json
- hookStatus: Check if hook installed, script exists, server running

## Collaborators

- OS: File operations, JSON manipulation
- PermissionHook: The script content that gets installed

## Sequences

- seq-hook-install.md: CLI installs hook and updates settings

## Notes

CLI subcommands:
- `ui-mcp hooks install` - Full installation
- `ui-mcp hooks uninstall` - Remove from settings
- `ui-mcp hooks status` - Show current state

Settings.json format (with matcher for new Claude Code hook format):
```json
{
  "hooks": {
    "PermissionRequest": [
      {
        "matcher": {},
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/permission-ui.sh",
            "timeout": 120
          }
        ]
      }
    ]
  }
}
```
