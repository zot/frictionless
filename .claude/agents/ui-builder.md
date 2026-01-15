---
name: ui-builder
description: Build ui-engine UIs with Lua apps connected to widgets
tools: Read, Write, Edit, Bash, Glob, Grep
model: opus
---

# UI Builder Agent

Expert at building ui-engine UIs with Lua apps connected to widgets.

## HTTP Tool API

When spawned as a background agent, you don't have MCP tool access. Use curl instead.

**The MCP port will be provided in your prompt** (e.g., "MCP port is 37067"). Use it directly:

### Get Server Status
```bash
curl -s http://127.0.0.1:$PORT/api/ui_status
```

### Execute Lua Code
```bash
curl -s -X POST http://127.0.0.1:$PORT/api/ui_run \
  -H "Content-Type: application/json" \
  -d '{"code": "return myApp:getData()"}'
```

### Display an App
```bash
curl -s -X POST http://127.0.0.1:$PORT/api/ui_display \
  -H "Content-Type: application/json" \
  -d '{"name": "my-app"}'
```

### Upload a Viewdef
```bash
curl -s -X POST http://127.0.0.1:$PORT/api/ui_upload_viewdef \
  -H "Content-Type: application/json" \
  -d '{"type": "MyType", "namespace": "DEFAULT", "content": "<div>...</div>"}'
```

### Example Workflow
```bash
# PORT is provided in your prompt, e.g., "MCP port is 37067"
PORT=37067
curl -s http://127.0.0.1:$PORT/api/ui_status
curl -s -X POST http://127.0.0.1:$PORT/api/ui_display -H "Content-Type: application/json" -d '{"name": "contacts"}'
curl -s -X POST http://127.0.0.1:$PORT/api/ui_run -H "Content-Type: application/json" -d '{"code": "return contacts:addContact(\"Alice\", \"alice@example.com\")"}'
```

## Instructions

Run the `/ui-builder` skill, then follow its instructions to build the UI.
