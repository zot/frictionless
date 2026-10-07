---
name: mini-spec
description: "**MANDATORY: Invoke BEFORE writing or modifying any code.** Without `design/design.md` you lack critical knowledge of component relationships, responsibilities, and code mappings — changes made without this context risk breaking architectural invariants. Do NOT explore code with other tools first. Load this skill before doing anything else."
---

# Mini-spec

## Load the model first

**IMMEDIATELY invoke `/minimap` using the Skill tool before doing anything else.** It carries the structural *model* this skill builds on: the 3-level spec→design→code layout, what each level is for, where artifacts live, the **root spec index**, what a **summary spec** is, and the **traceability links** (`Rn` → CRC card → code `// CRC:/Seq:` comment) that stitch the levels together. Start at the design docs and the root index, not at code — drop into code-level tools (Serena, Grep, etc.) only after they've oriented you. This skill adds the **process** — phases, traceability *maintenance*, gaps, migrations, trajectory tracking — on top of that model.

## Prerequisite: Version Check and Comment Patterns

**First**, run `~/.claude/bin/minispec check-version` to verify the tool is installed and matches this skill's version. If it fails, warn the user: the tool and skill must be the same version or there will be compatibility issues.

**Then**, run `~/.claude/bin/minispec query comment-patterns` to learn how to write a traceability comment in each extension, closers included.

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

**If this harness has no task tool at all** — no `TaskCreate`, no `TodoWrite`, nothing
under any other name — the requirement does not lapse, it relocates: list the phases in
your response before starting, and name each one as you enter and finish it. What is
mandatory here is that the user can see which phase you are in. The task list is *how*,
not *what*, and a mandate with no defined outcome in the world the reader is standing in
gets ignored whole — along with the version check and the migration check either side of
it, which are real.

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

## Why the levels matter

(*The 3-level model itself — what specs, design, and code each are, and where
they live — is in `/minimap`. This skill is the **process** that builds and
maintains them.*)

Each level exists because skipping it has a concrete cost:

- **Verification** — Design is smaller than code. The user can confirm you understood the task *before* you write hundreds of lines.
- **Preview** — The design tells the user what you're about to change. Without it, they discover unwanted modifications after the fact.
- **Reference** — During implementation, you look up the design instead of re-reading all the code. This keeps changes consistent across files.
- **Anchor** — Without a design document, iterative modifications cause **drift**: features silently disappear as code evolves across sessions. The design pins what must survive.
- **Traceability** — The specs→requirements→design chain ensures nothing is lost between what the user asked for and what gets built. When something breaks, you can trace backward to find out why.

The phases are not ceremony. They are cheaper than debugging a misunderstood requirement after 500 lines of code.

## Summary specs — maintenance

(*What a summary spec **is**, and the recurring kinds — CLI inventory, storage
layout, API surface, capabilities — are in `/minimap`. This is the maintenance
side: when to create one, and how to keep it true.*)

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

**With no task tool**, the same breakdown and the same completions go in your responses.

## Core Principles
- use SOLID principles, comprehensive unit tests
- when adding code, verify whether it needs to be factored
- Code and specs as MINIMAL as possible
- Before using a callback, see if a collaborator reference would be simpler
- write idiomatic code for the language you use
- avoid holding locks in sections that have significant functionality
- **No unanchored design:** every design artifact must trace back to a spec item and requirement. If you need to add something to the design, add it to specs first, then requirements, then design. This applies regardless of direction — even when documenting existing code, verify the spec anchor exists before updating design. This prevents features from existing only in the AI's interpretation.
- **Supersede at the source:** the mirror of "No unanchored design." A change is complete only when every directive describing the *old* behavior is removed or rewritten **at its source** — across specs, requirements, AND design prose. Anchoring keeps features from vanishing; superseding keeps stale directives from causing reverts: a future agent reads a leftover spec sentence or design bullet as current intent and "fixes" the code back to match, undoing the change that obsoleted it. Completion test for any change: could an agent reading only specs + design be led to undo it? If yes, a trap remains.
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

### Why superseding matters

Anchoring and superseding guard opposite failure directions. Anchoring fights
*omission* — a feature with no spec silently disappears. Superseding fights
*contradiction* — a directive that outlived the behavior it described silently
reappears. The second is the more dangerous: a contradiction in the design is a
trap that springs in the *revert* direction. A retired requirement has a forcing
function (the `retire` command strikes it through, appends the Tn, and prints a
reconcile reminder), but the *prose* that spawned it — the originating spec
sentence, the CRC bullet, the sequence step — has none. It rots in place until a
future agent reads it as current intent and "fixes" the code back to match. So
retirement is not done when the `Rn` is struck out; it is done when every
sentence that described the old behavior is gone or rewritten at its source.

### Balance your backticks — and the tool reports it where it reads

**Delimit a backtick or fence mention with a run longer than the one inside it, padded with
spaces, and never let it begin a line.** A run at the head of a line opens a fenced block
however well balanced it is inline, so a correct escape lands wrong purely from where the line
wrapped. The same care applies to a `<!--` in prose: unclosed, it runs to the next `-->`,
which in a design document is usually a sequence-diagram arrow far below.

`validate trajectory` reports it for the files its readers touch — both queue files, the
current file, and every carve — as a coverage note naming the file and the opener's line
(R302): a group never closed takes the rest of its file with it, and every shape-based check
is blind to what it swallowed. Measured 2026-09-05 on this repository's done file: 17 entries
read where 58 existed, nothing said. **For `design/` and `specs/` the rule is still one you
remember** — the design-document readers do not report an unclosed run yet (gap `O22`); on
`old-sdom` they did, after one unclosed run hid 78 of 115 gap entries for a day.

### Prose reaches the tool through a file — and the tool enforces it where it can

**Every trajectory verb whose payload is prose takes a `--…-file` form**: `add-item
--title-file`, `--status-file` and `--next-action-file`; `start --context-file`; `finish
--body-file` and `--discharged-file`. Use them. A backtick inside a double-quoted shell
argument is **command substitution**, and when it fires the text is simply *gone* from what
the tool receives, with nothing reporting it. The `update` verbs that take prose — `add-gap`,
`retire` — still take it as an argument here (gap `O23`); write the text to a file and
`"$(cat file)"` it until they do.

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

**Reconcile the root spec index.** Whenever you add, rename, or retire a
per-feature spec, straighten out the root index (the project's
`specs/index.md`) in the same pass: create it if it doesn't exist yet, then
make sure every spec has an entry under a system, with new summary specs and
themes registered. Run `~/.claude/bin/minispec query unindexed-specs` — it
lists any per-feature spec missing from the index (the spec-level analog of
`query uncovered`); the pass is clean when that list is empty.

**Deleting a spec.** Removing a `specs/*.md` file orphans every requirement whose
`**Source:**` names it, and `validate` reports `missing spec sources`. Three
situations look identical from the error message and are repaired differently:

- **Renamed.** Rewrite the `**Source:**` lines to the new path. Nothing retires.
- **Merged into another spec.** Repoint the `**Source:**` at the absorbing spec.
  The behavior lives on, so nothing retires. This is the case most often
  mistaken for a deletion, and mis-handling it retires requirements that are
  still true.
