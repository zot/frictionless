---
name: mini-spec
description: "**MANDATORY: Invoke BEFORE writing or modifying any code.** Without `design/design.md` you lack critical knowledge of component relationships, responsibilities, and code mappings — changes made without this context risk breaking architectural invariants. Do NOT explore code with other tools first. Load this skill before doing anything else."
---

# Mini-spec

## Design Docs First — Not Code

**When understanding a feature or planning a change, start with `design/design.md` and the relevant CRC cards/sequences BEFORE using code exploration tools.** Design docs are the project index — they show component relationships, responsibilities, and code file mappings more efficiently than code search. Only drop into code-level tools (Serena, Grep, etc.) after the design docs have oriented you.

## Prerequisite: Version Check and Comment Patterns

**First**, run `~/.claude/bin/minispec check-version` to verify the tool is installed and matches this skill's version. If it fails, warn the user: the tool and skill must be the same version or there will be compatibility issues.

**Then**, run `~/.claude/bin/minispec query comment-patterns` to learn the recognized comment patterns for traceability comments in code files.

## MANDATORY: Create Tasks First

**BEFORE reading any files or doing any work**, create tasks for applicable phases:

```
TaskCreate: "Spec Phase: [feature name]"
TaskCreate: "Requirements Phase: [feature name]"
TaskCreate: "Design Phase: [feature name]"
TaskCreate: "Implementation Phase: [feature name]"
TaskCreate: "Simplification Phase: [feature name]"
TaskCreate: "Gaps Phase: [feature name]"
```

Do NOT proceed until tasks exist. This is required for user visibility into progress.

## MANDATORY: Check for In-Flight Migrations

Before any phase, run `~/.claude/bin/minispec query migrations`. It
lists in-flight migration specs (the `*.md` files in
`specs/migrations/`, excluding `complete/`). Each file is an
in-process migration — record formats, APIs, or internal structures
are mid-flux. If any are present:

- Surface them to the user before doing other work.
- In-flight migrations take priority. Do not start unrelated changes
  that touch the same code paths until the migration is complete.
- If your task IS the in-flight migration, proceed.

Migrations are temporary by design — see "Migration Workflow" below.

---

## Overview

3-level architecture: specs → design → code.

```
specs/                # Human intent — current state of the project, in the user's own words
  migrations/         # In-flight migrations — temporary, get moved on completion
    complete/         # Completed migrations, numbered in landing order
design/               # AI translation — requirements.md, crc-*, seq-*, ui-*, test-*, manifest-ui.md
docs/                 # user-manual.md, developer-guide.md
src/                  # Code with traceability comments
```

### What each level is for

**Specs** are the human's voice. They describe what the system should
do in natural language, organized by feature area. The user writes or
approves them. They communicate intent — not implementation, not
internal structure. A spec should read like someone explaining the
feature to a colleague. For libraries, API signatures belong in specs
because they *are* the face of the project — they must be agreed upon
before design begins. Specs must state the language and environment
so the AI knows what it's building for.

Most specs are **per-feature** — one spec, one capability, one cohesive
slice of behavior. A second kind, **summary specs**, indexes existing
behavior along a cross-cutting axis (CLI surface, storage layout, API
set, capabilities) without introducing any new behavior of its own.
See *Summary specs* below.

**Design** is the AI's translation of specs into buildable structure.
Requirements extract testable statements. CRC cards assign
responsibilities to components. Sequences show how components
interact. The user reviews this translation before code is written —
catching a misunderstanding here costs minutes, not hours.

**Code** implements the design with traceability comments linking back
to the design artifacts that justify each component's existence.

### Why this matters

Each level exists because skipping it has a concrete cost:

- **Verification** — Design is smaller than code. The user can confirm you understood the task *before* you write hundreds of lines.
- **Preview** — The design tells the user what you're about to change. Without it, they discover unwanted modifications after the fact.
- **Reference** — During implementation, you look up the design instead of re-reading all the code. This keeps changes consistent across files.
- **Anchor** — Without a design document, iterative modifications cause **drift**: features silently disappear as code evolves across sessions. The design pins what must survive.
- **Traceability** — The specs→requirements→design chain ensures nothing is lost between what the user asked for and what gets built. When something breaks, you can trace backward to find out why.

