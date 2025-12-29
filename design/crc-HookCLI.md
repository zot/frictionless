# HookCLI

**Source Spec:** prompt-ui.md

## Responsibilities

### Knows
- hookScriptPath: Target path for hook script (.claude/hooks/permission-ui.sh)
- settingsPath: Claude settings file (.claude/settings.json)
- embeddedScript: Hook script content (embedded in binary or generated)

### Does
- Install: Create hook script, make executable, update settings.json with PermissionRequest hook
- Uninstall: Remove hook from settings.json, optionally delete script
- Status: Check if hook installed, script exists, server running

## Collaborators

- OS: File operations, JSON manipulation
- PermissionHook: The script content that gets installed

## Sequences

- seq-hook-install.md: CLI installs hook and updates settings

## Notes

CLI subcommands:
- `ui-mcp hooks install` - Full installation
- `ui-mcp hooks uninstall [--delete-script]` - Remove from settings
- `ui-mcp hooks status` - Show current state

Settings.json modification:
```json
{
  "hooks": {
    "PermissionRequest": [
      {
        "type": "command",
        "command": ".claude/hooks/permission-ui.sh",
        "timeout": 120
      }
    ]
  }
}
```
