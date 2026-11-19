# Carve: coherency — one structure for Frictionless and ui-engine

> **DRAFT (Daneel, 2026-11-18) — for Bill to edit before its first part is worked.**

Frictionless and ui-engine grew organically, and their structure shows it. Per-session
state is spread across at least ten ID-keyed maps owned by four objects, under two kinds
of session ID. Lua apps reach host features by spawning the host binary (`.ui/mcp`, a bash
script), which ark, the newer Frictionless, has already moved away from: there the host
runs the Lua environment, and host features reach Lua as in-process bindings. Long host
work (fossil, git, curl) runs synchronously on the session's executor and stalls the
whole session while it runs.

This carve moves both projects toward one coherent shape, a piece at a time, with the
system working after every step (the Ship of Theseus pattern). The target is blurry and
will come into focus as pieces move; parts below the top few are expected to change.

**The target shape, as far as it is in focus:**

- **Central Station for sessions.** A `*Session` root that owns its parts, reached
  through a `SessionRef` whose rigs (Lua, Browser, Exec) carry the behaviour, and a
  `Station` any rig can produce. IDs stay at the edges.
  **DECIDED (Bill, 2026-11-19): these are rigs, not views.** "View" read as restricted
  access and collided with ui-engine's on-screen views; a rig is the gear for working one
  part of the session. Renamed in ui-engine's code and docs and in the Central Station pattern.
- **Sewing for host work.** Host operations post continuations instead of blocking the
  executor, through a per-operation wrapper, Tailor.
- **Lua never spawns its host.** Host features are `ui.*` bindings. Lua learns of
  results through ui-engine variables, and change detection pushes them to the browser.

**Two repositories.** Parts marked *(ui-engine)* land in `~/work/ui-engine`; the rest land
here. Queue IDs (`#N`) are this repository's.

The design brainstorm behind this carve is in `.scratch/ark-and-frictionless.md`
(private, gitignored).

## Status

- [x] ~~**Item 1 — session rigs over today's maps (ui-engine).**~~ **LANDED (2026-11-18 — ui-engine, not queued: "Session references and dangerous execution".)**
- [x] ~~**Item 13 — a working closure actor (ui-engine).**~~ **LANDED (2026-11-18 — ui-engine, not queued: "Working closure actor and Tailor".)**
- [x] ~~**Item 2 — Tailor (ui-engine).**~~ **LANDED (2026-11-18 — ui-engine, not queued: "Working closure actor and Tailor".)**
- [x] ~~**Item 3 — the Lua shape of a host operation.**~~ **LANDED (2026-11-18 — ui-engine, not queued: "Host operations: asynchronous `ui.*` bindings".)**
- [x] ~~**Item 4 — a shared checkpoint package.**~~ **LANDED (2026-11-18 — not queued: "checkpoint package: one hardened checkpoint for Frictionless and ark".)**
- [x] ~~**Item 5 — app-console's checkpoint refresh on Tailor.**~~ **LANDED (2026-11-19 — not queued: "Async checkpoint status and install for app-console".)**
- [x] ~~**Item 6 — tell ark.**~~ **LANDED (2026-11-19 — not queued: request `ark-coherency-checkpoint` sent to ark.)**
- [x] ~~**Item 7 — the executor moves into the session (ui-engine).**~~ **LANDED (2026-11-19 — ui-engine, not queued: "Session executor owned by the session; stop never blocks".)**
- [x] ~~**Item 8 — the Lua session moves into the session (ui-engine).**~~ **LANDED (2026-11-19 — ui-engine, not queued: "The Lua session moves into the Session".)**
- [x] ~~**Item 9 — ID maps shrink to the edges (ui-engine).**~~ **LANDED (2026-11-19 — ui-engine, not queued: 9a, 9b and 9c.)**
- [x] ~~**Item 10 — app-console stops spawning the host.**~~ **LANDED (2026-11-19 — not queued: 10a, 10b and 10c.)**
- [ ] **Item 11 — retire synchronous `ui.exec`.** **IN PROGRESS (not queued): 11a and 11b landed; 11c open.**
- [x] ~~**Item 14 — one executor per session (ui-engine).**~~ **LANDED (2026-11-19 — ui-engine, not queued: "One executor per session: Lua runs only on the session's ChanSvc (R254-R263)"; ark told in `ark-one-executor`.)**
- [x] ~~**Item 12 — retire dangerous synchronous execution (ui-engine).**~~ **LANDED (2026-11-19 — not queued: 12a, 12b and 12c; ark told in `ark-dangerous-default`.)**