- **Deleted outright — the behavior is gone.** Retire each of its requirements
  (`minispec update retire Rn - "<spec> deleted"`), then:
  1. Regroup them in `requirements.md` under a feature block whose `**Source:**`
     is `specs/deleted.md`. Keep one block per dead spec and name it in the
     heading — `## Feature: search (deleted)` — so provenance survives the move.
  2. Record the spec in `specs/deleted.md`: its name, a one-line description of
     what it covered, and the requirement numbers it owned.
  3. Index `specs/deleted.md` in the root index like any other spec, and honor
     the retirement's step-6 obligation to reconcile design prose at its source.

`specs/deleted.md` is a **tombstone registry**, not a spec — it describes nothing
the system does. It exists so a dead spec's requirements keep a `**Source:**` that
resolves, and so a reader meeting a struck-through `R40` in an old CRC card can
still learn what it was for. Completed migrations need no equivalent: their Source
resolves forward to `complete/NNN-<name>.md` on its own.

**In none of the three cases do you delete the requirement lines.** That is the
one repair the error message seems to invite and the one that must never be taken
— see "Rn numbers are permanent" in the Requirements Phase below.

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
- Numbers are permanent — never renumber, never reuse (see below)
- Mark inferred requirements explicitly: "**R5:** (inferred) ..."
- Keep requirement text atomic and testable

**Rn numbers are permanent — never renumber, never reuse.** An `Rn` is not a
position in a list. It is an identifier that CRC cards, sequence steps, and code
comments point at, and its meaning is whatever it meant when those pointers were
written.

Renumbering is the only edit in this system that breaks everything while leaving
every check green. Afterwards all the numbers still exist, so `unknown CRC refs`
finds nothing, coverage stays satisfied, `validate` passes — and every anchor in
`design/` and `src/` now cites a different requirement than its author meant.
There is no detection and no repair short of re-reading every reference in the
project. Compare a *missing* number, which is loud and fixable: this failure is
silent and permanent, which is why the rule is absolute rather than a preference.

- **Append only.** A new requirement takes the next free number: the maximum
  assigned anywhere, including retired ones. **Ask the tool —
  `minispec query next-id req` — do not grep for it.** Retired requirements keep their
  numbers, so a grep that skips them hands out one already taken, and it reports a bare
  number with no evidence of what it counted.
- **A gap in the sequence is a symptom, not a defect.** If `validate` reports
  `numbering gaps`, a requirement was deleted. The repair is to put it back
  — normally as a retirement — never to close the gap by shifting numbers down.
- **Do not delete a requirement; retire it.** `minispec update retire` keeps the
  number and its original text in place behind a forwarding marker, so every
  existing reference still resolves. Deletion is what creates the gap that
  tempts the renumber.

The same discipline governs sequence-step IDs, for the same reason — see
"Numbered Sequence Anchors" in the Design Phase.

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

**Where requirement refs count.** `minispec validate` computes
requirements→design coverage from each CRC card's **top-line
`**Requirements:**` field only** (plus approved gaps). Refs written
anywhere else in the card body — e.g. a per-method `(R5, R6)`
annotation on a `## Does` bullet — are documentation; the validator
does not parse them, so they earn a requirement no coverage. A
requirement counts as covered only when it appears in some artifact's
top-line field, which is what `add-ref` maintains. Body-level
annotations are fine as human notes, but never let them be the *only*
home for a ref. (Requirements→code coverage is separate: it comes
from inline `Rn` refs in code traceability comments. Retired
requirements are skipped by both coverage checks yet still resolve as
references, so a ref to a retired Rn is never flagged as unknown.)

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

**What counts as an inline `Rn` ref.** `minispec validate` reads every
code file in the Artifacts manifest through its **language table** and counts the refs of
every **traceability comment**: a comment whose whole interior is `|`-separated fields —
`CRC:`, `Seq:`, `Test:` and a requirement list, each at most once, in any order — then an
optional description after `--`, `—`, `:`, or a `.` directly after the requirement list.
So all of these count:

- `// CRC: crc-Store.md | Seq: seq-crud.md#1.4 | R4, R5` — the governing header
- `// Seq: seq-crud.md#2.2 | R6` — a step pin with its refs
- `// R7: why this line exists` and `// R8. Why this line exists` — a bare annotation
- `foo() // R9` — trailing, after code

**Ranges** expand into every member: `// R5-R8` counts R5 through R8, the second `R` is
optional (`R5-8`), and ranges mix with comma lists (`// R5-R7, R10`).

What does **not** count, because the comment is not wholly fields — it reads as prose:

- `// computed lazily (R5)` and `// see R5 for the rationale` — the ref does not lead
- `// R5 handles the retry` — a ref followed by prose **with no separator**. Write
  `// R5: handles the retry`. (The regex reader this replaced counted that form.)

A requirement annotated only in prose form reads as "missing impl coverage." Lead with the
ref and a separator, or fold it onto the governing `// CRC:` header (with the card that owns
the `Rn`, per its `**Requirements:**` field), to anchor the code location.

**The comment leader is the language's**, and `minispec query comment-patterns` lists it
per extension: the form to write, and every form read. The examples here are Go. HTML reads
its page comments, the `//` and `/* */` comments inside its `<script>`, and the `/* */`
comments inside its `<style>`, each only where it is live. A file whose extension has no
table is reported as not read, never passed over; a project can define more languages in
its configuration (see `config-reference.md` in this skill directory).

**Block-comment languages:** where `comment-patterns` shows a closer (` -->`, ` */`,
` *)`), you MUST append it to every traceability comment. An unclosed block comment
silently swallows all subsequent code.

Mark implemented using minispec:
```bash
~/.claude/bin/minispec update check design.md crc-Store.md
```

Look out for language-specific "gotchas" like mixing functions and methods in Lua.

**Codify what you verified — don't leave behavior hand-checked.** When you
implement a behavior and confirm it works (a live run, a smoke test, an ad-hoc
script), capture that verification as a `test-*.md` design + a test **in the same
pass**. A hand-check proves it works *today*; the test is what catches the
regression three sessions from now, when a future agent refactors the code that
made it pass. This is a **default action of the Implementation phase** — not a
Design-phase afterthought, and not something to defer to a Gaps-phase `O` entry.

**Then design the injection and write it down — but do not run it yet.** A regression
test written after its bug is fixed passes on its first run, which tells you the
property holds today and nothing about whether the test can detect its absence. So
name the defect to re-introduce, the site to introduce it at, and what red should
look like, and write them into the `test-*.md` as `**Fire alarm:**` / `**Inject:**` /
`**Code:**` (see Test Case Format). That is the part needing the code fresh in mind.
Recording it is not bookkeeping: the injection written down is what makes running it
minutes of work rather than a re-derivation nobody undertakes.

**Pull it once, and pull it after the Simplification Phase** — or at the end of this
phase when there is no simplification pass to run. **A proof earned against code that
is about to be rewritten is a rehearsal, not a proof.** *Measured 2026-08-21:* a pass
pulled sixteen alarms at the end of Implementation, the simplifier then restructured
most of the functions they named, and all sixteen had to be pulled again — **32
inject-run-restore-record cycles for sixteen proofs**, with the first sixteen records
void within the hour. Pulling is the dominant cost of a pass, and half of it was
buying a record that could not survive the next phase. *Nothing was lost by waiting,
and that was measured too:* the one thing the early pull caught — a case asserting an
exit code where it should have been reading a refusal — the later pull caught
identically, because it is the same injection either way.

