# Minispec Configuration Reference

## The two config files

Configuration lives at two scopes, matching the project's two roots.

| file | scope | holds |
|---|---|---|
| `<repo root>/.minispec/config.toml` | the repository | `track`, plus any settings shared by every design root in it |
| `<design root>/.minispec.toml` | one design root | only what differs from the repository config |
| `<repo root>/.minispec.toml` | — | **an error, with no exception** |

The **design root** is the directory containing `design/`. The **repository root** is
the top of the version-controlled tree, which owns `.claude/`, `carves/` and the
trajectory files. They are often the same directory; a repository may hold several
design roots under one repository root.

**A top-level `.minispec.toml` is always an error**, because its inheritance would be
incoherent — it would have to inherit from `.minispec/config.toml`, a file inside its
own directory. Where the repository root *is* a design root, `.minispec/config.toml`
is that design root's configuration too; there is no second file. A symlink from the
old path to the new one is still the forbidden shape, not a supported alias.

### Inheritance

Settings resolve in three layers, each applied over the one before: built-in defaults,
then the repository config, then the design root's own file.

- **Scalars replace** — `design_dir`, `src_dir`, `track`.
- **Lists merge as a union between configuration layers** — repository entries kept,
  design-root additions appended, duplicates dropped, repository order preserved. The
  **first configuration layer to set a list replaces the built-in defaults** rather
  than adding to them, so a shipped list can still be narrowed.

The rule each layer follows is the same one: **state only what you add or change,
never what you keep.** A design root whose settings match the repository needs no file
at all.

A design root can *add* to an inherited list but cannot *remove* from one. The fix is
to drop the setting from the repository config and state it in each design root that
needs it — moving the decision down a level rather than overriding it from below.

Run `minispec query config` to see every effective setting with the file it came from.

## `track` — the repository-scoped setting

`track` lives in `.minispec/config.toml` and **only** there. A design root that sets it
is an error: it describes the repository, and a repository may hold several design
roots, so one of them cannot answer for the whole.

| value | meaning |
|---|---|
| `none` | the project is not under version control |
| `private-trajectory` | git; the trajectory files are ignored |
| `all` | git; the trajectory files are tracked with everything else |

The paths it governs are `PENDING.md`, `CURRENT.md` and `DONE.md` at the repository
root, plus `.carves/` whenever it exists. `.carves/` must be ignored under **both**
git-managed values — a directory whose whole purpose is privacy does not become public
because the queue did. Public `carves/` is never ignored.

**`track` is verified on every run, not merely stored.** The tool checks it against two
facts — whether the tree is git-managed, and the actual ignore state of those paths —
and refuses when they disagree. It duplicates what `.gitignore` says on purpose: a copy
re-checked every run cannot drift undetected, which is the whole reason to keep it.

### Creating and repairing it

`minispec init` is the **sole creator** of `.minispec/config.toml`, and one
`--track-<style>` flag is **mandatory** — the value cannot be inferred, and a default
would make the choice for someone without their noticing.

```
minispec init --track-none
minispec init --track-private-trajectory
minispec init --track-all
```

`init` also writes the ignore lines the value implies: `.minispec/backup` under any
git-managed value, and the trajectory filenames under `private-trajectory`.

To change the value later, or to fix a mismatch:

```
minispec init --track-<style> --repair
```

`--repair` is **symmetric**: it sets `track` *and* brings `.gitignore` into agreement,
adding lines the new value requires and removing lines it forbids. Those are opposite
outcomes, and the flag value is how you choose between them — **confirm the value with
the user before running it.**

Plain `init` refuses when `.minispec/` exists; `--repair` refuses when it does not.
The precondition is inverted rather than supplemented, so neither has to guess.

### With no configuration

Only `init`, `--version`, `help` and `check-version` run without a
`.minispec/config.toml`. Everything else refuses with a message telling the agent to
check whether the directory is a code project, ask the user whether they want one, and
then **report and wait** — running `init` is the user's decision, never the agent's.

Every project that predates `track` meets that refusal once. That is how it gets asked.

### If the config is malformed

`--repair` sets a value; it cannot undo arbitrary damage, so it refuses a file that
will not parse. **That refusal is the one case where an agent is authorised to edit the
configuration by hand.** Back the file up first, and skip the backup when it would be
byte-identical to one already there.

