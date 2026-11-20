# Carve: robustness — a pass over failure paths in Frictionless and ui-engine

The coherency carve (`carves/done/coherency.md`) set aside the robustness items found
while it was being brainstormed, so they would not widen its scope. This carve takes them
up, and is where later ones of the same kind go: paths that work in normal use but have
not been shown to hold when something fails partway, such as a connection drop, a session
shutdown, or a timeout.

**Two repositories.** Parts marked *(ui-engine)* land in `~/work/ui-engine`; the rest land
here.

The items came from the coherency brainstorm, which is private and lives in `.scratch/`.
Each item below states what it contributed.

## Status

- [ ] **Item 1 — afterBatch tolerates a connection dropping mid-push (ui-engine).** **OPEN (not queued.)**
- [x] ~~**Item 2 — `LuaSession.execute` waits forever after shutdown (ui-engine).**~~ **ASSESSED (Daneel, 2026-11-20.)**
- [x] ~~**Item 3 — endpoint executors leak on restart (ui-engine).**~~ **ASSESSED (Daneel, 2026-11-20.)**

## Decisions

- **DECIDED (Bill, 2026-11-20): robustness is its own carve.** The convergence on ark's
  skills-plus-binary model, set aside by coherency at the same time, is a direction rather
  than a robustness fix, and has its own carve (`carves/skills-and-binary.md`).

## Item 1

Confirm that afterBatch tolerates a connection dropping mid-push. Parked by the coherency
brainstorm as "should be confirmed tolerant", which makes it a check, not a known defect.

`Server.AfterBatch` (ui-engine `internal/server/server.go`) runs change detection, then
for each update reads its watchers (`b.GetWatchers`) and either queues the message to the
session's batcher or, without one, sends it to each watcher's connection through
`wsEndpoint.Send`. A connection can close between the watcher read and the send, and
nothing re-checks.

What the brainstorm established: a dropped connection is normal. With no browser the Lua
work still runs, and a later connection resyncs because `onDisconnect` clears
sent-tracking. What it did not establish is what a send or a batcher flush to a connection
that has just closed does: whether it errors quietly, logs, panics, or leaves sent-tracking
claiming the browser has a value it never received.

- **Done when** a test closes a connection between the watcher read and the send, and
  shows the push neither panics nor blocks, and that the next connection to the session
  gets every current value.

## Item 2

The brainstorm found that a late stitch could reach `LuaSession.execute` after the
session's Lua goroutine had returned, and wait forever on a result nobody would send. Its
fix was for `execute` to select on `r.done`.

**Assessed 2026-11-20: no longer reachable.** The working closure actor and one executor
per session (coherency Items 13 and 14) removed the Lua goroutine. `execute` now posts
through `SvcSync` (ui-engine `internal/server/svc.go`), which returns `ErrStopped` when the
post fails because the actor has stopped, and, while waiting, selects on the actor's `done`
too: a closure that never launched returns `ErrStopped`, and one already running is waited
for.

## Item 3

`cleanupSessionSvc` had no callers, so each destroyed session left its endpoint executor's
goroutine behind.

**Assessed 2026-11-20: fixed by construction.** Coherency Item 7 moved the executor into
`*Session`, so it ends with the session, and its stop never blocks.