**Which properties most need one: the ones that are invisible when violated.** A wrong
number is loud and any test catches it. A wrong *order* is silent — the work still
happens, the output still arrives, and the failure is intermittent and looks exactly
like working. The same goes for coalescing (the extra passes merely cost time),
allocation counts (the program is merely slower), "exactly one of these runs at a
time" (usually true anyway under light load), and any refusal or guard that the happy
path never exercises. For that whole class a test can assert the property, pass
forever, and be checking nothing at all — so spend your injections there rather than
on whatever is easiest to break.

**An injection that breaks the build teaches nothing, because the test never ran.** A
compile error is not a red test; it is the absence of a test result, and it reads the
same in a terminal as a failure. If the injection will not build, it has not reached
the property — rewrite it until the suite runs and the *assertion* is what objects.
The same applies to an injection that rings on plumbing (a missing file, a nil
dereference) rather than on the assertion you meant: record that as what it is instead
of counting it as a proof.

- **The cheap cases have no excuse.** Pure, deterministic logic — state
  machines, parsers, ownership/routing decisions, config defaults — tests with a
  fake collaborator (a small interface double) and a zero-value struct: no DB,
  no server, no fixtures. Choose scenarios that avoid the expensive-to-reach
  paths and you still pin the decision logic.
- **Anchor the test like any artifact.** Add `test-*.md` to `design.md`
  Artifacts mapped to the test file (so its refs harvest and future anchors
  there are seen), then `minispec update check` it once it passes.
- **`O`-gap is the exception, not the escape hatch.** Logging "missing tests" as
  an Oversight gap is for behavior genuinely disproportionate to test now — needs
  live external infra, a full rebuild, a real GPU. When you take that exception,
  say *why* in the gap. Everything a fake-and-zero-value can reach is written,
  not deferred.

**Upon completion**, run `~/.claude/bin/minispec phase implementation` to verify traceability, then run the Simplification Phase.

5. Simplification Phase
Invoke the `code-simplifier` agent on the recently modified code. This refines code for clarity, consistency, and maintainability while preserving functionality.

`code-simplifier` is a standard Anthropic plugin. If it is not available, skip the simplification pass and tell the user they can install it with `claude plugin install code-simplifier`. The alarm pulls below still run.

**CRITICAL: Preserve all traceability comments.** The `// CRC:`, `// Seq:`, and requirement references (`R123`) in code comments are load-bearing — they connect code to the design artifacts that justify its existence. Removing or reformatting them breaks the traceability chain that `minispec validate` checks. Simplification means cleaner *logic*, not fewer comments.

**This phase is where the fire alarms get pulled, and that is why the pull waits for it.**
A refactor is *precisely* when a property moves: it is licensed to change structure while
preserving behaviour, and "preserving behaviour" is adjudicated by the very tests whose
adequacy the alarm was proving. A simplification pass will happily restructure the function
**and** tighten its test in one go, which is exactly the pair that voids a proof — and
nothing about it turns anything red. The suite stays green, the injection becomes a memory,
and an alarm proved wired before the pass is now wired to a different building. A proof taken
*after* the pass is the only one that describes the code that ships.

So when the pass returns, work the whole injection list the Implementation Phase wrote down:
for each alarm, re-introduce the defect at its `**Inject:**` site, confirm red, restore, diff
to prove the restore was clean, and write the `**Pulled:**` line. **Writing that date without
running the injection is the one thing that must never happen** — it converts a record into a
claim, which is the whole failure the field exists to prevent. The phase is not finished until
every alarm it covers has rung.

*Alarms that were already recorded before this pass need the same treatment, and the tool says
which:* run `~/.claude/bin/minispec query alarms --unverified` and pull every one it names.
Measured in this project 2026-08-17, a single simplification pass restructured five functions
and voided six alarms; all six still rang, and nothing in the green suite would have said so
had they not.

**Re-pulling the recorded alarms is necessary and not sufficient — inject *past* the list as
well as through it.** The list says which properties someone thought to guard; a pass that
restructures a function can leave a property with **no alarm at all**, and re-running every
alarm the pass disturbed cannot find that. Measured 2026-08-18: after a pass, disabling an
entire branch left the whole suite green, because the property it carried — a blank line
between two blocks — **merges no lines and drops none**, so every check built on *what
survived* was structurally blind to it. So after re-pulling, break one more thing the pass
touched that no alarm names, and see whether anything objects. Silence there is a missing
alarm, not a passing one.

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
- **Oversights (On):** Missing tests *that were genuinely disproportionate to write in the Implementation phase* (say why — the cheap deterministic cases get a test, not an O-gap), tech debt, enhancements, security concerns, etc.
- **Approved (An):** Approved gap. Permanent — written without a checkbox. Good for "don't do it this way" requirements.
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

### Conformance deviations are gaps, and they link both ways

When a requirement states a rule the code does not yet honor everywhere,
**each deviation is its own gap** — not a paragraph of spec prose. Prose
cannot be queried, is never checked off, and drifts out of date silently;
a gap is greppable, carries a checkbox, and gets closed. So a spec states
the rule and says "deviations are tracked as gaps"; the gap list holds the
inventory.

**One gap per deviation, not one gap listing several.** Splitting them is
what makes each independently closable, and it surfaces ordering
constraints a combined body hides — dependencies between deviations only
become visible once they are separate entries that can block one another.

**Link both ways.** A gap whose repair will require editing a requirement
names that `Rn` in its body *and* says the requirement edit is part of the
repair. The requirement then carries a short back-link — a consistent,
greppable phrase such as `see gap <ID>` — noting that it is provisional
and what changes when the gap closes.

The back-link is the load-bearing half, and the one people skip. The
forward link (gap → requirement) is discovered by whoever works the gap,
who is already looking. The reverse is for everyone else: without it a
requirement reads as settled current intent, and a future agent
"fixes" code to match a clause that was already slated for removal —
precisely the revert trap "Supersede at the source" exists to prevent.
Write the requirement text so it still describes today truthfully, with
the pending change marked; do not pre-apply an edit that has not landed,
which would make the requirement a lie in the other direction.

Two forms not to confuse with this: a **retired** requirement (`Tn`)
already back-links by construction, since `minispec update retire` writes
the `~~Rn:~~ (Retired Tn — see Rxxx)` marker; and a gap that merely *cites*
a requirement as context needs no back-link, because nothing about that
requirement changes when the gap is repaired. Back-link only where the
repair edits the requirement.

To audit: for every open S/R/D/C/I/O gap naming an `Rn`, ask whether
repairing it changes that requirement's text. If yes, the requirement must
carry the back-link.

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

   To stderr it also prints a **supersede-at-source reminder** naming
   `R<old>`'s originating spec (its feature's `**Source:**`). Treat
   that reminder as a checklist item, not noise — it points at step 6,
   which applies to *every* retirement, not just migrations.

   If a CRC card or inline code comment still references the
   retired Rn but the code no longer fulfills it, update the
   reference to the replacement Rn. (References to retired Rn in
   code that was removed are fine — the comment went with the
   code.)

