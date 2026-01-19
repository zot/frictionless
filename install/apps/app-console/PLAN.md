# Apps Dashboard - Planning

## Purpose
Dashboard for discovering, launching, and tracking quality of frictionless apps.

## User Stories

### Discovery
- As a user, I want to see all available apps so I can know what's built
- As a user, I want to quickly understand what each app does

### Navigation
- As a user, I want to launch any app directly from the dashboard
- As a user, I want to return to the dashboard easily

### Quality Tracking
- As a developer, I want to see testing status at a glance
- As a developer, I want to identify which apps need attention
- As a developer, I want to see known bugs across all apps

## Claude Interactions

The apps dashboard is a **command center** for UI development with Claude.

### Chat Panel (always visible)
- User chats with Claude about apps
- Selected app provides context
- Claude can explain apps, suggest next steps, answer questions

### Action Buttons
- **Build** - Ask Claude to build the app (shown when no app.lua)
- **Open** - Launch the app in browser (shown when has app.lua)
- **Test** - Ask Claude to run ui-testing skill (shown when has app.lua)
- **Fix Issues** - Ask Claude to fix known issues (shown when has known issues)

### Create New App
- **"+" button** opens a panel with:
  - Name field (becomes directory name, kebab-case)
  - Description textarea (what the app should do)
  - Create button → sends to Claude to build

### Events to Claude
```json
{"app": "apps", "event": "chat", "text": "...", "context": "contacts"}
{"app": "apps", "event": "build_request", "target": "my-app"}
{"app": "apps", "event": "test_request", "target": "contacts"}
{"app": "apps", "event": "fix_request", "target": "contacts"}
{"app": "apps", "event": "create_app", "name": "my-app", "description": "..."}
```

### Build Progress (Claude → UI)
When building an app, Claude updates the UI with progress:
```lua
apps:addApp("my-app")                              -- app appears in list
apps:setBuildProgress("my-app", 10, "designing...")
apps:setBuildProgress("my-app", 40, "writing code...")
apps:setBuildProgress("my-app", 70, "creating viewdefs...")
apps:setBuildProgress("my-app", 90, "linking...")
apps:setBuildProgress("my-app", nil, nil)          -- done, rescan from disk
```

| Progress | Stage |
|----------|-------|
| 10% | designing |
| 40% | writing code |
| 70% | creating viewdefs |
| 90% | linking |
| nil | done |

## Decisions

1. **What counts as an "app"?** → Directories under `apps/` with `requirements.md` (apps can be unbuilt)

2. **Metadata to show:**
   - App name (directory name)
   - Description (first paragraph from requirements.md)
   - Test status (X/Y passing from TESTING.md)
   - Known issues count
   - Build progress (0-100, nil when not building)
   - Build stage (designing/writing code/creating viewdefs/linking, nil when not building)

3. **Detail view contents:**
   - App name as header
   - Description
   - Test checklist (read-only checkboxes)
   - Known issues list
   - Fixed issues list (collapsed by default)

## Layout

```
+------------------+-----------------------------+
| Apps        [+]  | contacts                    |
|------------------|  A contact manager with...  |
| > contacts 17/21 | [Open] [Test] [Fix Issues]  |
|   tasks    5/5   |-----------------------------|
|   apps     --    | Tests (17/21)               |
|                  | [x] Badge shows count       |
|                  | [ ] Delete removes contact  |
|                  |-----------------------------|
|                  | Known Issues (2)            |
|                  | 1. Status dropdown broken   |
+------------------+-----------------------------+
| Chat                                           |
| Agent: Which app would you like to work on?    |
| You: Test the contacts app                     |
| [____________________________________] [Send]  |
+------------------------------------------------+
```

### New App Panel (replaces details when + clicked)
```
+------------------+-----------------------------+
| Apps        [+]  | New App                     |
|------------------|                             |
| > contacts 17/21 | Name: [_______________]     |
|   tasks    5/5   |                             |
|   apps     --    | Description:                |
|                  | [                         ] |
|                  | [                         ] |
|                  |                             |
|                  | [Cancel]         [Create]   |
+------------------+-----------------------------+
```

## Notes

- TESTING.md format:
  - `- [x]` = passing
  - `- [ ]` = failing/untested
  - `### N.` under "Known Issues" = open bugs
  - `### N.` under "Fixed Issues" = resolved bugs