The phases are not ceremony. They are cheaper than debugging a misunderstood requirement after 500 lines of code.

### Summary specs

A **summary spec** is a spec that doesn't introduce behavior — it
*indexes* behavior owned by per-feature specs, along one cross-cutting
axis. Per-feature specs answer "what does this feature do?"; summary
specs answer "what's the full set of X in this project?"

Examples that recur across projects:

- A **CLI inventory** spec lists every subcommand and flag (e.g.
  `specs/cli-commands.md`).
- A **storage layout** spec lists every record class with key/value
  layout (e.g. `specs/record-formats.md`).
- An **API surface** spec lists every public binding exposed to a
  scripting or extension layer (e.g. `specs/lua-api.md`).
- A **capabilities** spec lists every named feature with motivation
  and objective (e.g. `specs/features.md`).

When to create one:

- A question of the form "what's the full set of X across this
  project?" keeps coming up, and answering it requires touching many
  per-feature specs.
- A cross-cutting axis has enough items that someone (or some future
  you) would want a directory to navigate them.

Maintenance rules:

- **Per-feature specs are canonical; summary specs are mirrors.**
  When the two disagree, the per-feature spec wins. Update the summary
  to match.
- **Per-feature anchoring does not maintain summary specs.** Mini-spec's
  normal anchoring (specs → requirements → design → code) catches the
  per-feature edits but cannot tell that a CLI-inventory or
  capabilities spec should also have been updated. Updating summary
  specs is the maintainer's job, performed explicitly.
- **Pin the summary-spec list somewhere persistent** — typically
  CLAUDE.md or the project's top-level reference doc — so a future
  agent or maintainer knows which summary specs to keep in sync when
  they add, rename, or retire something along the relevant axis.
- **Don't anchor new requirements from a summary spec.** Rn numbers
  belong to the per-feature spec that owns the behavior. A summary
  spec entry references that spec; it does not own the contract.

## Task Tracking

**During implementation**, break down into per-file tasks:
```
TaskCreate: "Implement view.ts changes"
TaskCreate: "Implement viewlist.ts changes"
TaskCreate: "Update design docs"
```

**Mark phases complete** with TaskUpdate as you finish them.
**Use Quality Checklist items** as tasks before finalizing.

## Core Principles
- use SOLID principles, comprehensive unit tests
- when adding code, verify whether it needs to be factored
- Code and specs as MINIMAL as possible
- Before using a callback, see if a collaborator reference would be simpler
- write idiomatic code for the language you use
- avoid holding locks in sections that have significant functionality
- **No unanchored design:** every design artifact must trace back to a spec item and requirement. If you need to add something to the design, add it to specs first, then requirements, then design. This applies regardless of direction — even when documenting existing code, verify the spec anchor exists before updating design. This prevents features from existing only in the AI's interpretation.
- in HTML, use the slimmest DOM Possible. Fewer elements makes everything in the browser better: less memory, more speed, better responsiveness

### Why anchoring matters

Specs and design docs are the project's memory bank. AI context dies every
session — code changes compound across sessions without any single agent
seeing the full history. Unanchored code has no justification trail: a
future session can't tell whether a function was designed or accidental,
required or leftover. When that session makes changes, unanchored features
silently disappear because nothing in the design said they should exist.

Anchoring is cheap (a few lines of spec + a requirement number). The cost
of *not* anchoring is discovering, three sessions later, that a feature
vanished during an unrelated refactor and no one noticed because the design
never mentioned it. The spec is the pin that says "this must survive."

## Cross-cutting Concerns

`design.md` Cross-cutting Concerns section: Patterns spanning components (auth, errors, logging, routing, theming).
Referenced from other design artifacts: Cards, sequences, and layouts can all say "see cross-cutting: auth"

