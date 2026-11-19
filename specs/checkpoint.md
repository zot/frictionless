# Checkpoints

**Language:** Go
**Environment:** the `checkpoint` package (`github.com/zot/frictionless/checkpoint`),
imported by the frictionless binary and by ark

## Purpose

A checkpoint is a fossil commit of one app's files, so an agent building an app can save,
compare and roll back as it works. Each app keeps its own repository,
`{app_dir}/checkpoint.fossil`.

Checkpointing lives in one Go package that both Frictionless and ark import. Before this,
Frictionless had a bash version in the `mcp` script and ark a Go port, and each had half
of the hardening: ark's verified the checkout and never touched the working tree when
switching branches, Frictionless's made baseline safe. The package has both.

## Commands

The frictionless binary takes `frictionless checkpoint CMD APP [ARG]`, where APP names a
directory under `{base_dir}/apps/`. The `mcp` script's `checkpoint` command hands off to
it (`specs/helper-scripts.md`). ark runs the same commands through its own binary.

| Command | Description |
|---------|-------------|
| `checkpoint save APP [MSG]` | Save current state (no-op if unchanged) |
| `checkpoint list APP` | List checkpoint history |
| `checkpoint rollback APP [N]` | Roll back to the Nth checkpoint; with no N, undo the last |
| `checkpoint diff APP [N]` | Show the diff from checkpoint N (default 1) to the current files |
| `checkpoint clear APP` | Reset to baseline (same as baseline) |
| `checkpoint baseline APP` | Set the current state as the new baseline |
| `checkpoint count APP` | Print the number of checkpoints, not counting the baseline |
| `checkpoint update APP [MSG]` | Save the current files on the "updates" branch |
| `checkpoint local APP [MSG]` | Save the current files on the "local" branch |

- A command that fails prints the reason on stderr and exits non-zero. Success and failure
  are told apart by exit status alone.
- After a command that changes checkpoints (save, rollback, clear, baseline, update,
  local), the host tells app-console to refresh its checkpoint display. This is best
  effort: when no server is running it does nothing, and it never changes the exit status.

## Go API

```go
// Find locates fossil: ~/.claude/bin/fossil first, then fossil on PATH
// (fossil.exe on Windows). It returns ErrNoFossil when there is none.
func Find() (string, error)

// App is one app's checkpoints.
type App struct {
    Dir    string // the app directory
    Fossil string // the fossil binary
}

func (a App) Save(msg string) (saved bool, err error)
func (a App) List() ([]string, error)          // one line per checkpoint, newest first
func (a App) Rollback(n int) error             // n == 0: undo the last checkpoint
func (a App) Diff(n int) (string, error)
func (a App) Baseline() (restored []string, err error) // the preserved branches it restored
func (a App) Count() (int, error)              // 0 when there is no repository
func (a App) Update(msg string) error
func (a App) Local(msg string) error
func (a App) Changed() (bool, error)           // see "Changed since install"
func (a App) Status() (Status, error)          // see "Status"
func (a App) Install() error                   // see "Install"

// Status is one app's checkpoint state.
type Status struct {
    Count      int  // trunk checkpoints, not counting the baseline
    Downloaded bool // the app has original.fossil
    Changed    bool // see "Changed since install"
}

// Command runs one checkpoint command for an app under appsDir, as the binaries do,
// writing what it prints to out. onChange is called after a command that changes
// checkpoints; it may be nil.
func Command(args []string, appsDir string, out io.Writer, onChange func()) error
```

The methods return values and print nothing; `Command` does the printing.

When fossil is missing, the error says how to install it: where to put the binary, and
the download for each platform.

## Safety

Each rule here exists because a checkpoint once damaged an app or lost history.

- **Every fossil step is checked.** A failing step stops the command, and the error names
  the fossil command and carries fossil's own output.
- **The checkout is the app's own.** Before acting, the package asks fossil which
  repository and checkout root it would use from the app directory, and stops unless both
  are the app's. Fossil otherwise walks up into an enclosing checkout, such as the
  project's own repository.
- **Opening never extracts over the working tree.** A checkout is opened with `--nested`
  (bind here even inside an enclosing checkout) and `--keep` (leave the files alone). A
  checkout bound to the wrong repository is reported, never silently replaced.
- **Branch switches never touch the working tree.** Saving to "updates" or "local" and
  returning to trunk changes which branch the next commit goes to, and nothing else.
- **Commits hash content.** Fossil's `mtime-changes` setting is off on every opened
  checkout, so a same-size edit within the same second as the last checkpoint is not
  missed.

### Baseline

`checkpoint baseline` replaces the repository, so it must never lose history. It sets the
old repository aside instead of deleting it, builds the new baseline, and restores the
preserved "updates" and "local" branches into it. Only when every step has succeeded,
and each preserved branch is present in the new repository, is the old one discarded.

If any step fails, the old repository and its checkout are put back exactly as they were,
byte for byte. The command says so on stderr and exits non-zero; it never reports success
for a baseline it did not set.

Preserved branches are recognized by their exact names, so a branch such as
`local-experiment` is not mistaken for `local`.

## Changed since install

When app-console installs a downloaded app, `Install` sets a baseline and copies the
repository to `{app_dir}/original.fossil`. `Changed` reports whether the app's files now differ from
the tip of trunk in `original.fossil`: a file changed, added or removed. Fossil's own
files (`checkpoint.fossil`, `original.fossil`, the checkout database) and the files the
checkpoint repository ignores do not count. An app without `original.fossil` was not
downloaded, and is never changed.

App-console's check before this (`fossil diff --from original`) never worked: fossil
cannot resolve the name `original`, and the error was discarded, so every downloaded app
read as unchanged.

## Status

`Status` reports an app's count, whether it was downloaded, and whether it has changed
since install. It is what a display polls, so it must be cheap and harmless:

- It reads the repositories directly, naming each one (`-R`), so fossil cannot walk up
  into an enclosing checkout.
- It never opens, verifies or configures a checkout, and writes nothing. (`Count` goes
  through the checkout and reapplies its settings, a write, on every call.)
- An app without a repository has a count of 0.

## Install

`Install` marks an app as installed from a download: it sets a baseline (as `checkpoint
baseline` does), then copies `checkpoint.fossil` to `original.fossil`. If the baseline
fails, nothing is copied; if the copy fails, no partial `original.fossil` is left behind.

`original.fossil` is in the checkpoint ignore-glob, so a later save never commits it into
the app's checkpoints. (Before, it was not, and the first save after an install committed
it.)

## Lua

The host gives each session's Lua two asynchronous operations, built as host operations
(ui-engine's `specs/host-ops.md`): each returns at once, the fossil work runs off the
session's executor, and the result lands in a field the caller names. Both are defined
where the host builds the session's `mcp` global, so ark gets them through flib.

```lua
ui.checkpointStatus(target, field, names)       -- target[field] = {name = {count, downloaded, changed}}
ui.checkpointInstall(target, field, name, cb)   -- target[field] = true once installed
```

- `names` is a list of app names under the host's apps directory. The result has one
  entry per name: `count`, `downloaded` and `changed`, or `error` with the message when
  that app's status could not be read. One unreadable app does not hide the others.
- `checkpointInstall` runs `Install` for one app; its callback is how the install flow
  continues.