## Format

Both files are **TOML**, decoded strictly. All settings are optional; an omitted one
inherits, per the rules above.

- **An unknown key is an error** naming the file and the key, and the command stops. A
  setting the tool does not read would otherwise be a line with no effect and no way to
  find out.
- **A YAML configuration is an error.** The format was YAML until 2026-09-25. A
  `.minispec/config.yaml` or `.minispec.yaml` is named with the instruction to convert it
  to TOML by hand, and nothing else runs until it is. The keys keep their names:

  | YAML | TOML |
  |---|---|
  | `track: all` | `track = "all"` |
  | `src_dir: lib` | `src_dir = "lib"` |
  | `code_extensions: [.go, .ts]` | `code_extensions = [".go", ".ts"]` |

## Schema

```toml
design_dir = "design"               # path to the design directory, relative to the design root
src_dir = "src"                     # path to the source directory, relative to the design root
code_extensions = [".go", ".ts"]    # file extensions to scan for traceability comments
track = "private-trajectory"        # repository config only: none, private-trajectory or all
```

## Comments in code files

Code files are read through a **language table** chosen by extension: Go, JavaScript,
TypeScript, Lua, shell, Python, Pascal, C, C++, Java, Emacs Lisp, HTML (with the script and
style inside it), Markdown and CSS are built in. A file whose extension has no table is
reported as not read. `comment_patterns` and `comment_closers` are **retired**; a
configuration that still sets either is refused, naming the key as retired. How a
traceability comment is written in each extension is what the tool reports:

```bash
~/.claude/bin/minispec query comment-patterns
```

**Block-comment extensions need their closer, every time.** Where the report shows a
closer (` -->` for `.md` and `.html`, ` */` for `.css`, ` *)` for Pascal), append it to every traceability
comment you write. An unclosed block comment silently swallows all the code after it
until the next accidental closer, which is catastrophic and very hard to diagnose.

## Languages

To read an extension no table covers, or to change how one is read, define the language in
either configuration file. A definition is sdom's bracket table written field for field in
snake case, and overrides the built-in table for the extensions it names:

```toml
[[languages]]
name = "zig"
extensions = [".zig"]
comment = { prefix = "// ", suffix = "\n", kind = "comment" }   # the form written; required

  [[languages.brackets]]          # groups in MATCHING order: longer markers first
  open = ["//"]
  close = "\n"
  allowed_inner = []              # [] = raw: nothing inside is recognized
  kind = "comment"                # the same kind as `comment`, so it is read as one

  [[languages.brackets]]
  open = ['"']
  close = '"'
  escape = '\'
  allowed_inner = []

  [[languages.brackets]]          # no allowed_inner = code mode
  open = ["{"]
  close = "}"
```

- **`allowed_inner` absent is code mode, `[]` is raw**, and a list names the only openers live
  inside. Comments and plain strings are raw; brackets are code.
- **Order is matching order**: the first opener that matches wins, so `"""` comes before `"`.
- **A design root's definition replaces a repository one of the same name whole**, and adds
  one of a new name.
- **Every definition is checked when the configuration loads**; a malformed one is an error
  naming the file, the language and the group.

`languages-example.toml` in this skill directory defines C, C++, Java, Go and Python exactly
as the built-in tables read them, with every field explained. Copy from it.

## Precedence

1. **CLI flags** (`--design-dir`, `--src-dir`) override everything
2. **`<design root>/.minispec.toml`** — the design root's own settings
3. **`<repo root>/.minispec/config.toml`** — settings shared across the repository
4. **Built-in defaults** apply when no layer sets a value

Layers 2 and 3 merge rather than simply overriding: scalars replace and lists union.
See *Inheritance* above.

## Verifying

After creating or updating either config file, run:

```bash
~/.claude/bin/minispec query config             # every setting, with the file it came from
~/.claude/bin/minispec query comment-patterns   # how to write a comment per extension, closers included
~/.claude/bin/minispec query project            # the two roots, and the resolved paths
```

`query config` is the one to reach for when a value is not what you expected — the
effective configuration is not what any single file says, so it reports each setting's
origin rather than leaving you to read two files and know the precedence by heart.