## Traceability

`design.md` Artifacts section: design files with code file checkboxes.

**Use minispec commands for checkbox operations:**
```bash
# View current artifact states
~/.claude/bin/minispec query artifacts

# Before modifying code: uncheck the artifact
~/.claude/bin/minispec update uncheck design.md crc-Store.md

# After implementation matches design: check the artifact
~/.claude/bin/minispec update check design.md crc-Store.md
```

**Code changes:** Uncheck artifact, ask user: "Update design, specs, or defer?"
**Update design:** Read code, update design file, re-check artifact.

## Workflow

**First:** Read specs. Specs must indicate language/environment.

**Then:** Proceed through phases

1. Spec Phase
Create in `specs/`: human-readable descriptions organized by feature
area. Specs are the user's intent in their own words. For applications,
this means behavior and user-facing concepts. For libraries, include
the public API signatures — they are the contract that design must
satisfy. Do not include internal structure or implementation choices.

**Upon completion**, run `~/.claude/bin/minispec phase spec` to verify spec files exist, then offer Requirements Phase. Do not jump to Design.

2. Requirements Phase
Create `design/requirements.md`: merge all specs into numbered requirements.

Format:
```markdown
# Requirements

## Feature: [feature-name]
**Source:** specs/feature.md

- **R1:** [requirement from spec]
- **R2:** [requirement from spec]
- **R3:** [inferred requirement - marked as such]

## Feature: [another-feature]
**Source:** specs/another.md

- **R4:** [requirement]
```

Guidelines:
- Each spec item becomes exactly one numbered requirement (R1, R2, ...)
- Numbering is global across all features (not per-feature)
- Mark inferred requirements explicitly: "**R5:** (inferred) ..."
- Keep requirement text atomic and testable

**Upon completion**, run `~/.claude/bin/minispec phase requirements` to verify format, then offer Design Phase. Do not jump to Implementation.

3. Design Phase
Create in `design/`:
- `design.md`: Intent + Artifacts (design files → code file checkboxes)
- `crc-*`: CRC cards (see format below)
- `seq-*`: sequence diagrams (≤150 chars wide; number their steps — see "Numbered Sequence Anchors" below)
- `ui-*`: ASCII layouts, reference CRC cards
- `test-*`: test designs (see format below)
- `manifest-ui.md`: routes, theme, global components

**Design Traceability:** All design artifacts must reference requirements:
```markdown
# ClassName
**Requirements:** R1, R3, R7
```

Use minispec to add requirement references:
```bash
~/.claude/bin/minispec update add-ref crc-Store.md R5
```

**Artifacts Format** (must be exact for `minispec` tool parsing):
```markdown
## Artifacts

### CRC Cards
- [x] crc-Store.md → `src/store.ts`
- [x] crc-View.md → `src/view.ts`, `src/viewlist.ts`

### Sequences
- [x] seq-crud.md → `src/store.ts`, `src/view.ts`

### UI Layouts
- [ ] ui-dashboard.md → `web/html/dashboard.html`

### Test Designs
- [ ] test-Store.md → `src/store_test.ts`
```
The Artifacts section is a **manifest of all design files** except design.md and requirements.md. Every crc-*, seq-*, ui-*, test-*, and manifest-*.md must be listed.

Format rules:
- Section headers (`### CRC Cards`, etc.) are optional grouping
- Each line: `- [x] design.md → code-file(s)` or `- [ ] design.md`
- Multiple code files: comma-separated after `→`
- Backticks around code paths are optional
- Checkbox state applies to all code files on that line

**Numbered Sequence Anchors:** Number the steps in your sequence diagrams using dotted notation so code can pin to specific steps. Place the number wherever the diagram style allows:

- Tree/outline: `1.4. step description` on the line itself
- UML actor-lane: `1.4` on its own line directly above the arrow
- Mermaid/pseudo-Mermaid: `1.4` at the start of the step