## Decisions

- **DECIDED (Bill, 2026-11-17): a Lua app never spawns its host.** Host functionality
  reaches Lua as in-process bindings. The executable's subcommands are for external
  callers: the CLI and skills.
- **DECIDED (Bill, 2026-11-17): the closure declares the last stitch** (option A). Only
  the last stitch of a sewing runs afterBatch. Progress pushes are the explicit
  exception.
- **DECIDED (Bill, 2026-11-18): migrate a piece at a time.** Each part must leave the
  system working and be revertible alone. Each piece has exactly one home: when a piece
  moves, its old access path becomes a wrapper over the new home and never keeps a copy.
- **DECIDED (Bill, 2026-11-18): the first plank is Tailor plus checkpoint** (Items 1–5).
  Converging Frictionless on ark's skills-plus-binary model is a later plank, not part of
  this one.
- **DECIDED (Bill, 2026-11-18): do not change too many things at once.** Synchronous
  `ui.exec` stays until nothing uses it (Item 11).
- **Out of scope here:** the robustness items found during the brainstorm
  (`LuaSession.execute` should select on `r.done`; afterBatch's check-then-push under a
  dropped connection). They belong to a robustness pass. The `cleanupSessionSvc` leak is
  the exception, because Item 7 fixes it by construction.

## Item 1

Add `SessionRef`, its rigs and `Station` in ui-engine's `internal/server`, re-exported
through `cli`. The first implementation sits entirely on today's maps (`SessionManager`'s
ID maps, `WebSocketEndpoint.sessionSvc`, `Server.luaSessions`). That is Central Station's
"registry plus ID" route for recovering the root, so it is correct from the start. Nothing
existing changes; this is the seam every later part moves behind.

```go
type SessionRef struct{ /* registry plus vended ID today; weak.Pointer[Session] later */ }

func (r SessionRef) Valid() bool
func (r SessionRef) Lua() LuaRig         // run code, read state
func (r SessionRef) Browser() BrowserRig // afterBatch, connected?
func (r SessionRef) Exec() ExecRig       // post passive, post active, Tailor
func (r SessionRef) Station() Station     // every rig; any rig can produce it
```

`Valid()` is true while the Lua session's `done` channel is open. `LuaSession` needs an
exported accessor for it, since `done` is unexported.

### Item 1's decisions

- **DECIDED (Bill, 2026-11-18): ui-engine has no "current session".** It hands out a
  `SessionRef` for any vended ID, and the host supplies the current one. "Current" already
  lives in Frictionless's MCP server (`currentVendedID`, set at start, cleared at stop),
  where 15 call sites in `internal/mcp` and `flib` read it today.
- **DECIDED (Bill, 2026-11-18): synchronous execution is marked dangerous.**
  `Server.ExecuteInSession` delegates to `PrimitiveExecuteInSession(..., dangerous)` with
  `dangerous` true. A dangerous call panics, or, when the session config property
  `AllowDangerousExecution` is set, logs a warning once per calling site. The property
  defaults to set while callers move (Item 12). Specified in ui-engine's
  `specs/session-ref.md`, R168 to R172.
