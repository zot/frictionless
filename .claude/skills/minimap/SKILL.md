---
name: minimap
description: The mini-spec structural model in miniature — the 3-level spec→design→code layout, where artifacts live, the root spec index, and the traceability links that connect them. Load to navigate or search a mini-spec project across specs, design, and code (the model, not the maintenance process). Foundation for mini-spec.
---

# Minimap

How a mini-spec project is laid out, so you can find your way around it —
read it, search it, follow a clue from any layer to the others. This is the
**model**, not the process. Load `/minimap` to *navigate*; load `/mini-spec`
to *maintain* (it loads this skill underneath).

## Design docs first — not code

When understanding a feature or planning a change, start with
`design/design.md` and the relevant CRC cards/sequences **before** using
code-exploration tools (grep, symbol search, reading source). Design docs are
the project index — they show component relationships, responsibilities, and
code-file mappings more efficiently than reading code. Drop into code-level
tools only after the design docs have oriented you. And when the project keeps
a **root spec index** (below), read that *first of all* — it maps every spec
in one file, so you orient without grepping the `specs/` tree.

## The three levels

```
specs/                # Human intent — current state, in the user's own words
  migrations/         # In-flight migrations — temporary, moved on completion
    complete/         # Completed migrations, numbered in landing order
design/               # AI translation — requirements.md, crc-*, seq-*, ui-*, test-*, manifest-ui.md
docs/                 # user-manual.md, developer-guide.md
src/                  # Code with traceability comments
```

(`src/` stands in for the project's code root, whatever it's named.)

**Specs** are the human's voice — what the system should do, in natural
language, organized by feature area. Intent, not implementation. For
libraries, API signatures live in specs because they *are* the contract.
Most specs are **per-feature** (one spec, one capability). A second kind,
**summary specs**, indexes existing behavior along a cross-cutting axis
without adding behavior of its own (below).

**Design** is the AI's translation of specs into buildable structure.
`requirements.md` extracts testable statements; CRC cards assign
responsibilities to components; sequences show how components interact.

**Code** implements the design, carrying traceability comments that link each
piece back to the design artifacts and requirements that justify it.

## The root spec index

A project with many specs keeps a **root index** — one file (e.g.
`specs/index.md`) that maps every per-feature spec by **system**, registers
the **summary specs**, and names the **cross-cutting themes**. The per-feature
specs are the leaves; the index is the root that says where to look.

**Use it to orient cheaply.** Read the root index *first* — at session start,
or before working in an unfamiliar area — then open only the specs it points
you to. One small read replaces grepping the whole `specs/` tree or opening
files just to learn what they cover; orientation drops from O(corpus) toward
O(1) tokens. This is the spec-level companion to *Design docs first*.

Entries are *pointers, not copies* — the named spec stays canonical, the index
mirrors it. A **theme** gathers every spec a cross-cutting concern touches and
states its invariant once, so contradictions between leaves become visible.
(Keeping the index complete is a maintenance task — `/mini-spec` owns it.)

## Summary specs

A **summary spec** doesn't introduce behavior — it *indexes* behavior owned by
per-feature specs, along one cross-cutting axis. Per-feature specs answer "what
does this feature do?"; summary specs answer "what's the full set of X in this
project?" Recurring kinds:

- **CLI inventory** — every subcommand and flag (e.g. `specs/cli-commands.md`).
- **Storage layout** — every record class with key/value layout (e.g. `specs/record-formats.md`).
- **API surface** — every public binding exposed to a scripting/extension layer (e.g. `specs/lua-api.md`).
- **Capabilities** — every named feature with motivation and objective (e.g. `specs/features.md`).

For navigation: **per-feature specs are canonical; summary specs are mirrors.**
When the two disagree, trust the per-feature spec. (Keeping the mirrors in sync
is a maintenance task — `/mini-spec` owns it.)

## Traceability — the links you follow

The three levels are stitched together by a chain of grep-able references.
This is what lets you start from a vague clue in any layer and follow it to the
others — the core navigation move.

```
spec item ──> requirement (Rn) ──> design artifact (CRC card / sequence) ──> code (// CRC:/Seq:/Rn)
```

- **Requirements** are numbered, testable statements — `R1`, `R2`, … —
  globally numbered across all features, living in `design/requirements.md`.
  Each spec item becomes exactly one `Rn`.
- **CRC cards** (`design/crc-*.md`) assign responsibilities to a component:
  a `## Knows` / `## Does` / `## Collaborators` body, with a top-line
  `**Requirements:** R1, R3, R7` naming the requirements it covers.
- **Sequences** (`design/seq-*.md`) show component interactions. Steps are
  numbered with dotted IDs (`1.4`), so code can pin to a specific step.
- **Code** carries traceability comments naming the artifacts that justify it:

  ```
  // CRC: crc-Store.md | Seq: seq-crud.md#1.4 | R4, R5
  ```

  The `| Rn, Rn` tail is optional but common — it links a specific code
  location straight to requirements.

**Following a thread (both directions):**

- Requirement → everywhere it lands: `grep -rn "R5" design/ src/`
- Sequence step → its implementations: `grep -rn "seq-crud.md#1.4" src/`
- Code → its justification: read the `// CRC:/Seq:/Rn` comment, open that
  artifact, follow its `**Requirements:**` line back to the spec.
- Spec → its code: spec → find the `Rn` in `requirements.md` → grep that `Rn`.

The links are bidirectional by design: every artifact points down toward code
and the comments point back up toward intent, so a clue in one layer reaches
all three.

**Retired requirements** still resolve. A requirement obsoleted by a later
change is rewritten in place as `**~~R5:~~** (Retired Tn — see R10) <original
text>`, so an old reference in design or code still leads somewhere — to the
replacement (or to "no replacement"). Don't read a struck-through `Rn` as a
dead end; it's a forwarding address.