6. **Reconcile obsoleted spec *and* design prose — at the source.**
   Retiring a requirement has a forcing function — the `retire`
   command, the Tn entry, the `~~Rn:~~` marker, and the stderr
   reminder. The *prose* that described the old behavior has none.
   Two layers rot silently:
   - **Originating spec prose.** The requirement was born from a
     sentence in its feature's `**Source:**` spec — "current truth,
     the human's intent," the most authoritative trap of all. Follow
     the `**Source:**` the reminder names and rewrite or delete the
     sentence that spawned the retired `Rn`.
   - **Design prose.** CRC `## Does` descriptions, method signatures,
     and sequence diagrams that described state A do not flag
     themselves as stale. Grep `design/` for the changed method
     names, old signatures, and renamed types, and rewrite every CRC
     bullet and seq diagram that still describes the old behavior to
     match state B.

   This is **not migration-only.** Every retirement — standalone or
   part of a migration — owes this reconciliation, and the `retire`
   reminder prompts it each time. `minispec validate` cannot catch
   it: it checks that requirements are *referenced*, not that the
   prose around the reference is accurate. (Step 5 reconciles the
   *Rn references*; this step reconciles the *descriptions* those
   references annotate — a distinct, easily-missed pass.)

7. **Move the migration spec(s).**

   **Precondition — the prose grep.** Before completing, grep the
   retired module/type/old-behavior names across **both** `specs/`
   and `design/`. Every hit must be either gone or framed as a
   historical record (a retirement note, a `complete/` migration
   spec) — **zero stale-as-live mentions.** A grep alone can't tell a
   trap from an accurate "documents the absence" record, so this is
   your judgment, not the tool's. Apply the completion test: could an
   agent reading only specs + design be led to undo the migration? If
   yes, a trap remains — fix it before moving the spec.

   Then run:

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

## Trajectory Tracking (PENDING / CURRENT / DONE)

Specs → design → code anchor the project's **structure** — what exists and
why. They do not track its **trajectory**: what's queued, what's in flight,
what just landed. The harness task tool (`TaskCreate`/`TaskUpdate`) is
session-local and dies with the session. Trajectory tracking is the durable,
cross-session spine the structural docs and the ephemeral tasks both lack.

It is **tool-agnostic**: it tracks any kind of work — a mini-spec pass, a UI
pass, a plain investigation — each item naming the skill that runs it, or
none. It ships with mini-spec but is not about mini-spec.

**Every shape is normative in `trajectory-format.md`** (in this skill directory),
loaded on demand the way `config-reference.md` is: siting, the three file
shapes, the item and done entries, the carve status block, part and subpart
numbering, the marker vocabulary, and the ID rule. **Read it before writing or
repairing any of these files.** This section keeps only what a format cannot
carry — why the layer exists, and the judgment it asks of you.

### The three files

Named for the states an item passes through: **pending → current → done**
(future → present → past). They are called **the pending file**, **the current
file** and **the done file** throughout, so a path never needs qualifying;
`trajectory-format.md` says where they live.

- **the pending file** — the work queue, ordered by intent, and the index back
  to the roadmap, planning scratch, and feature designs.
- **the current file** — working context for the active item, and nothing else.
- **the done file** — the completion ledger.

### Working the queue

- **The top item is active.** Ordering is by intent, which is a judgment, not a
  score: position carries the priority and the number is only an identifier.
- **Finishing an item is `minispec pending finish <N>`**, and it needs no commit to exist:
  the item number is the identifier (Bill, 2026-09-15), so `finish` runs *before* the commit
  and the carve flip lands in the commit that lands the work. It writes all four surfaces in
  one act: the source file first (the carve where the work lives), then the
  current file's `## Active` section, then the move from the pending file to the done file.
  Source first, because that is the copy a future reader trusts, and the one nobody thinks
  to check.

  *The order is stated because it is the reason, not because it is a checklist* — the verb
  performs it. Give it `--body-file` and the done entry's body is placed in the same write, so
  there is no anchor to get wrong and no hand edit into a file git shows no diff for. The
  identifier slot takes `--discharged`, which the tool joins to the `#N` it owns.

  **A gap-sourced item must also say what happened to its gap: `--resolve` or `--no-resolve`.**
  Neither is a refusal, and there is no default, because a default would guess which of the two
  happened. `--no-resolve` is a record rather than a shrug: a gap left open by decision and one
  left open by oversight are identical in `design.md`, and the completion is the only place that
  difference is known.

  **Opening an item is `minispec pending start <N> --context-file <f>`**, and queueing one is
  `minispec pending add-item --from <doc>#<part>|<gap-id> "<title>" --status <text>` — the tool
  mints the number and writes both sides of the item↔part link. The only hand edit left in an
  item's round trip is the carve flip that names the item's commit, which rides in the next
  commit's tail.
- **One commit per batch, and the commit names every item it lands.** Finish the items,
  then `minispec pending commit-message --out <file>` composes the message — subject
  `#N, #M: titles`, body `Items #N, #M.` and each entry's done-file body — for
  `git commit -F <file>` with your sign-off appended. The item number is the identifier
  and `git log --grep '#N'` is the path from a part to its change, so a commit that forgets
  its items breaks the pointer silently; that is why the message is the tool's to compose
  and not yours to remember. Items may share a commit freely (Bill, 2026-09-15: three items
  had taken six commits, one per item and one tail each, the day this was decided).
  **The post-commit census re-pulls go back into that commit by amend, while it is
  unpushed**: `pending commit-message --amend` returns `HEAD`'s message unchanged with the
  new items after it — the previous message is part of the record, so an amend appends and
  never rewrites — and refuses when `HEAD` is on a remote, where a follow-up commit is the
  answer. A checkpoint commit an item needed while it was worked — a delegated pull checks
  out a commit — is folded into the batch with `squash`, never `fixup`, before the batch
  commit; never `#`-led headers, which git strips.
- **The current file is a resume buffer.** To pause an item, lift its context
  into a sub-item under that item's `##` heading in the pending file, then reset
  the current file — freeing it for whatever you pick up next.
- **Never let the current file become a log.** Finished work goes to the done
  file. This is the rule most often broken, because leaving the last item's
  context in place costs nothing at the moment you do it.
- **Standing context is not a log, and clearing it is the opposite mistake.** A log
  records *what happened*, which the done file owns; standing context records *what
  is still true* — answers already obtained from the user, pointers to work outside
  this repository, the state of things — and nothing else holds it, since the pending
  file is per-item, the done file is history, and a carve is per-problem. It lives in
  its own `##` sections beside the active item's, which is why the active item has a
  heading of its own for a tool to address. `trajectory-format.md` has the shape.

### When a queue operation goes wrong: `pending revert`

**There is one level of undo and one of redo over the trajectory files, and you should
know it exists before you need it.** A safety mechanism nobody knows about is not a
safety mechanism — which is why this sits here rather than only in the tool's help.

```
minispec pending revert     # undo the most recent trajectory change
minispec pending replay     # redo what revert undid
```

**It is a slot, not a stack.** The most recent change is revertable and replayable;
nothing older is recoverable. That is enough for the real emergency — a command that did
the wrong thing thirty seconds ago — and it deliberately avoids owning a history git
already owns better. **Any new queue operation discards what was revertable**, without
ceremony.

Four things worth knowing before you reach for it:

- **It refuses rather than clobbering.** If you hand-edited a trajectory file since the
  change, revert stops and names which file and where its backup is. You are better
  placed than the tool to reconcile a hand edit with a pending undo, so it does not
  guess. The backups are in `.minispec/backup/`.
- **Exactly one of revert / replay is legal at any moment**, and a refusal tells you
  which. There is no "revert twice."
- **It covers the trajectory files only** — not `design/`, not your source. The `update`
  verbs have no undo, and neither does anything else. Reverting a queue operation does
  not touch the code you wrote under it; the done entry that names its commit is how you
  find that work.
