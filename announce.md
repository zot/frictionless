# Announcement Draft

## For: Hacker News, Dev.to, Twitter/X, etc.

---

## Hacker News (Show HN)

**Title:** Show HN: Frictionless – Hot-loadable apps where Claude Code modifies state in real-time

**Text:**
I built an app platform for Claude Code where the AI can actually collaborate with you inside running apps—not just generate code, but read and modify app state as you interact.

The pain point: when you ask Claude to build a UI, you watch it burn tokens on boilerplate—API endpoints, state management, frontend wiring. Then something breaks and Claude debugs across layers. Then you restart and lose your test data.

Frictionless skips all that:

- Claude writes app logic, not plumbing—way fewer tokens per feature
- Changes hotload instantly, even backend changes—no restarts, no lost state
- Rename a data field and all existing instances update automatically

Example: `/ui-thorough make a contacts app with search` → working app with persistence.

The bidirectional integration is the interesting part: your app can surface info to Claude, and Claude can push changes back. Building dashboards, forms, or prototypes becomes a conversation.

MIT licensed, written in Go. Demo: https://youtu.be/Wd5n5fXoCuU

GitHub: https://github.com/zot/frictionless

---

## Twitter/X Thread

**Tweet 1:**
Built something for Claude Code users: Frictionless

Apps where Claude doesn't just write code—it collaborates with you inside the running app. Change backend data, UI updates. Claude modifies state, you see it instantly.

No API layer. No frontend JS. No restarts.

🔗 https://github.com/zot/frictionless

**Tweet 2:**
The pain: ask Claude to build a UI and watch it burn tokens on API endpoints, state management, frontend wiring. Then debug across layers. Then restart and lose your test data.

Frictionless skips all that. Claude writes logic, not plumbing. Way fewer tokens. No restarts.

**Tweet 3:**
The wild part: rename a field in your data model and all existing instances get their fields renamed too. While the app is running.

That's what "fully hot-loadable" means.

Demo: https://youtu.be/Wd5n5fXoCuU

**Tweet 4:**
MIT licensed. Works as a Claude Code MCP server.

Install: tell Claude "Install using github zot/frictionless readme"

Or grab a binary and wire it up manually.

Happy to answer questions.

---

## Dev.to / Blog Post

**Title:** Building Apps Where Claude Code Is a Real-Time Collaborator

**Tags:** #claude #ai #lua #webdev #opensource

**Content:**

### The Problem

When you ask Claude Code to build a UI, you watch it burn tokens on boilerplate—API endpoints, state management, frontend wiring. Then something breaks and Claude spends even more tokens debugging across layers. Then you restart the server and lose your test data.

You want to iterate fast, not watch Claude write DTOs.

### The Solution

Frictionless skips all that:

1. **Claude writes logic, not plumbing** — Way fewer tokens per feature
2. **Instant hotload** — Even backend changes, no restarts, no lost state
3. **Structural hotload** — Rename a field, all existing instances update

### Claude as Collaborator

The interesting part: apps integrate bidirectionally with Claude Code.

Your app can surface information to Claude (via MCP). Claude can modify your app's state directly. It's not "Claude generates code" → "you run it". It's a live collaboration where both sides can read and write.

### Example

```
/ui-thorough make a contacts app with search and inline editing
```

Claude builds it. You use it. You ask Claude to add a feature. It hot-loads. No restart.

### Try It

MIT licensed. Install by telling Claude:

```
Install using github zot/frictionless readme
```

Or grab a binary: https://github.com/zot/frictionless/releases

Demo: https://youtu.be/Wd5n5fXoCuU

---

## LinkedIn (if needed)

**Post:**
Released an open-source project: Frictionless

It's an app platform for Claude Code where the AI collaborates with you inside running applications—reading and modifying state in real-time, not just generating code.

The result: Claude writes app logic instead of burning tokens on boilerplate. Changes hotload instantly—no restarts, no lost state.

Built with and for Claude Code. MIT licensed.

GitHub: https://github.com/zot/frictionless
Demo: https://youtu.be/Wd5n5fXoCuU
