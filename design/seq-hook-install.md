# Sequence: Hook Installation

**Source Spec:** prompt-ui.md

## Participants

- User: Developer running CLI command
- HookCLI: CLI command handler
- OS: File system operations

## Sequence

```
+----+     +-------+     +--+
|User|     |HookCLI|     |OS|
+-+--+     +---+---+     +-++
  |            |           |
  | ui-mcp hooks install   |
  |----------->|           |
  |            |           |
  |            | Check .claude/ exists
  |            |---------->|
  |            |           |
  |            |   exists  |
  |            |<- - - - - |
  |            |           |
  |            | mkdir -p .claude/hooks
  |            |---------->|
  |            |           |
  |            | Write .claude/hooks/permission-ui.sh
  |            |---------->|
  |            |           |
  |            | chmod +x permission-ui.sh
  |            |---------->|
  |            |           |
  |            | Read .claude/settings.json
  |            |---------->|
  |            |           |
  |            | settings (or empty)
  |            |<- - - - - |
  |            |           |
  |            |---+       |
  |            |   | Merge hook config
  |            |<--+ into settings
  |            |           |
  |            | Write .claude/settings.json
  |            |---------->|
  |            |           |
  | Success:   |           |
  | Hook installed         |
  |<- - - - - -|           |
+-+--+     +---+---+     +-++
|User|     |HookCLI|     |OS|
+----+     +-------+     +--+
```

## Notes

### Settings.json Merge

If settings.json exists, merge the hook config:
```json
{
  "existing": "settings",
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

If PermissionRequest hooks already exist, append to array.

### Error Conditions

- **.claude/ doesn't exist**: Create it
- **settings.json malformed**: Report error, don't overwrite
- **No write permission**: Report error

### Uninstall Flow

1. Read settings.json
2. Remove permission-ui.sh entry from PermissionRequest array
3. Write settings.json
4. Optionally delete script (with --delete-script flag)