- **A carve is not restored, on purpose.** Revert marks the part `**REVERTED (#N.)**`
  instead. That is what keeps the part's vended number visible rather than silently
  un-vending it — the queue rolls backward while the carve moves forward.

**When an attempt is abandoned rather than replayed**, the part returns to
`**OPEN (not queued.)**` and its number goes back into the pool. Aborting an attempt is
not aborting the part: it is still open and still to be done. Nothing is written to the
done file, which records completions and would be diluted by non-events. If a released
number is handed out again, the tool says so.

**And the worktree anchor.** Every transition first records the whole working tree — untracked
files included, ignored paths excluded — at `refs/minispec/snapshot`, outside the stash so
nothing can pop or clear it. It is reference, never undo: `git show refs/minispec/snapshot:<path>`
gives a file back as it stood before the transition, and `refs/minispec/snapshot^1` is the
commit that was checked out. The tool never restores from it; you do, by hand.

### Interleaving with migrations

A **state item** and a **migration** are two orthogonal lifecycles,
composable as the work demands:

- A **migration** distills a brainstorm into a concise A→B document that may
  span several steps; it runs the phases and lands in `complete/NNN-`. It
  can be done all-at-once and may never enter the pending queue.
- A **state item** is a unit of queued work, paused and resumed via the
  current file.

They compose; they do not nest by rule. To change styles mid-flight, park the
active item the usual way — the pending file is a stack you can push onto — and
the current file is free for the migration. The freedom to intermix is the point;
neither style is imposed.

### Carves — the layer above the item

An item is a unit of *work*. A **carve** is the layer above it: one coherent
problem decomposed into items that may each need a different skill. It is the
document those items point back at.

It exists because coherence and schedulability have different natural units. A
problem is coherent at the size of "the review console" — change one decision
and the others move. Work is schedulable only in pieces that fit one session
with one skill loaded. So no session can hold the whole problem and the queue
can only hold pieces. Something has to carry the whole, and that is the carve.

Both a carve and a migration are documents that spawn work, but their
properties are close to inverted:

| | Migration | Carve |
|---|---|---|
| End state | Defined (A→B); expires by design | None; decays as parts land |
| Completion | Ritual: prose grep, `migration-complete`, `complete/NNN-` | `update finished-carve`: a move to `done/` with its links rewritten |
| Spawns | One coherent change, phases run once | N items, scheduled independently over months |
| Content | How to get from A to B | Decisions, open forks, and the split |
| Kinds of work | One | Deliberately several |

A migration says *the system is at A and must reach B*. A carve says *here is a
problem area, here is what we have settled, here is how it breaks into
schedulable pieces*. The lifetime difference is the sharpest: migrations are
temporary by design, while a carve can stay open for months.

**Why a carve lives with the public design docs and not with the private
trajectory files:** it is almost entirely facts about the code, and those belong
where someone reading the project can find them. Deliberately not under `specs/`,
which describes how the system *is*; a carve is a work-management artifact, the
same reason migrations were exiled to `specs/migrations/`. The paths themselves
are mandated in `trajectory-format.md`.

**Promotion is a judgment call, and it belongs to the maintainer.** The test is
*is this a meaty task?* — substantial enough to stay open a while and worth
naming, because the name is what makes the surrounding work manageable. No count
of queue items decides it: a one-item document can be carved on the expectation
of more, and a two-item one can stay a working note if nobody needs the name.
Keep the number of live carves small; they are a management tool, and a directory
full of them stops being one.

**An agent proposes; it does not decide.** Raise it once a document has spawned a
second item — that is the prompt to ask, not a rule that fires. Below that,
usually stay quiet. Treating the count as the criterion gets it wrong in both
directions at once: it refuses a meaty single-item problem while mechanically
promoting anything that happens to spawn two.

Promote forward, moving an existing carve when it is next touched rather than in
a sweep, and rewrite the links that pointed at the old path as part of the move.
Promotion also forces an editorial pass separating the decision from the private
reasoning behind it, and writing for a stranger is the cheapest clarity check
available.

**When a carve is finished, the move is `minispec update finished-carve <carve>`**: it
refuses a carve with an open part, moves it to `done/`, and rewrites every link in both
directions from where each resolves — a plain rename, nothing staged. A move made by hand
leaves every relative link pointing where the carve used to be; `update repair-links`
repairs that after the fact, and `query links` reports what it could not.

**Read the whole carve before working one of its parts.** A carve keeps a part's information
in several places and only some of them are keyed by the part number — the status line and the
elaboration, both reachable by grepping `Item N`. A scoping decision in `## Decisions`, or a
design paragraph elsewhere, is keyed by nothing, and grep finds only the phrasings you can
guess. Measured on three occasions in one project: a decision that superseded its own part's
elaboration **the day both were written** was missed for hours and the work was scoped wrong;
an item accumulated five separate pieces before anyone totalled them out loud; and a design
fork was put to the maintainer whose both halves were already answered in carve text nobody
had read. Budget the read as part of the item — it is not preparation for the work, it is the
first step of it.

**A carve answers what is *scheduled*, never what is *possible*, and reading it for the second
question is how this project lost an item.** A part marked `OPEN` means the work is unscheduled;
it does not mean the capability is absent. Measured 2026-08-26: a session read `OPEN` as *the
capability does not exist* and built a fold marker — spec, requirement, design, code, four alarms,
two gaps — for something the binary had carried for five days. **Ask the tool what exists**
(`minispec --help`, `minispec query carves`; `query sdom` is on the reclaim list) and the document what is planned; a document cannot answer the first
question and will look as though it did.

*Read the sibling carves too where a part names one, and the reason is the same one that makes
a carve worth having:* parts move between carves, and a decision often lands in the document
that lost the part rather than the one that gained it. Where a carve keeps each decision with
the part it governs, a partial read becomes safe; until then this is the mitigation, and it is
cheaper than the alternative by a wide margin — a carve is one document, and re-deriving a
decision you already made is not.

**Three disciplines.** The first two tend to happen by instinct. The third does
not, and it is the one that matters. `trajectory-format.md` has the shapes; what
follows is why each one is worth the trouble.

1. **Per-part status, in a block at the top.** A carve outlives the length
   anyone reads end to end, so "what is still open?" has to be answerable from
   the first screen. Left to grow where the work happened, status lands
   two-thirds down and is effectively invisible.

   The rule underneath the shape is *nothing stated twice*: the block owns the
   title and the status, the body owns the detail. A status table restating body
   prose is worse than none, because the two will disagree and nothing will say
   which is right.

   **`NOT VERIFIED` is the marker to reach for deliberately**, and the one no
   tool can ever check. A repaired *symptom* reads exactly like a satisfied
   *requirement*; only a person who read the code can tell them apart. That
   misreading is the most expensive one this block prevents, so state the
   negative rather than leaving a part unmarked — unmarked is indistinguishable
   from unconsidered.
2. **Dated, attributed decisions.** `DECIDED (name, date)`, append-only, so a
   reader can tell a settled call from a musing and whose it was. This is the
   single highest-value habit in the format.

   Its companion: **supersede in place.** When a decision overturns an earlier
   one, say so *at the earlier one*. A dated `DECIDED` sitting beside an unmarked
   paragraph that contradicts it will be read as current, because nothing about
   the unmarked paragraph looks provisional.
