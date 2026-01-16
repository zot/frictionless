---
name: mini-spec
description: **Use proactively**. Use a central index to explore the project. Update it properly when changing it. Covers specs, design, and code
---

# Mini-spec

3-level architecture: specs → design → code.

```
specs/    # Human specs (language, environment required)
design/   # SOURCE OF TRUTH: crc-*, seq-*, ui-*, test-*, manifest-ui.md
docs/     # user-manual.md, developer-guide.md
src/      # Code with traceability comments
```

## Task Tracking

Use todos (TodoWrite) to track progress through phases and tasks:
- Create a todo for each phase you'll work through
- Break implementation into per-file or per-component todos
- Mark phases complete as you finish them
- Use the Quality Checklist items as todos before finalizing

## Core Principles
- use SOLID principles, comprehensive unit tests
- when adding code, verify whether it needs to be factored
- Code and specs as MINIMAL as possible
- Before using a callback, see if a collaborator reference would be simpler
- write idiomatic code for the language you use
- avoid holding locks in sections that have significant functionality

## Cross-cutting Concerns

`design.md` Cross-cutting Concerns section: Patterns spanning components (auth, errors, logging, routing, theming).
Referenced from other design artifacts: Cards, sequences, and layouts can all say "see cross-cutting: auth"

## Traceability

`design.md` Artifacts section: design files with code file checkboxes.

**Code changes:** Uncheck `[x]`→`[ ]`, ask user: "Update design, specs, or defer?"
**Update design:** Read code, update design file, re-check box.

## Workflow

**Read specs first.** Specs must indicate language/environment.

1. Spec Phase
Create in `specs/`: human readable, natural language descriptions

**Upon completion**, offer to update the design (Design Phase). Do not jump to Implementation.

2. Design Phase
Create in `design/`:
- `design.md`: Intent + Artifacts (design files → code file checkboxes)
- `crc-*`: CRC cards (see format below)
- `seq-*`: sequence diagrams (≤150 chars wide)
- `ui-*`: ASCII layouts, reference CRC cards
- `test-*`: test designs (see format below)
- `manifest-ui.md`: routes, theme, global components

**Upon completion**, offer to update the implementation (Implementation Phase). Do not jump to Gaps.

3. Implementation Phase
Add traceability comments:
```
// CRC: crc-Store.md | Seq: seq-crud.md
add(data): Item {
```
Mark implemented: `[ ]`→`[x]` in Artifacts.

Look out for language-specific "gotchas" like mixing functions an methods in Lua.

**Upon completion always do the Gaps Phase.**

4. Gaps Phase

`design.md` Gaps section tracks (use S1/D1/C1/O1 numbering):
- **Spec→Design (Sn):** Spec features without design artifacts
- **Design→Code (Dn):** Designed features without code
- **Code→Design (Cn):** Code without design artifacts
- **Oversights (On):** Missing tests, tech debt, enhancements, security concerns, etc.

Nest related items with checkboxes:
```markdown
- [ ] O1: Test coverage gaps
  - [ ] Feature A (5 scenarios)
  - [ ] Feature B (3 scenarios)
```

**Upon completion**, offer to update Documentation (Documentation Phase).

5. Documentation Phase, Optional -- offer to user after Gaps
Create `docs/user-manual.md` and `docs/developer-guide.md` with traceability links.

## CRC Card Format
```markdown
# ClassName
**Source Spec:** feature.md
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
- [ ] CRC Cards: nouns/verbs covered, no god classes, Source Spec linked
- [ ] Sequences: participants from CRCs, ≤150 chars wide
- [ ] UI Specs: ASCII layouts, refs to CRCs and manifest-ui.md
- [ ] Traceability: design files in Artifacts, code files have checkboxes
- [ ] Tests: test-*.md for key behaviors