A file may contain more than one numbered diagram. Items in the first numbered diagram begin with `1.`, the second with `2.`, and so on (`1`, `1.1`, `1.1.1`, `2`, `2.1`, ...). The first segment K is the diagram index. Numbers are local to the file: `1.4` in seq-foo.md is unrelated to `1.4` in seq-bar.md.

Reference a numbered step from code with `Seq: seq-foo.md#1.4`. File-only refs (`Seq: seq-foo.md`) remain valid for diagrams that aren't numbered.

**Why number:** the anchor creates a bidirectional, grep-able link.
- Agent generating code: drop `seq-foo.md#1.4` in a traceability comment as a promise that this code implements that step.
- Agent making a code change: follow the anchor to verify what the diagram says the step does.
- Human reading code: `grep "seq-foo.md#1.4" src/` finds every implementation of that step.

For this to work, the number must be uniquely findable in the diagram source (avoid prose that starts with dotted numbers at the same indentation). Within a single file, every dotted ID may appear at most once. Append new steps with new numbers; renumbering existing steps orphans the code that pins to them — same discipline as Rn IDs.

The validator checks per-K tree contiguity (under K.x, children must be K.x.1, K.x.2, … with no gaps), K-sequence contiguity within the file (Ks are 1, 2, 3, …), and intra-file ID uniqueness. Unnumbered seq files are silently skipped — numbering is opt-in per file.

**Upon completion**, run `~/.claude/bin/minispec phase design` to verify coverage, then offer Implementation Phase. Do not jump to Gaps.

4. Implementation Phase
Add traceability comments with optional inline requirement refs:
```
// CRC: crc-Store.md | Seq: seq-crud.md#1.4 | R4, R5
add(data): Item {
```

The third `| Rn, Rn` section is optional but recommended — it links specific code locations directly to requirements, enabling implementation coverage validation.

**Block-comment languages:** The `minispec query comment-patterns` output lists any `comment_closers`. If a closer exists for the file extension, you MUST append it to every traceability comment. An unclosed block comment silently swallows all subsequent code. See `config-reference.md` (in this skill directory) if you need to configure closers for a new language.

Mark implemented using minispec:
```bash
~/.claude/bin/minispec update check design.md crc-Store.md
```

Look out for language-specific "gotchas" like mixing functions and methods in Lua.

**Upon completion**, run `~/.claude/bin/minispec phase implementation` to verify traceability, then run the Simplification Phase.

5. Simplification Phase
Invoke the `code-simplifier` agent on the recently modified code. This refines code for clarity, consistency, and maintainability while preserving functionality.

**CRITICAL: Preserve all traceability comments.** The `// CRC:`, `// Seq:`, and requirement references (`R123`) in code comments are load-bearing — they connect code to the design artifacts that justify its existence. Removing or reformatting them breaks the traceability chain that `minispec validate` checks. Simplification means cleaner *logic*, not fewer comments.

**Upon completion**, proceed to Gaps Phase.

6. Gaps Phase

**Traceability Verification:**

Run `~/.claude/bin/minispec phase gaps` to validate the gaps section, then run `~/.claude/bin/minispec validate` for full coverage check:

1. **Specs ↔ Requirements:** Each spec item maps to exactly one requirement in `requirements.md`
2. **Requirements ↔ Design:** Each requirement is referenced by at least one design artifact
3. **Requirements ↔ Code:** Each requirement appears as an inline Rn ref in at least one code file

`design.md` Gaps section tracks (use S1/R1/D1/C1/I1/O1/A1/T1 numbering):
- **Spec→Requirements (Sn):** Spec items not captured in requirements.md
- **Requirements→Design (Rn):** Requirements without design artifacts referencing them
- **Design→Code (Dn):** Designed features without code
- **Code→Design (Cn):** Code without design artifacts
- **Implementation (In):** Requirements with design coverage but no inline Rn ref in any code file
- **Oversights (On):** Missing tests, tech debt, enhancements, security concerns, etc.
- **Approved (An):** Approved gap. Permanent — written without a checkbox.
- **'Tired (Tn):** Retired requirement — obsoleted by a later change. Each Tn names the original Rn, the replacement Rn (or "no replacement" if removed outright), and the reason (usually a migration or refactor). Retired Rn entries stay in requirements.md with their original text but get a `~~Rn:~~ (Retired Tn — see Rxxx)` marker so old design/code references still resolve. Permanent — written without a checkbox.