3. **Migrate on landing.** When a part lands, its decisions belong in `specs/`
   and `design/` as requirements, spec prose, or a comment at the code; the
   carve then points at where they went. A migration gets a forcing function for
   free (the retire reminder, the prose grep, the completion ritual). A carve
   gets none, so its decisions rot silently while still reading as current. This
   is not hypothetical: a stale line in one working note sent a later carve down
   a wrong path, and that carve had to mark the note stale by hand.

Adopting discipline 3 also reframes the planning scratch usefully, as a staging
area for reasoning that has not earned a public home yet rather than a permanent
one.

### What a project chooses, and what it does not

**Almost nothing is a project setting.** Siting, filenames, the carve directory,
privacy, and how parts are keyed are all mandated or per-document properties —
see `trajectory-format.md`. Two projects running this layer should produce files
a stranger can read interchangeably, which is the whole reason the shapes are
written down rather than described.

What is genuinely yours: the **routing labels** (which skill runs an item, or
none), where the **planning scratch** lives, and any **batching rules** about how
much work an item should hold.

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
**Code:** store_test.go
**Fire alarm:** what to break so this test goes red, and what red looks like
**Inject:** links.go:SyncLinkPath
**Pulled:** 2026-08-05 — rang
```
Cover: happy path, errors, edge cases.

### The fire-alarm fields

A test that has never failed is an assertion that happened to be true when you
wrote it. Before trusting a new guard, break the thing it guards and watch it
scream — then **write down what you broke**, because the proof expires the moment
its subject is rewritten and nothing in a green suite will say so.

- **`**Fire alarm:**`** — the injection in prose, and what the failure should look
  like. Prefer the *real* historical defect over a strawman: restoring the actual
  line proves the test catches *that*, while shuffling something by hand proves only
  that the test dislikes shuffling.
- **`**Inject:**`** — `file:symbol` for the site the injection **edits**, comma-separated
  when an alarm has more than one. This is the field a tool can use and the one that is
  easy to get wrong: name what the injection *changes*, never a symbol it merely
  *consults*. "Return early when `Lookup` already knows the path" edits the caller —
  `Lookup` is context, and recording it there points every future check at the wrong
  function.
- **`**Pulled:**`** — the date it was actually run and what happened. **The census reads the
  *leading* date**, so a re-pull moves that date; appending *"re-pulled today"* further along
  the line records the history for a human and leaves the alarm stale for the tool. Measured
  2026-08-18: three alarms re-pulled, all three rang, and the census stayed red until the
  leading dates moved. **Absence is
  meaningful and must not be filled in by guessing:** an alarm with no `Pulled` is a
  *prescription* (here is the injection to run) rather than a *record* (I ran it, it
  rang). Those read identically in prose and are entirely different claims — the same
  reason `NOT VERIFIED` earns its own words in a carve's status block. **Write the date
  and the failure's signature; never a commit hash.** A batch commit is amended for the
  census and a checkpoint is squashed into it, so a hash written during the work names a
  commit that is rewritten away — a record pointing at nothing, in the field that exists
  to be a record. The same reason took the hash out of the carve's `LANDED` line on
  2026-09-15; the line carries what cannot go stale (Bill, 2026-09-04).
- **`**Code:**`** — the test file, so the alarm and the test it vouches for are
  linked in the direction a tool can follow.

**Why `Inject:` is a field rather than something to read out of the prose.** The
forward chain already runs test→code; the person rewriting a function is in the code
and never opens the test doc. `Inject:` is the back-link that makes "what alarms cover
what I am about to change" a grep instead of a memory — and the back-link is always
the half that gets skipped, because the forward one is found by whoever is already
looking.

### Delegating the re-pull, and the one thing a delegate must never send back

**Re-pulling is where alarm work actually costs.** Measured 2026-08-20 by `/context` at the
end of a long session: tool results **190.9k tokens, 19% of the window** — the largest single
category and roughly four times what the exchanges themselves cost. Almost none of it was the
census. It was the **read-edit-test-restore-diff cycle**: nine pulls and re-pulls, two
past-the-list probes and one cross-project reproduction, inside a single item. That cycle is
mechanical, it is long, and every byte of it lands in the context of whoever is *also* holding
the design decision the item is about.

So hand the cycle out, one alarm to one agent:

```bash
~/.claude/bin/minispec query alarms --unverified --brief
```

Each brief is a complete spawn prompt — sites, test files, the `**Fire alarm:**` prose
verbatim, and the contract. Spawn one `alarm-puller` per brief, **with worktree isolation**,
since the job is to corrupt source on purpose and a puller working in the live tree is a
puller that can lose your uncommitted work.

**Commit before you fan out, and this is a harder precondition than the census's.** A
worktree is a checkout of a **commit**; uncommitted edits and untracked files are not in it.
So a puller sent at an alarm whose subject you wrote this session lands in a tree where the
function does not exist — and what it reports is a build failure or a missing symbol, which
costs a spawn and a round trip to learn something `git status` would have said for free. The
census's blindness is at least *legible* (it says `unchecked`); this one is silent, because
the puller's report is perfectly well-formed and about a different tree than you think.

Measured 2026-08-20 while building this, and measured by **asking git rather than reasoning
about it** — the rule one section down applies to worktrees too. `git worktree add --detach
<dir> HEAD` over a tree holding five uncommitted injection sites produced a checkout with the
old `alarm.go` present, the new `brief.go` absent, and zero occurrences of the function three
of the alarms name. The probe cost one command; five confused reports would have cost five
spawns and the round trip to work out why they disagreed with each other.

**Make the worktree yourself, at HEAD, and check where the delegates actually are.** A
worktree is a checkout of *some* commit, and which one is the harness's decision rather than
yours. Measured 2026-08-20 on this layer's first real run: nine pullers spawned from a tree at
`32d194a` all sat at `cac3baf` — `origin/main`, 103 commits behind — because the harness cut
from the remote tracking branch, and every site the briefs named had been written in those
commits. `git worktree add --detach <dir> HEAD` by hand, the path in the prompt, no isolation
flag; then `git worktree list` after spawning says where they are. Pushing is not the answer:
a worktree shares the object database, so the commit is already reachable with no network.

**And the sibling the module needs.** `tool/go.mod` replaces `simple-dom` with a *relative*
path, resolved from the worktree. A worktree outside `~/work` fails its baseline with
*replacement directory does not exist* until a `mini-spec-tool` sibling sits beside it —
measured 2026-09-06, six pullers stopped at the gate; one cloned the sibling for itself, which
worked and was a copy at a head nobody had checked. Link it before spawning.

**What the run proves about the protocol, which is the other half.** Four pullers reached the
stop condition and four stopped: no injection, empty `git diff`, clean tree, and one worked out
`cac3baf` and named it. A puller that had guessed instead would have returned a well-formed
five-part report about a codebase that no longer exists — and *that* report is
indistinguishable from a good one. The stop condition is not defensive politeness; it is the
thing standing between a delegated loop and confident fiction.

**One environment hazard, measured the same run.** A Go workspace file *above* the repository
(`/home/deck/work/go.work` listing `./mini-spec/tool`) routes `go test` inside a worktree at the
**main checkout**, not the worktree. Three pullers hit it and worked around it with `GOWORK=off`;
one reported a green baseline for a file its own worktree does not contain. Go refused loudly
that day; **measured again 2026-09-27, it did not** — `go list -m` in a worktree under the
workspace named the main checkout's module and nothing objected. A delegate whose commands can
reach the live checkout is a delegate whose isolation is nominal.

**The fix is a workspace of the worktree's own, not `GOWORK=off`.** A pull proves the alarm rings
on the code *as it is built*, so it runs in the development environment, local dependencies
included, and `GOWORK=off` swaps those for published versions, testing a program nobody ships.
Write a `go.work` at the worktree root naming the worktree's module and each local dependency;
Go takes the nearest workspace file, so it wins over the one above. Write it by hand: `go work
init` inside a directory an outer workspace already governs refuses, since the outer file
already exists. Then check `go list -m` names the worktree's module before trusting a baseline.

**A second environment hazard, measured 2026-08-23, and this one fails silently.** Go's test
cache keys on source and build inputs and does **not** track files a test reads at run time. So
an alarm whose injection edits a *corpus file* rather than a Go symbol produces `ok … (cached)`
on the injected tree — the pre-injection result, reused. Measured on `test-Carve.md#12`, which
injects into `carves/reference-discipline.md` while its test reads `carves/*.md` through
`os.ReadFile`: the delegate applied the injection, ran the suite, and got a green that meant
*the test did not run*. **Run every pull with `-count=1`**, in the command the brief carries
rather than trusting anyone to remember it. The delegate found this by reading the word
*cached*, not because anything objected — then re-established its baseline and re-ran the five
alarms it had already finished, which is the right instinct: a batch whose command was wrong is
a batch of unproven records. Banked as `O113` on `old-sdom`; the brief here carries no command, so the rule is yours.

