---
name: write-file
description: Write content to a file with path validation. Use when you need to create or overwrite files in .ui/ or .claude/ directories.
---

# Write File Skill

Safely write content to files with guardrails.

## Usage

Run the script with the file path, piping content via stdin:

```bash
cat << 'EOF' | .claude/skills/write-file/scripts/write-file.sh /absolute/path/to/file
Your content here
Multiple lines work
EOF
```

## Guardrails

The script validates:
- **Path must be absolute** (starts with `/`)
- **Path must be in allowed directories**:
  - `.ui/` - UI app files
  - `.claude/` - Claude config files
  - `/tmp/` - Temporary files
- Paths outside these directories are rejected

## Examples

### Write a Lua file

```bash
cat << 'EOF' | .claude/skills/write-file/scripts/write-file.sh /home/user/project/.ui/apps/myapp/app.lua
-- My Lua app
MyApp = { type = "MyApp" }
MyApp.__index = MyApp

function MyApp:new()
    return setmetatable({ title = "Hello" }, self)
end

app = MyApp:new()
mcp.value = app
EOF
```

### Write an HTML viewdef

```bash
cat << 'EOF' | .claude/skills/write-file/scripts/write-file.sh /home/user/project/.ui/apps/myapp/viewdefs/MyApp.DEFAULT.html
<template>
  <div>
    <h1 ui-value="title"></h1>
  </div>
</template>
EOF
```

## Error Messages

- `Error: No file path provided` - Missing path argument
- `Error: Path must be absolute` - Path doesn't start with `/`
- `Error: Path not in allowed directories` - Tried to write outside `.ui/`, `.claude/`, or `/tmp/`