- `@undecided:` whether a call made from inside an executor, which is a guaranteed
  deadlock, panics regardless of the config. Detecting it needs the goroutine ID, read
  from `runtime.Stack`, which Go does not support directly. (resolution: Bill's call.)
- `@undecided:` whether `LuaRig.Session()` (R164) is narrowed. It hands out the
  `LuaSession`, whose own `ExecuteInSession` goes around the dangerous gate. (resolution:
  settle in Item 1's design phase.)

## Item 2

Tailor, a monadic wrapper scoped to one sewing operation. It exists for that operation
and then disappears.

- **Construction:** `ref.Exec().Tailor(name, deadline)`.
- **Stitches:** `Sew(fn, updateOnErr)` posts a passive stitch; `SewUpdate(fn)` posts an
  active one, which runs afterBatch. `End(fn)` and `EndUpdate(fn)` are the terminal
  stitches; after them the Tailor is closed, a later `Sew` fails loudly, and it leaves the
  pending list. All four post and never wait.
- **Every stitch** recovers panics (which count as errors), logs errors, and checks
  `IsValid()` before posting. An atomic in-stitch flag gates Tailor's access to Lua state;
  the accessor never hands out the raw `*LuaSession`.
- **Passive posting** needs one method on `WebSocketEndpoint`: the same post as
  `ExecuteInSessionAsync` without the afterBatch call. `updateOnErr` decides inside the
  posted closure whether to call afterBatch.
- **The pending list lives on `*Session`** (on the session's registry entry until Item 7),
  registered when the Tailor is created. Each entry records the name, the creation time,
  the last stitch's time and caller site (`runtime.Caller`), and a state: waiting, failed,
  or long-lived. Long-lived chains are reported apart from waiting ones (Cry Wolf).
  Destroying a session logs whatever is still pending.
- **The deadline** is declared by the operation. When it passes, the Tailor closes as
  failed. The Tailor carries a `context.Context`, cancelled on `End`, on the deadline, and
  on session shutdown; workers run child processes with `exec.CommandContext`, so a
  timeout also kills the child. Long-lived Tailors have no deadline.
- **A sewing that never completes is a leak, not a hang.** Nothing waits on it. The
  pending list and the deadline exist to manage the leak.

### Item 2's decisions

- **DECIDED (Bill, 2026-11-18): no debug page in Item 2.** `ExecRig.PendingTailors()` is
  the data source; a page can be added later.
- **DECIDED (Bill, 2026-11-18): an error ends the sewing.** A stitch that errors or panics
  closes the Tailor as failed.
- **DECIDED (Bill, 2026-11-18): misuse is loud, lateness is not.** A stitch posted after
  `End` panics; one posted after a failure, deadline or session end is logged once and
  dropped.
- **DECIDED (Bill, 2026-11-18): stitches reach Lua through `Tailor.State()`,** an
  `*lua.LState` valid only inside the Tailor's running stitch, never the `LuaSession`.
  Specified in ui-engine's `specs/tailor.md`.

## Item 3

How a host operation looks from Lua, under the `ui.*` namespace.

- **Results reach Lua through variables.** The terminal stitch writes the result into
  Lua state (an app field, for example), and afterBatch pushes it to the browser through
  ordinary change detection. Lua code reacts to state, not to a callback.
- **A callback is allowed where signalling through variables does not make sense.**
- The binding starts the operation and returns at once; the Go side owns the Tailor and
  decides which stitch is terminal, so Lua never spells "last stitch".

### Item 3's decisions

- **DECIDED (Bill, 2026-11-18):** namespace `ui.*`, because this is part of Frictionless
  UI development.
- **DECIDED (Bill, 2026-11-18):** variables first; a callback only where variables do not
  fit.
- **DECIDED (Bill, 2026-11-18): the caller names the target.** Every host operation is
  called as `ui.op(target, field, args..., [cb])`. The terminal stitch writes
  `target[field]` (the result) and `target[field .. "Error"]` (the message, or nil) in one
  `EndUpdate`. Settled against Item 5: app-console's results are per app and its field
  names are its own, so a fixed variable per operation does not fit.
- **DECIDED (Bill, 2026-11-18): the callback is optional, trailing, `cb(result, err)`,**
  called after the fields are written.
- **DECIDED (Bill, 2026-11-18): no pending marker for now.** Add one when a caller wants
  it.
- **DECIDED (Bill, 2026-11-18): ui-engine provides the helper,** `LuaRig.DefineHostOp`;
  a host supplies only the operation's argument reading and its `work` function.
- **LANDED (2026-11-18, ui-engine 994a0ce):** `specs/host-ops.md`, R205-R216,
  `internal/server/hostop.go`. The deadline is applied to `work`'s context; the Tailor
  closes as failed at twice the deadline as a backstop, and a `work` that ignores its
  context delivers nothing.

## Item 4

Lift checkpointing into a Frictionless package that both binaries import. ark's
`cmd/ark/checkpoint.go` is the second real instance beside the bash `.ui/mcp`, so the
generalization is earned. The frictionless binary uses the package for its own Lua
bindings; ark gets the bindings through flib and its subcommands from the same package.

- **Fossil discovery moves inside the package**, so apps never see a fossil path. It
  should follow whatever ark settles on for its own fossil fix.
- **"Has this downloaded app changed since install"** (app-console's
  `fossil diff --from original`) becomes an operation of the package too.

### Item 4's decisions

- **DECIDED (Bill, 2026-11-18):** a public package, `github.com/zot/frictionless/checkpoint`,
  lifted from ark's `cmd/ark/checkpoint.go`. Its methods return values; one shared
  `Command` does the printing for both binaries, and each host passes its own `onChange`
  for the app-console refresh. ark deletes its copy later, through Item 6.
- **DECIDED (Bill, 2026-11-18):** fossil discovery is `~/.claude/bin/fossil`, then PATH
  (`fossil.exe` on Windows): the lookup ark said it is moving to.
- **DECIDED (Bill, 2026-11-18):** the frictionless binary gains `checkpoint`; the bash
  `mcp checkpoint` keeps its fossil download and hands everything else off. Skills and
  agents calling `.ui/mcp checkpoint` keep working.
- **DECIDED (Bill, 2026-11-18):** no Lua bindings here; Item 5 brings the first.
- **Found 2026-11-18: each version carried half the hardening.** ark's has the walk-up
  and working-tree fixes (verified checkout, `--nested --keep`, no `update trunk`,
  `mtime-changes` off, checked steps); Frictionless's bash still runs `fossil update trunk`
  in four places. Frictionless's bash has the baseline set-aside and exact branch names
  (R192-R194, from ark's own October request); ark's deletes before verifying and matches
  branches by substring. The package carries both.
- **Found 2026-11-18 (measured): the changed-since-install check never worked.**
  `fossil diff --from original` fails with "cannot resolve name: original" and app-console
  discards stderr, so every downloaded app reads as unchanged. `Changed` is specified anew:
  the working files against the tip of trunk in `original.fossil` (R215-R217).
- **LANDED (2026-11-18, fb6126c):** `specs/checkpoint.md`, R201-R217, package
  `checkpoint/`, `frictionless checkpoint`, and the `install/mcp` hand-off. O9 resolved;
  O10 (ignore-glob omits `original.fossil`) recorded. ark keeps its own copy until Item 6.

## Item 5

The first real consumer. App-console's checkpoint refresh today loops over every app and
runs `checkpoint count` plus a `fossil diff` for each, two child processes per app,
serially, on the executor. It becomes one host operation: a single stitch out, the fossil
runs on the Go side (in parallel if worthwhile), and a single terminal stitch back that
writes every app's count and changed flag into app-console's state.

This part removes app-console's `.ui/mcp checkpoint` calls and its direct `fossil` calls.

### Item 5's decisions

- **DECIDED (Bill, 2026-11-19):** one host operation, `ui.checkpointStatus(target, field,
  names)`, defined in `setupMCPGlobal` so ark gets it through flib. It returns every
  named app's count, downloaded flag and changed flag in one table; app-console's
  accessors read from it. Sequential first; parallel only if a measurement asks for it.
  The 10-second cache and the CLI's refresh nudge stay.
- **DECIDED (Bill, 2026-11-19):** a read-only `App.Status()` in the package. A status
  poll verifies an open checkout but never writes settings: `Count` reapplies them on
  every call, a database write per app every ten seconds, contending with an agent's
  save.
- **DECIDED (Bill, 2026-11-19):** install moves into the package as `App.Install()`
  (baseline, then copy to `original.fossil`), exposed as `ui.checkpointInstall`; the
  install flow continues in its callback. O10 closes here: the ignore-glob gains
  `original.fossil`.
- **DECIDED (Bill, 2026-11-19):** app-console loses every `.ui/mcp checkpoint` and direct
  `fossil` call; its own requirements gain the asynchronous refresh.
- **LANDED (2026-11-19, 7d608fd):** R218-R225; checked live on a scratch server built from
  the tree: the rocket arrived asynchronously and the changed-since-install pencil showed for
  the first time. O10 resolved.
- **Found 2026-11-19 (live check):** `install/apps/prefs/.fslckout` was tracked and bundled,
  binding every fresh install's prefs app to a developer repository. Removed in 7d608fd;
  the bundler does not yet refuse fossil files.

## Item 6

One message to ark once Items 1–5 have settled the shape, not before. It covers: the
checkpoint package lifted out of `package main` (Item 4), ark's app-console needing the
same in-process binding, and ark's own app (`apps/ark/app.lua`, the `arkCmd` helper),
which spawns the ark binary through synchronous `ui.exec` at about two dozen call sites.

- **LANDED (2026-11-19):** `requests/ark-coherency-checkpoint.md`, sent. It covers the
  checkpoint package and the four defects ark's copy still carries, the in-process
  checkpoint operations ark gets through flib (answering part of `windows-defects-frictionless`),
  ark's 23 synchronous `ui.exec` spawns and `DefineHostOp`, and the bundler fix. It states
  that nothing is released yet.

## Item 7

Move the executor out of `WebSocketEndpoint.sessionSvc` into `*Session`. The rig
implementation changes; callers do not. The payoff is concrete:
`cleanupSessionSvc` has no callers today, so every destroyed session leaks its executor
goroutine. Owned by the session, the executor ends with it. Test: the goroutine count
after a session is destroyed.

- **LANDED (2026-11-19):** ui-engine 9923979. The executor is born in `NewSession` and
  stopped by `DestroySession` after backend teardown, and by a failed `CreateSession`, a
  second leak the item did not name. Stopping exposed a latent hang: after `ChanSvc.Stop`
  every late post blocked forever on the one-slot channel, so the actor gained a done
  channel, `ErrStopped`, and a per-job `launched` flag that lets a synchronous caller
  tell an abandoned closure from one still in flight (R218-R228; R203 and R204 retired).

## Item 8

Move the Lua session out of `Server.luaSessions` and the store adapter's copy into
`*Session`. `Server.ExecuteInSession` becomes a thin wrapper over the rigs.

- **LANDED (2026-11-19):** ui-engine 6993984. The store adapter's copy turned out to be
  write-only, so it was deleted rather than moved. `Server.luaSessionFor` is now the one
  lookup by vended ID. `ExecuteInSessionAsync` became `ExecRig.PostUpdate`; the
  synchronous `ExecuteInSession` finds the session the same way but stays off the rigs,
  since no rig may wait (R237). Found in passing: a `main.lua` that fails during creation
  leaks its Lua session, recorded as ui-engine gap O8.

## Item 9

Once nothing inside needs them, the ID maps shrink to edge lookups: an ID arriving on the
wire (backend socket, MCP, URL) becomes a `SessionRef` once, and pointers take over.

Split 2026-11-19 (Bill), one piece at a time:

- **9a — the store adapter. LANDED (2026-11-19):** ui-engine dddba18. Its `backends` copy is
  gone and `nextServerVarId` is on each `LuaBackend`. `varToSession` turned out to be always
  empty (its only writer, `CreatePathVariable`, had no callers) and could not have worked
  anyway, since variable IDs repeat across sessions; every variable operation now names its
  session, so `session:destroyVariable` destroys for the first time. Found: creating a
  server child variable panics (ui-engine gap O9), and `lib/lua/session.lua` never loads,
  leaving `Variable:getValue` and `getProperty` unreachable (gap O10).
- **9b — the endpoint. LANDED (2026-11-19):** ui-engine 3246634. Each connection holds its
  `*Session`, set once at connect, in place of `sessionBindings`; `HasConnectionsForSession`
  and `Broadcast` read the session's own connection set instead of scanning every
  connection. `reconnectTokens` was not moved but deleted: nothing in ui-engine,
  Frictionless or ark issued or checked a token, and the spec never asked for one.
- **9c — URL paths. LANDED (2026-11-19):** ui-engine 12b310b. Not moved but deleted: only
  tests called `RegisterURLPath` and `ResolveURLPath`, and nothing lets a backend register
  a path. The feature (`interfaces.md`, presenters at URL paths) is specified but unbuilt,
  with two other unconnected pieces (`internal/router`, `web/src/router.ts`); ui-engine
  gap O11. Building it decides where routes live.
- **Edges, on purpose:** the `SessionManager` registry and its vended-to-internal ID maps,
  the connection registry, and the viewdef manager's `sentViewdefs` (in a package that
  cannot name a `Session`) stay keyed by ID. They are not unfinished work.

## Item 10

The rest of app-console's host spawns: `.ui/mcp linkapp` and the remaining `.ui/mcp`
calls become `ui.*` bindings, with linkapp in the Item 4 package or beside it. This is
also where the bash `.ui/mcp` stops mattering on Windows.

### Item 10's decisions

- **DECIDED (Bill, 2026-11-19): a junction on Windows, as ark does.** ark's `MakeDirLink`
  writes a directory junction (no privilege, no Developer Mode, no `mklink`) and repairs
  stale ones at start, since a junction's target is absolute. Unix keeps its relative
  symlink, so R85 now holds on Unix only.
- **Split 2026-11-19 (Bill).** The carve counted only `linkapp`; app-console had twelve
  other bash spawns. One piece at a time:
  - **10a — activation. LANDED (2026-11-19):** Frictionless 4f380ca. The `applink`
    package (lifted from ark's dirlink, which ark can later import back),
    `frictionless linkapp`, and synchronous `ui.linkApp` / `ui.unlinkApp` /
    `ui.deleteApp`; the bash `install/linkapp` is gone, closing gap I1.
  - **10b — the GitHub install. LANDED (2026-11-19):** Frictionless bf96ee9. The
    `githubinstall` package and the asynchronous `ui.githubInvestigate` /
    `ui.githubInstall` host operations replace the `mkdir`/`curl`/`unzip`/`mv`/`rm`
    pipeline and both `curl` fetches; app-console investigates and installs through
    callbacks, so the session no longer stalls on the network. The live install found
    `install/apps/prefs/.fslckout` still on GitHub (removed locally in 7d608fd,
    unpushed); a push clears it.
  - **10c — local file checks and desktop open. LANDED (2026-11-19):** Frictionless
    17dbe6e. No new Go: ui-engine's portable `ui.isDir`, `ui.listDir` and `ui.mkdir`
    (its R146-R148) replace `test -d`, both `ls` and `mkdir -p`, with dot-entries
    skipped as `ls` did. The conflict check lost its one-second throttle, since a stat
    is cheap. Both `xdg-open` methods had no callers (the viewdef's links already open
    source and readme in a new tab) and were deleted. app-console now runs no shell.

## Item 11

Retire synchronous `ui.exec` (ui-engine `internal/lua/portable.go`, `#141`, R150). It
blocks the executor for the child's whole run. It stays until every caller has moved to
an async form or to a dedicated binding, retrofitted as each caller is touched for other
reasons. Then it is removed.

- **Known callers (2026-11-18):** ark's `apps/ark/app.lua`, through `arkCmd`. The bundled
  Frictionless apps do not use it; they still shell out through `io.popen` and
  `os.execute`, and they should move to async forms rather than to `ui.exec`.
- **Done when** a search of ui-engine, Frictionless and ark finds no caller, and the
  async form covers what the callers needed.
- **Split 2026-11-19 (Bill).** Measured that day: ark's app was the only caller (23
  `arkCmd` sites plus `checkServer`), and no async form existed for it to move to.
  - **11a — the async form. LANDED (2026-11-19):** ui-engine "ui.execAsync: run a program
    off the session; ui.exec deprecated (R250-R253)". `ui.execAsync(target, field, argv
    [, cb])` is a host operation built into every session before `main.lua`, with a
    per-call limit (60 seconds, `argv.timeout`, 0 for none); `ui.exec` is deprecated.
  - **11b — the bundled apps. LANDED (2026-11-19):** Frictionless "Bundled apps run no
    shell: portable primitives and ui.execAsync (Item 11b)". The last 12 shell calls in
    mcp, claude-panel and prefs moved: `ls`, `mkdir -p`, `base64 -d` and `uname` to the
    portable primitives, `git` and the `curl` release check to `ui.execAsync`. The update
    check takes a callback so prefs' Check now reports the new answer, and the update
    event names its asset by Go's `os-arch`, which the release files use. No bundled app
    shells out any more.
  - **11c — ark, then removal.** Tell ark `ui.execAsync` exists and ask whether `arkCmd`
    becomes in-process bindings (ark spawns its own binary, which Lua should not) or async
    calls. `ui.exec` is removed from ui-engine once ark has moved.

## Item 12

Move every caller off dangerous synchronous execution, then flip
`AllowDangerousExecution`'s default so a dangerous call panics. The once-per-site warnings
from Item 1 are the list of remaining callers.

- **Free first retrofits:** ui-engine's `PushViewdefs` and `triggerSessionRefresh` run a
  no-op only to trigger change detection, and nothing waits for their result. They become
  `ExecRig.PostUpdate`.
- **Callers known on 2026-11-18,** all running off the executor and therefore safe:
  Frictionless's `SafeExecuteInSession` (about a dozen MCP tool and `Stop` sites), and
  flib's `WithLua` and `WithLuaUpdate`. Request-and-response callers that need a result
  may stay synchronous by calling `PrimitiveExecuteInSession` with `dangerous` false once
  each is reviewed as never running on an executor.
- **A wrapper hides its callers (measured 2026-11-18).** Frictionless's test run warned
  from one site, `internal/mcp/server.go:95`, which is `SafeExecuteInSession` itself. Every
  MCP tool call goes through it, so the per-site list reports one entry for about a dozen
  callers. Retrofitting the wrapper (to the primitive with `dangerous` false, once its
  callers are reviewed) clears them all at once; the same holds for flib's `WithLua` and
  `WithLuaUpdate`.
- **Done when** no site warns in normal use, and the default is flipped.
- **Split 2026-11-19 (Bill).** Recounted that day: the same two ui-engine sites, and
  Frictionless's `SafeExecuteInSession` (11 callers) and flib's `WithLuaUpdate`. Item 14's
  single executor means each wait now has one queue to wait on, not two.
  - **12a — ui-engine's own sites. LANDED (2026-11-19):** ui-engine "ui-engine makes no
    dangerous call: PushViewdefs and refresh post (R264)". `PushViewdefs` and
    `triggerSessionRefresh` post with `ExecRig.PostUpdate` and do not wait.
  - **12b — Frictionless. LANDED (2026-11-19):** Frictionless "Reviewed synchronous
    execution: SafeExecuteInSession and WithLuaUpdate use the primitive (Item 12b)". All 11
    `SafeExecuteInSession` callers run off the executor (HTTP and MCP tool handlers,
    startup, the subscription poll goroutine), recorded in `crc-MCPServer.md`; so does
    ark's one `WithLuaUpdate` caller. Both now use the primitive with `dangerous` false.
    flib is a library, so its rule for future embedders lives in `specs/mcp.md`. The stale
    "Lua executor goroutine" wording for `WithLua` was corrected on the way.
  - **12c — the flip (ui-engine). LANDED (2026-11-19):** ui-engine "Dangerous execution
    panics by default (R265)". Measured first with a trial flip across ui-engine,
    Frictionless and ark: only two ui-engine test call sites tripped it (the warn test,
    which now sets the property itself, and Item 14's test, which now uses the primitive),
    and ark passed in full. The trial also found gap O12 in ui-engine: a dangerous call made
    straight from a `go` statement is reported at `runtime/asm_amd64.s`. ark gets the new
    default through flib and was told in `ark-dangerous-default`.

## Item 13

Port ark's `ChanSvc` (`~/work/ark/svc.go`) into ui-engine's `internal/server/svc.go`.
ui-engine's is broken: a closure actor must preserve order, and this one does not.
Ranked ahead of Item 2, because Tailor depends on it.

ui-engine's `Svc` sends each closure from a freshly spawned goroutine so the poster never
blocks. Two such goroutines race, so closures posted by one goroutine can run out of
order. Found 2026-11-18 when Item 2's tests ran a Tailor's stitches out of order. The same
path carries incoming browser messages (`websocket.go`, one `Svc` per WebSocket message)
and session timers, so R99 ("multiple `setImmediate` calls queue in order") has not held.
ark found and fixed the same defect on 2026-07-31.

ark's version: a post is a direct send on the caller's goroutine; the actor's goroutine
never runs a closure, only a queue and an in-flight flag; each closure runs on its own
goroutine and posts `next` back when done; the channel is one slot deep; panics are not
recovered. The contract is ark's `specs/db-concurrency.md` §Delivery, R3344 to R3350 and
R3607 there.

### Item 13's decisions

- **DECIDED (Bill, 2026-11-18): port ark's ChanSvc, not a new lock-free queue.** The
  atomic two-list design discussed first (queue plus free list, compare-and-swap heads)
  was set aside: ark's approach is already proven and needs no free list, whose
  concurrent pops would have had an ABA problem.

## Item 14

One executor per session (ui-engine). Ranked ahead of Item 12, because it is a crash and
Item 12 is a cleanup; with one executor, Item 12 also has one place to wait instead of two.

Each session runs two executor goroutines: the session's own (Item 7) and the
`LuaSession`'s, fed by `executorChan` (`internal/lua/runtime.go`, `startExecutor`). Rig
work, host-op delivery and `Server.ExecuteInSession` nest the second inside the first, so
they take turns. Two paths do not:

- **Change detection** runs on the session executor: `LuaSession.AfterBatch` calls
  `DetectChanges` directly, not through `execute`, and evaluating a method path runs Lua.
  It runs only while a browser is connected, which is the normal case.
- **flib's `WithLua`** posts straight to the `LuaSession` executor, bypassing the session
  executor. The dangerous-execution warning cannot see it.

**Measured 2026-11-19** with a throwaway probe in `internal/server` (no race detector on
this machine: no cgo): a connection attached, `PostUpdate` no-ops on one goroutine, direct
`LuaSession.ExecuteInSession` writes on another. A method-bound variable's `App:probe()`
ran on the session executor while the Lua executor was inside a closure 46 and 51 times
out of 101, and two of seven runs ended in gopher-lua's `fatal error: concurrent map read
and map write` (`LTable.RawGetString`), which `recover` cannot catch: the whole process
dies. ark calls `WithLua` at three sites (`server.go` 2743, 2779 and 6048), so ark is
exposed now. Frictionless does not call `WithLua`.

- **Not yet classified:** the 17 `r.execute` call sites inside `internal/lua`. Some are
  only reached nested; the hot-loader's reload and timer firings are the likely direct
  entries.
- **Likely shape (not decided):** `LuaSession.execute` runs on the session executor
  instead of a goroutine of its own, so every path into the Lua state is one queue.
- **Done when** every entry into a session's Lua state goes through one executor, a
  regression test that counts overlaps (the probe's method, not a crash, which would kill
  the test binary) stays at zero, and ark has been told.
- **Fix LANDED (2026-11-19):** ui-engine "One executor per session: Lua runs only on the
  session's ChanSvc (R254-R263)". Fix A, decided by Bill: `LuaSession` lost its own
  goroutine and runs on the session's ChanSvc; work already on it calls `RunInSession`,
  work from outside posts and waits. ChanSvc gained `Drained()`, because `done` closes at
  the stop sentinel while a closure may still be running, so shutdown closes the VM only
  after the drain and never waits itself. Sewing, per Bill: hot reload and `main.lua`
  compile off the executor, so a file that does not compile touches no session. Eleven
  unused `LuaSession` methods were deleted. All 17 entry points were classified first:
  `WithLua` and the hot-loader were the everyday direct entries. Checked live in
  Frictionless with a browser connected, including a hot reload of a broken file.
- **Ark told (2026-11-19):** request `ark-one-executor`. Its four callers (three `WithLua`, one `WithLuaUpdate`) run
  off the executor (HTTP handlers, startup, its listen loop), so they keep working; the message is that the
  crash is fixed once ark takes this ui-engine, and that a `WithLua` closure now runs on
  the session executor, so it should do its I/O before the call.