**Sweep the worktrees when the batch is done, and count it as part of the run.** A puller edits
files on purpose, so its worktree is never *unchanged* and never auto-cleans; they accumulate at
roughly 13M each and nothing reports them. Measured 2026-08-23: fifteen worktrees holding
**185M**, and **31** orphan `worktree-agent-*` branches — more branches than trees, because
earlier sweeps removed the trees and left the refs behind. The sequence, and the two checks that
make it safe to run without reading every tree by hand:

    git worktree list --porcelain                 # every tree clean before you start
    git worktree remove <path>                    # no --force: refuses if one is dirty
    git worktree prune
    git merge-base --is-ancestor <branch> main    # nothing unique in the branch
    git branch -d worktree-agent-...              # -d, not -D: refuses if that was wrong

**The refusals are the design.** `remove` without `--force` and `branch -d` rather than `-D`
mean the tool stops you when your inventory was wrong, instead of the discipline having to be
right every time. That run reclaimed 184M and refused nothing. *And the tool now says what moved:*
`minispec pending changes` reports every path changed, added or deleted since the last queue
transition, with the two commands that restore each — landed 2026-09-14 as
[trajectory-tool.md](../../../carves/done/trajectory-tool.md) Item 14 — so the inventory above
starts from a report rather than from memory.

**The contract is the whole design, and it is one sentence: evidence, never a verdict.** What
comes back is the command, its output before the injection, the diff applied, the output
after, and the diff after restoring. *You* decide whether the alarm rang, and *you* write the
`**Pulled:**` line. Not because a delegate is untrustworthy in general, but because this
particular judgment is exactly the one that fails from inside the loop. Measured, three times
in one session: an injection at a correctly named site that **did not ring** because the rule
had two guards; an injection that rang across three packages while being **incapable** of
reaching the property it named; and an `**Inject:**` field naming a symbol the injection only
*consulted*. Every one of those reads as a clean pull from inside and as a defect from
outside.

**A `**Pulled:**` date written by the agent that ran the injection is not a record.** It is
the agent's own verdict on its own work, wearing the format of evidence — the same conversion
of a record into a claim that bumping a date without re-running the injection performs, and
the reason that move is forbidden two sections up. The date is written by the reader, from the
evidence, or it is not written.

*What delegation is not for:* the census itself. A subagent that runs `query alarms` and
reports the interesting lines is strictly worse than `--unverified` on every axis — it costs a
spawn per run, forever, to perform an omission a filter performs for nothing, and unlike a
filter it can be **wrong**. Delegate the loop; filter the list.

### The census is blind to whatever git cannot see — scaffolding, and dated as such

`minispec query alarms` answers freshness by asking git when each `**Inject:**` symbol last
changed. **So it can only speak for code git already holds.** On an untracked file, or a
symbol written since the last commit, there is no history to search and the honest answer is
`unchecked` — which is not `verified`, but a census that reports both in one line reads as
fine at a glance.

**Take the census after committing, and budget the re-pull.** Measured in this project
2026-08-17: the pre-commit reading was *78 alarms, 0 stale* with `validate` green; the commit
made one new package visible to git and **nine alarms went stale at once**, turning `validate`
red. Nothing about the code changed between those two readings — only whether the checker
could look. An earlier instance the same week hid seven unresolvable anchors the same way.

Re-pulling nine took about twenty minutes, because every injection was written down. That is
the argument for `**Inject:**` and `**Fire alarm:**` being *fields* rather than recollection,
arriving from an unexpected direction: they are what makes an expected, schedulable cost out
of one that would otherwise be a re-derivation nobody performs.

**This section is temporary and should be deleted rather than maintained.** `unchecked`
currently conflates three unrelated reasons a question could not be asked — no git at all, a
subject with no history *yet*, and a subject that is a document with no resolvable symbol —
and only the middle one is closable, by committing. When the tool separates them and names
the closable one, it can say this itself, on the runs where it is true and silently on the
rest. Delete this then; a notice that fired on all three would fire forever and be muted
along with the ones that matter.

## Delegating a measurement — and the half that is not delegable

**Exploratory measurement is the other half of the context bill**, and it surfaced only once the
re-pull cycle was handled. Measured by `/context` on 2026-08-21: tool results **171k tokens, 17%
of the window**, against 19% the day before — nearly the same share, and almost none of it
re-pulls this time. It was corpus censuses, throwaway probes, regex classification, and
before-and-after diffs across two repositories. *Every finding landed in a gap, a carve or a
commit message; the raw output stayed resident and was re-derivable from none of it.*

**The line is not "delegate measurement."** *Delegating the re-pull* already says why a census
must not be delegated — a subagent that runs a query and reports the interesting lines is
strictly worse than a filter on every axis. The rule here is narrower, and it is a boundary
rather than a permission:

> **A measurement whose question is already precise is delegable. The search for what to measure
> is not.**

*Sorted from real probes, which is what makes the boundary measured rather than stipulated.*

| delegable | not delegable |
|---|---|
| how many part lines carry text before their marker (0 of 68 in one project, 11 of 34 in a second) | *why is the marker not being read?* |
| anchored versus unanchored regexes per layer (50 and 8) | *is this leniency safe?* |
| which continuation indents are in live use (0, 2, 4 and 6) | *which of these three repairs is right?* |
| whether a fit heuristic swallows a given marker (no — 2 of 5 lines taken) | *what should we measure next?* |

Everything on the left is a throwaway script, a handful of numbers, and two or three failed
compiles nobody needs to see. The entry on the right that proves the boundary is a truncation
bug: finding it took a hypothesis chain — the part-line reader? no; the part reader? no; the
content span? yes — and the diagnosis was **revised twice inside a banked gap before it was
right**. A delegate handed *find why the marker is not read* returns an answer nobody can cheaply
check, and that is the shape that produced the one hard failure in this project's adversarial
delegation rig.

**Spawn `measurer`, one per question.** `.claude/agents/measurer.md` carries the protocol; the
brief carries only the question, the population, and the paths.