Nest related items with checkboxes (only S/R/D/C/I/O take checkboxes; A and T are permanent and never carry one):
```markdown
- [ ] R1: Requirement R5 has no design artifact
- [ ] O1: Test coverage gaps
  - [ ] Feature A (5 scenarios)
  - [ ] Feature B (3 scenarios)
- A1: Dangling methods, these are never called
  - Maluba.go: Maluba.Frobnicate, Maluba.Enreify
- T1: R1598 retired by R1833 (2026-04-23 ec-rekey)
  - reason: EC keys moved from (fileID, chunkIdx) to chunkID
- T2: R1099 retired by R1281 (2026-04-09 tag-embeddings)
  - reason: V key gained trailing tvid varint
```

If you encounter legacy `- [ ] An:` lines, drop the `[ ]` —
`minispec validate` reports them as `permanent gaps with checkbox`.

Use `minispec update add-gap` to add gaps; it writes the right
shape automatically (no checkbox for A/T, checkbox for the rest).

**Upon completion**, offer to update Documentation (Documentation Phase).

7. Documentation Phase, Optional -- offer to user after Gaps
Create `docs/user-manual.md` and `docs/developer-guide.md` with traceability links.

## Migration Workflow

Specs in `specs/` describe how the system *is* — they're the
canonical "current state." Migration specs describe how to get from
state A to state B. They have a built-in expiration: once
implemented, the "Problem" they describe no longer exists.

To keep `specs/` from accumulating stale migration narratives:

1. **Create migration specs in `specs/migrations/`**, not in
   `specs/`. One file or several — one per coherent migration.
2. **Run the mini-spec phases** on the migration specs as normal
   (Spec → Requirements → Design → Implementation → Simplification
   → Gaps).
3. **When implementation lands**, the migration is complete. The
   code now embodies state B.
4. **Update the affected `specs/*.md` files** to describe state B
   as the current truth — fold in record formats, API contracts,
   or other steady-state material that the migration changed.
5. **Retire obsoleted requirements.** For each obsolete Rn run:

   ```
   ~/.claude/bin/minispec update retire R<old> R<new> "<reason>"
   ```

   Use `-` instead of `R<new>` if there is no replacement. The
   command rewrites the R<old> line in `requirements.md` to
   `**~~R<old>:~~** (Retired Tn — see R<new>) <original text>` AND
   appends a new Tn entry to `design.md` Gaps in one atomic step.
   Outputs the assigned Tn.

   If a CRC card or inline code comment still references the
   retired Rn but the code no longer fulfills it, update the
   reference to the replacement Rn. (References to retired Rn in
   code that was removed are fine — the comment went with the
   code.)

6. **Move the migration spec(s)** by running:

   ```
   ~/.claude/bin/minispec update migration-complete <name>
   ```

   The command moves `specs/migrations/<name>.md` to
   `specs/migrations/complete/<NNN>-<name>.md` where NNN is the
   next zero-padded three-digit prefix. Numbers are assigned at
   completion time, not creation time, so concurrent in-flight
   migrations don't fight over numbers and the prefix reflects
   actual landing order. Outputs the new path.

`specs/migrations/complete/` is the migration history — a
chronological record of what changed and why. `specs/` always
reflects the present.

## CRC Card Format
```markdown
# ClassName
**Requirements:** R1, R3, R7

short description

## Knows
- attribute: description
## Does
- behavior: description
## Collaborators
- OtherClass: why
## Sequences
- seq-scenario.md
```
Principles: Single Responsibility, minimal collaborations, PascalCase.

