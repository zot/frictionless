# Frictionless UI - User Guide

Build interactive, fully hot-loadable apps by chatting with Claude. Describe what you want, and Claude generates working Lua apps with HTML templates.

And your apps can even talk to Claude: your apps poke Claude and Claude pokes back. Right in the state.

What does "fully" hot-loadable mean?

- Both front-end changes and backend changes are hot-loadable.
- All your state is in the backend and hotloading preserves it.
- You rename a field of a prototype, all its instances' fields get renamed.
- Yeah even structural changes to your data. *That's* what **fully hot-loadable** means.

## Quick Start

In Claude Code, type:

```
/ui show
```

That's it. The **App Console** opens in Playwright. You're in.

## App Console

Your command center. Everything happens here.

```
+---------------------+-----------------------------+
| Frictionless [R][+] | contacts                    |
|---------------------|  A contact manager with...  |
| > contacts 17/21    | [Open] [Test] [Fix Issues]  |
|   tasks    5/5      |-----------------------------|
|   my-app   ████░    | Tests (17/21)               |
|   new-app  --       | [✓] Badge shows count       |
+---------------------+-----------------------------+
| Todos        | [Chat] [Lua]                       |
| 🔄 Reading   | You: add a search box              |
| ⏳ Design    | Agent: Done! Try searching...      |
+---------------------------------------------------+
```

**Left:** Your apps. Badges show test status (`17/21`), build progress (`████░`), or nothing yet (`--`).

**Right:** The selected app. Description, buttons, test results, issues.

**Bottom:** Chat with Claude. Watch the todos. That's where the magic happens.

## Creating Apps

Click **[+]**, name it, describe it, hit Create. Claude fleshes out the requirements. Click **Build**. Done.

Want more control? Write `requirements.md` yourself. Claude respects that.

## Build Modes

Two ways to build. Toggle with the icon in the status bar.

### Fast Mode (Rocket)

Move fast, break nothing:
- Direct code changes
- Auto-checkpoints (rollback anytime)
- Instant hot-reload
- Design docs? We'll catch up later.

This is your playground. Experiment freely.

### Thorough Mode (Diamond)

The full ceremony:
- Updates `requirements.md`
- Updates `design.md` (data model, methods, events)
- Generates code and viewdefs
- Audits and simplifies

When you need it right, not just right now.

## Background Mode

The hourglass/arrows toggle:

- **Hourglass:** Claude works, you wait. Event loop pauses.
- **Arrows:** Claude works in background. Keep clicking around.

Long build? Background it.

## Action Buttons

| Button               | Shows when       | Does what                              |
|----------------------|------------------|----------------------------------------|
| **Build**            | No code yet      | Generates everything from requirements |
| **Open**             | Has viewdefs     | Shows the app right there in the panel |
| **Test**             | Built            | Runs the test checklist                |
| **Fix Issues**       | Has known issues | Claude hunts bugs                      |
| **Make it thorough** | Has checkpoints  | Syncs design with your fast-mode chaos |
| **Review Gaps**      | Has gaps         | Cleans up design/code mismatches       |

### Make it Thorough

Fast mode is fun, but your `design.md` is now lying. This button fixes that.

It reviews all your checkpointed changes, updates the docs to match reality, and clears the checkpoint history. Your design is now telling the truth again.

### Review Gaps

Testing sometimes reveals that the code and design disagree. Those get logged as "Gaps" in `TESTING.md`. This button has Claude sort them out.

## App Files

Each app lives in `.ui/apps/<app-name>/`:

| File                | What it is              | Who writes it                      |
|---------------------|-------------------------|------------------------------------|
| `requirements.md`   | What you want           | You (or Claude expands your notes) |
| `design.md`         | How it works            | Claude (thorough mode)             |
| `app.lua`           | The brains              | Claude                             |
| `viewdefs/*.html`   | The face                | Claude                             |
| `TESTING.md`        | Test checklist + issues | Claude + you                       |
| `checkpoint.fossil` | Fast-mode time machine  | Auto-managed                       |

### requirements.md

Plain language. Be as vague or specific as you want:

```markdown
# Contact Manager

A simple contact list with search.

## Features
- Add, edit, delete contacts
- Search by name or email
- Show contact count
```

Claude will run with it.

### design.md

The blueprint. Claude generates this with:
- **Data model:** Fields, types, what they mean
- **Methods:** Every function, explained
- **Events:** What triggers Claude responses
- **ViewDefs:** Which templates exist

You don't write this. But you should read it when things get weird.

## Chatting

The chat panel is where you drive:

- "Add a search box"
- "Make the delete button red"
- "Show a confirmation before deleting"
- "The save button doesn't work"

Claude reads the app's design, makes the changes, hot-reloads. You see it instantly.

## Checkpoints

Fast mode auto-saves before every change:

```
.ui/mcp checkpoint list contacts
# a2b9eee add search box
# c07d7ee fix button styling
# eb21856 initial state
```

Don't like it? "Undo" or "rollback to 2" in chat.

Done experimenting? "Clear checkpoints" or click **Make it thorough**.

## Lua Console

Click the **Lua** tab. Run code directly:

```lua
contacts:add()
contacts.current.name = "Test User"
contacts:save()
```

Poke at your app's internals. See what breaks. Fix it.

## Tips

1. **Start small.** Core feature first. Iterate.
2. **Fast mode for exploration.** Thorough mode for keepers.
3. **Check `.ui/log/lua.log`** when things go sideways.
4. **Click Open** to actually use your app.
5. **Hit Refresh [R]** if the list looks stale.

## Themes

Frictionless ships with multiple visual themes. Switch themes in the **Prefs** app.

**Available themes:**
- **LCARS** (default) — Dark theme with orange accents
- **Clarity** — Light editorial theme with slate blue
- **Midnight** — Modern dark theme with teal
- **Ninja** — Playful teal theme

Your selection persists across sessions. See [Themes](themes.md) for customization.

## Troubleshooting

**App ignores your clicks?**
Event loop stopped. Type `/ui events` to wake it up.

**Changes not showing?**
Hot-reload is automatic. If it's not working, check `.ui/log/lua.log` for syntax errors.

**Build stuck?**
Check todos for progress. Trash icon clears everything—try again.

**Still stuck?**
Ask Claude. That's literally what it's here for.
