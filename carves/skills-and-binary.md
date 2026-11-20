# Carve: skills-and-binary — Frictionless reaches Claude through skills and its binary

ark is a newer Frictionless and carries Bill's current thinking on it. Two conventions
from ark landed in Frictionless through the coherency carve (`carves/done/coherency.md`):
the host runs the Lua environment, so a Lua app never spawns its host, and host features
reach Lua as in-process bindings. A third did not: ark reaches Claude through skills and
its own executable's subcommands, with no MCP server, because an MCP server's tools cost
context in every session whether they are used or not. Frictionless still runs one.

This carve moves Frictionless to that model. The coherency carve named it as a later plank
and set it aside (DECIDED, Bill, 2026-11-18). The target is blurry and is expected to come
into focus as the inventory below is done.

## Status

- [ ] **Item 1 — inventory what the MCP server does that skills and the binary do not.** **OPEN (not queued.)**

## Decisions

- **DECIDED (Bill, 2026-11-20): its own carve.** It is a direction, not a robustness fix,
  so it is not part of `carves/robustness.md`.

## Item 1

Measured 2026-11-20: the MCP server registers nine tools, `ui_configure`,
`ui_open_browser`, `ui_run`, `ui_status`, `ui_install`, `ui_update`, `ui_display`,
`ui_audit` and `ui_theme`. The skills already reach the running server through the `.ui/mcp`
command (`run`, `checkpoint`, `audit`, `theme`, `linkapp`), so part of a CLI exists.

For each tool: is there a subcommand or skill that does the same, what would be missing
without the tool, and what about it depends on the server being an MCP server (the
server's lifetime, notifications back to Claude, the event loop the apps use to reach
Claude). The answer splits the rest of this carve into items.

- **Done when** every tool has a row: its replacement, or what has to be built first.