**It runs without worktree isolation, and that is deliberate rather than an omission.** A
measurement's population routinely spans two repositories, only one of which a worktree could be
cut from — and isolation buys nothing here because the job is to read rather than to corrupt. It
would also inherit both hazards *Delegating the re-pull* records: a worktree cut from the remote
tracking branch, and a build-system workspace file routing commands back at the live checkout.
What replaces it is narrower and checkable: **the probe is written outside every repository it
measures**, and the agent reports its `git status` at the end.

**The brief must name the population, and this is the requirement the whole thing turns on.** Not
*the carve files* but *which files, found how*. A brief naming only a question invites the
delegate to choose the scope, and choosing the scope is choosing the question.

**Define it with a command that already knows the document model, never with a text match
over the files.** Measured on this section's own first run: a population given as
`grep -lx '## Status'` returned eleven checkbox-less lines in the second project, and
**seven were inside a fenced *sample* of a status block, in a document about status
blocks**. The project's own reader is fence-blind by construction and excludes that file
correctly; a line-oriented predicate is not, and it then ran to end of file because the
document had no later heading to stop at. `minispec query carves` lists the real set.
*The delegate counted exactly what it was given and was right to* — which is the design
working rather than failing: a scope error surfaces as a wrong number you can see, instead
of as a quiet correction you cannot.

*The motivating error is worth carrying, because it shows what a prompt fixes and a habit does
not.* The day this was proposed also produced its own counter-example: eight marker-shaped runs
inspected, all eight genuine, **zero false positives** banked into a gap — overturned by counting
all fifteen and finding that five were body prose. Nothing about the eight was wrong; the sample
was not the population, and the report did not say so. **That error does not survive being
written as a prompt**, which is the argument for the agent over the discipline: a delegate that
must be told what to count cannot eyeball, while a person who already knows the rule can.

**What comes back is counts and instances, never scrollback** — the inverse of `alarm-puller`,
whose evidence *is* the test output. Returning the raw run would reimport precisely what the
delegation was for. The report carries the command instead, so the run is reproducible without
being repeated into your context.

**And it reports what it could not classify.** `LEFTOVER` is a field rather than a footnote,
because a probe that silently drops what it could not reach reports clean over the part it never
saw — which is the defect most of these measurements are taken to find, arriving inside the
instrument.

**The contract is `alarm-puller`'s and is not re-derived: evidence, never a verdict.** What comes
back is the population, the command, the counts and the leftovers — never *so the rule is safe*.
You decide what the numbers mean, because that judgment is the one that fails from inside the
loop.

*One thing stated rather than implied:* the `model:` in the agent definition is **inherited, not
measured**. This project's nine-puller rig measured `sonnet` for `alarm-puller` briefs, and that
result does not transfer to a different job. The rig is cheap to re-run, and the more useful
finding from it applies here directly — what decided every case was **which check the delegate
happened to run**, not how strong it was, which is why the enumeration rule is written into the
agent rather than left to the tier.

## Quality Checklist
- [ ] Requirements: all spec items captured, numbered (R1, R2, ...), inferred items marked
- [ ] CRC Cards: nouns/verbs covered, no god classes, Requirements linked
- [ ] Sequences: participants from CRCs, ≤150 chars wide
- [ ] UI Specs: ASCII layouts, refs to CRCs and manifest-ui.md
- [ ] Traceability: design files in Artifacts, code files have checkboxes, all Rn referenced
- [ ] Tests: test-*.md for key behaviors
- [ ] Fire alarms: every guard written *after* its bug was fixed has been broken on purpose and confirmed red, with `**Inject:**` naming the edit site and `**Pulled:**` the date — and any alarm whose subject was rewritten since has been pulled again. `~/.claude/bin/minispec query alarms --unverified` reports every alarm that carries a decision, with the count still covering all of them; take that reading **after committing**, since it is blind to code git cannot yet see. Re-pulls can be fanned out — see *Delegating the re-pull*
- [ ] Summary specs: any cross-cutting axis touched by this change has been mirrored in the relevant summary spec (CLI inventory, storage layout, API surface, capabilities, …) — see the project's pinned list
- [ ] Root spec index: every per-feature spec is mapped under a system in the root index (created if absent); `~/.claude/bin/minispec query unindexed-specs` returns empty
- [ ] Phase validation: `~/.claude/bin/minispec phase <phase>` passes after each phase
- [ ] Full validation: `~/.claude/bin/minispec validate` passes
- [ ] Trajectory validation: `~/.claude/bin/minispec validate trajectory` passes — repository-scoped, so it runs once per repository and never per design root; `make validate` runs both

## Minispec Tool

The `minispec` CLI tool (at `~/.claude/bin/minispec`) performs structural operations on design files.

**IMPORTANT:** Always use minispec commands instead of manual editing for:
- Checking/unchecking artifact checkboxes
- Adding requirement references to CRC cards
- Querying artifact states and coverage
- **Finding the next free `Rn`, gap or item number** — `query next-id`. Never compute it
  with a grep. The tool counts every file that can hold one and reports what it read; a
  grep counts what you thought to point it at, silently misses retired and ranged forms,
  and hands you a number that collides with an existing ID

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
~/.claude/bin/minispec query coverage R5-8   # Design files covering each Rn (ranges honored; no args = all)
~/.claude/bin/minispec query uncovered       # List Rn without design refs
~/.claude/bin/minispec query gaps            # List gap items
~/.claude/bin/minispec query requirements    # List all requirements
~/.claude/bin/minispec query migrations      # List in-flight migration specs
~/.claude/bin/minispec query alarms          # Every recorded fire alarm with its state
~/.claude/bin/minispec query alarms --unverified   # Only the ones that carry a decision; the count still covers all
~/.claude/bin/minispec query alarms --unverified --brief  # One spawn prompt per alarm, for a delegated re-pull
~/.claude/bin/minispec query carves          # Open, landed and stateless parts per carve, from the status block only
~/.claude/bin/minispec query carves --open   # …and the open parts and stateless lines themselves
~/.claude/bin/minispec query implementation R5-R8   # Where each Rn is implemented, file:line and comment
~/.claude/bin/minispec query implementation '@status'  # …or by a regexp over requirement text
~/.claude/bin/minispec query next-id req     # Next free Rn (counts retired ones too)
~/.claude/bin/minispec query next-id gap     # Next free number for every gap type
~/.claude/bin/minispec query next-id item    # Next free queue ID (pending + done files)

# Trajectory items — the three verbs that write the queue files, and the slot beneath them
~/.claude/bin/minispec pending add-item --from <doc>#<part> "<title>" --status <text>   # mint + both sides
~/.claude/bin/minispec pending add-item --from O136 "<title>" --status <text>          # a gap is a source too
~/.claude/bin/minispec pending start <N> --context-file <f>     # write the current file's Active section
~/.claude/bin/minispec pending finish <N> --body-file <f>      # all four surfaces, one act; before the commit
~/.claude/bin/minispec pending finish <N> --resolve            # ...and close the gap it named
~/.claude/bin/minispec pending finish <N> --no-resolve         # ...or record that it did not
~/.claude/bin/minispec pending commit-message --out <f>        # the batch's message, naming every item it lands
~/.claude/bin/minispec pending commit-message --amend --out <f> # HEAD's message plus the items since, appended
~/.claude/bin/minispec pending revert                          # undo the most recent trajectory change
~/.claude/bin/minispec pending replay                          # redo what revert undid

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