## Test Case Format
```markdown
# Test Design: ComponentName
**Source:** crc-ComponentName.md
## Test: name
**Purpose:** what this validates
**Input:** setup and data
**Expected:** verifiable outcome
**Refs:** crc-*.md, seq-*.md
```
Cover: happy path, errors, edge cases.

## Quality Checklist
- [ ] Requirements: all spec items captured, numbered (R1, R2, ...), inferred items marked
- [ ] CRC Cards: nouns/verbs covered, no god classes, Requirements linked
- [ ] Sequences: participants from CRCs, ≤150 chars wide
- [ ] UI Specs: ASCII layouts, refs to CRCs and manifest-ui.md
- [ ] Traceability: design files in Artifacts, code files have checkboxes, all Rn referenced
- [ ] Tests: test-*.md for key behaviors
- [ ] Summary specs: any cross-cutting axis touched by this change has been mirrored in the relevant summary spec (CLI inventory, storage layout, API surface, capabilities, …) — see the project's pinned list
- [ ] Phase validation: `~/.claude/bin/minispec phase <phase>` passes after each phase
- [ ] Full validation: `~/.claude/bin/minispec validate` passes

## Minispec Tool

The `minispec` CLI tool (at `~/.claude/bin/minispec`) performs structural operations on design files.

**IMPORTANT:** Always use minispec commands instead of manual editing for:
- Checking/unchecking artifact checkboxes
- Adding requirement references to CRC cards
- Querying artifact states and coverage

```bash
# Version check (run on skill load)
~/.claude/bin/minispec check-version         # Verify tool and skill versions match

# Phase-specific validation (run after each phase)
~/.claude/bin/minispec phase spec            # Verify spec files exist
~/.claude/bin/minispec phase requirements    # Verify requirements format
~/.claude/bin/minispec phase design          # Verify design files and coverage
~/.claude/bin/minispec phase implementation  # Verify code traceability
~/.claude/bin/minispec phase gaps            # Verify gaps section

# Full validation
~/.claude/bin/minispec validate              # Run all validations

# Queries
~/.claude/bin/minispec query artifacts       # Show all artifacts with checkbox states
~/.claude/bin/minispec query uncovered       # List Rn without design refs
~/.claude/bin/minispec query gaps            # List gap items
~/.claude/bin/minispec query requirements    # List all requirements
~/.claude/bin/minispec query migrations      # List in-flight migration specs

# Updates - artifact checkboxes (in design.md)
~/.claude/bin/minispec update check design.md crc-Store.md     # Check artifact
~/.claude/bin/minispec update uncheck design.md crc-Store.md   # Uncheck artifact

# Updates - requirement references (in CRC cards)
~/.claude/bin/minispec update add-ref crc-Store.md R5          # Add requirement to CRC
~/.claude/bin/minispec update remove-ref crc-Store.md R5       # Remove requirement from CRC

# Updates - gaps (S/R/D/C/I/O get checkboxes; A/T are permanent)
~/.claude/bin/minispec update add-gap O "Test coverage needed" # Add oversight gap
~/.claude/bin/minispec update resolve-gap O3                   # Mark gap resolved
~/.claude/bin/minispec update approve-gap D3                   # Convert gap to approved (A) type
~/.claude/bin/minispec update add-gap T "R5 retired by R10"    # Add retired-requirement gap

# Updates - migrations
~/.claude/bin/minispec update retire R5 R10 "2026-04-27 schema-v2"   # Retire R5, replaced by R10
~/.claude/bin/minispec update retire R7 - "no replacement"           # Retire with no replacement
~/.claude/bin/minispec update migration-complete schema-v2           # Move spec to complete/ with NNN- prefix
```

Use the tool to:
- Run phase-specific checks after completing each workflow phase
- Verify design file formats are parseable
- Find uncovered requirements quickly
- Toggle checkboxes atomically (avoid manual checkbox edits)
- Add/remove requirement references to CRC cards
- Add gaps with auto-numbering
