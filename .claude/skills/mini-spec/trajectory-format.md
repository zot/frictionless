# Trajectory file formats — normative reference

The shapes of the trajectory layer: the three queue files, the carve, and the markers
that record state in both. `SKILL.md` says *why* the layer exists and what judgment it
asks of you; this file says what the files must look like. Load it on demand.

**It is normative.** Where this file and any other description of these shapes disagree,
this file wins — `SKILL.md` included. Anything that has to restate a shape carries a
section *generated* from this file rather than a hand-written copy beside it: two
hand-maintained normative copies of one format is the failure the whole layer exists to
prevent.

Three readers, and the last two are why a normative statement is needed rather than
prose:

1. **The tool implements it.** Several commands encode format knowledge and would
   otherwise each invent their own reading.
2. **An agent repairs from it.** When a file is damaged, predates the tool, or was
   hand-written by someone with only prose to copy from, the shapes are needed *precisely
   when the tool is what cannot help.*
3. **Old layers migrate to it — this is a target, not a survey.** A project already
   running a hand-maintained trajectory has files in some earlier shape, and the point of
   writing the shapes down is that an agent can bring them **up to code** for the tool.
   So where a variant is merely what somebody happened to write, this document names the
   one form and the others migrate; it does not bless every shape it found. Backward
   compatibility is not offered, because the migration is an edit and the alternative is
   carrying every historical variant forever.

**The format carries no version number.** Conformance is the signal: a file matching this
document is current, and one that does not is legacy and gets migrated. A tool-created
file is recognizable by the lifecycle preamble creation writes into it. A number in every
file, forever, buys nothing until there is a second version to name — and that is when it
earns its place.

---

## Siting

Mandated, not configured. There is no config key for any of it.

| Path                       | What                                                |
|----------------------------|-----------------------------------------------------|
| `<repo root>/PENDING.md`   | the pending file — the work queue                   |
| `<repo root>/CURRENT.md`   | the current file — context for the active item      |
| `<repo root>/DONE.md`      | the done file — the completion ledger               |
| `<repo root>/carves/`      | public carves                                       |
| `<repo root>/carves/done/` | finished carves, so they do not crowd out live ones |
| `<repo root>/.carves/`     | private carves — the dotted sibling, optional       |

**The repository root, not the design root.** The trajectory goes *across* techniques, so
its scope is the repository. One repository holds one work queue and may hold several
design roots — this project holds two, `tool/` and `example/` — so siting the queue under
either would be wrong. A carve is cross-approach by definition, which puts `carves/` at
the same level for the same reason.

**No project prefix.** The filenames are exactly the three above. A project that wants its
name on the queue puts it in the title (`# Pending: Ark State`), which nothing parses.

**`specs/migrations/` is a retired location.** It was the documented siting until
2026-08-04. Files found there are legacy and want moving; the directory still holds
in-flight migration specs, which are a different artifact.

**Privacy is declared in `.gitignore`, recorded as `track` in `.minispec/config.toml`,
and verified on every run.** `track = "private-trajectory"` requires the three queue files to
be ignored; `track = "all"` requires them not to be; `.carves/` is required ignored under
either, because a directory whose whole purpose is privacy does not become public because
the queue did. See `config-reference.md`.

---

## The pending file

```markdown
# Pending

<lifecycle preamble: ordering, the ID rule, what completing an item does>

---

## 8. **<title>** (<skill that runs it — or omit>). <one-line status>.
Source: [<doc>](<path>), part `#<key>`.
Next: <next action>.

## 12. **<title>** …
Source: [<design root>/design.md](<path>), gap `<gap ID>`.
```

- **Ordered by intent — the top item is active.** Position is the priority; the number is
  only the identifier, so the numbers run out of sequence and that is expected.
- **Each item is a `##` heading.** That is what lets a paused item's context nest beneath
  it as a sub-item.
- **Everything under the heading sits at column zero.** No hanging indent: it looked like
  one only until a line wrapped, and the tool never needed it. Files written under the older
  three-space stencil still read; the tool writes flush left.
- **An entry is a pointer, not the design.** Subject, the skill that runs it (or nothing),
  a one-line status, and one doc link. Rationale, findings and open questions live in the
  linked document. No instructions in entries.
- **The one-line status is part of the heading; the `Next:` line is a line of its own.** That
  is why `pending add-item` requires the status and not the next action: a tool may decline to
  write a line, but a heading it mints without the status is *two thirds of a line*, finished
  by hand in a file with no diff. Ten of ten entries carried both when the verb could produce
  neither.
- **Completing an item removes its entry.** Remaining entries keep their numbers.
- **A source is a carve part or a gap, told apart by shape.** The `Source:` line names the
  document as a markdown link, then either `part `` `#<key>` `` or `gap `` `<gap ID>` `` — the
  word says which, a part key carries `#` and a gap ID (`O136`, `R42`, `T7`) never does. A gap
  source names one ID: a range or a list is not a source, because an entry discharges one thing.
  Repairing a gap is the second most common thing an item does, and until the verb accepted the
  shape (2026-08-26) every such entry was typed by hand around it. A reader reports which kind it
  read; a writer emits one form for each and nothing else.

## The current file

```markdown
# Current

Working context for the active item only — never a log (that's the done file). To
pause: lift this into a sub-item under that item's `##` heading in the pending file,
then reset here, freeing it for what you pick up next.

---

## Active

_No active item._

## <a standing section>

<what outlives every item: answers already given, pointers that are expensive to find
again, the state of things>
```

- **The active item lives under `## Active`, and that section is the only region a tool
  may clear.** It ends at the next heading of the same level or higher, so the active
  item's context may use `###` and below as freely as it likes.
- **With nothing active the section holds a one-line placeholder**, not the last item's
  leftovers.
- **Standing context lives in its own `##` sections**, before or after the active one.
  It is what outlives every item and has nowhere else to live: answers already obtained
  from the user, pointers to work sitting outside this repository, the state of things.
  The pending file is per-item, the done file is history, and a carve is per-problem — so
  material that belongs to the project rather than to any one item is lost without this.
- **Exactly one `## Active`.** Two make the region ambiguous, and a tool that cannot tell
  which one it was asked to clear refuses rather than picks.
- **It is a resume buffer.** To change styles mid-flight, park the active item's context
  as a sub-item under its `##` heading in the pending file — the pending file is a stack
  you can push onto — freeing the current file, then resume later from the parked
  sub-item.
- **Never a log.** Finished work goes to the done file. Standing context is not a log: a
  log is *what happened*, which the done file owns, while standing context is *what is
  still true*, which nothing else holds.

*The heading was added 2026-08-18, after a reset verb deleted 320 lines of standing
context from a live current file.* Before it the active item was defined as **everything
after the `---` rule**, so a verb honouring that definition deletes whatever else is down
there — correctly targeted against a wrong model of the document. A named region can be
found; a region defined as *the rest of the file* can only be assumed. A file with no
`## Active` is legacy: add the heading beneath the rule, above the placeholder.

## The done file

```markdown
# Done

Completed items, most-recent first.

---

- **YYYY-MM-DD — #8 / R189–R198: <title>.** Part `<doc-path>#<key>`.
  <gaps banked, sources touched, what a future reader needs>
```

- **Most-recent-first.**
- **The identifiers lead.** Date, then what this discharged — the queue ID, the
  requirement range, or both — then the title. A reader scanning a long ledger sees the
  joins first, which is the whole reason they are at the front rather than buried in the
  title.
- **The identifier slot is flexible and deliberately so.** `#8` for a queue item,
  `O201` for a gap closed without one, `R3392–R3397` for a requirement range, several
  separated by `/`, or nothing at all for an incident that discharged no ID.
- **Each entry that came from the queue carries its queue ID.** It is the join a carve's
  status block points at, and an entry without one leaves that pointer dangling.
- **`Part <doc-path>#<key>` records the reverse link** — which part of which carve this
  discharged. The queue ID is the forward join; this is the one that lets you start from
  a landed change and find the reasoning behind it.
- **Enough to reconstruct the change without re-reading the code:** gaps banked, sources
  touched, and whatever a future reader would otherwise have to re-derive.

*This is ark's shape, adopted 2026-08-16 with the part pointer added — see the carve for
why the format changed rather than ark.*

## Item IDs

- **Assigned once at creation and never reused.** A gap in the sequence is expected and
  is never closed by renumbering: a queue entry pointing at a reused number resolves to
  different work with nothing able to detect it.
- **The next free ID is the maximum assigned anywhere — across the pending file *and* the
  done file.** Each alone misleads in its own direction: the pending file's maximum is too
  low right after several items complete, and the done file's is too low whenever the
  highest IDs are still live.
- **Do not write the next-free ID down.** A copied counter is a second copy of the
  numbering state, and it goes stale silently.

---

## Carves

A carve is one coherent problem decomposed into parts that may each need a different
skill. Whether a document earns promotion to a carve is judgment and lives in `SKILL.md`;
what follows is only its shape.

### The status block

A block at the top, before the body. A carve outlives the length anyone reads end to end,
so "what is still open?" has to be answerable from the first screen.

```markdown
## Status

- [x] ~~**Item 1 — record and resolve.**~~ **LANDED (2026-08-04 — `#3`.)**
- **Item 2 — the bootstrap.** **SPLIT (Bill, 2026-08-07.)** No checkbox: the sub-items
  carry the state.
  - [x] ~~**2.1 — root detection.**~~ **LANDED (2026-08-07 — `#1`.)**
  - [ ] **2.2 — the config move.** **OPEN (#8.)**
- [ ] **Item 4 — fail fast when onboarding does not take.** **OPEN (not queued.)**
```

- **Ordered by intent**, like the queue: position is the priority, the number is only the
  identifier.
- **The block owns the title and the status; the body owns the detail.** Further down,
  each part's elaboration is keyed by **bare number, with no title and no description** —
  `**Item 1** (…)`. Neither repeats the other, so there is no second copy to drift. A
  status *table* restating body prose is worse than none, because the two will disagree
  and nothing will say which is right. **A decision about one part is detail**, so it sits
  with that part rather than in `## Decisions` — see *Decisions* below for the test and the
  shape.
- **The machine view across every live carve is one grep:** `grep -rn '^- \[ \]' carves/*.md`.

### Parts, subparts, and the split rule

- **Parts are numbered**, and the number is the join between the one-line status and the
  elaboration below. Matching by title text breaks the moment a title is edited.
- **Subparts are namespaced by their parent and numbered from 1** — `3.1`, `3.2` — with an
  elaboration keyed by bare number exactly as a part has.
- **Numbering is append-only.** Once anything points at `3.2`, renumbering it leaves the
  pointer resolving to different work with nothing able to detect it. Gaps are expected.
- **Split until no part is completed by more than one item.** When a part would take
  several items, issue subparts so each is discharged by exactly one.
- **A parent keeps no checkbox.** It is complete when its subparts are — derived, not
  stored, because a parent box is a second copy of a fact the subparts already carry, and
  two copies are what disagree.
- **The number appears exactly at the scheduling boundary, and never at two levels for the
  same work.** Checkable: for any part, either the part carries a queue `#N` or its
  subparts do, not both.

**The link is asymmetric, deliberately.** A part records exactly one item; a queue entry
records a list of parts, because one item may discharge parts in several documents.
Splitting fixes many-items-per-part; it was never meant to forbid many-parts-per-item.

### Identifying a part

**`<doc-path>#<key>.<subkey>`** — for example `carves/trajectory-tool.md#9.1`.

The fragment resolves *inside* the named document, so the key scheme is a per-document
property and no project-wide convention is needed: `carves/x.md#7` is whatever `x.md` keys
as `7`, and cannot mean a `#7` from somewhere else.

**There is one key form, and the key is the fragment.** The status block writes
`**Item 9 — <title>**`, the elaboration below is keyed `**Item 9**`, subparts are `9.1`,
`9.2`. The **key** — what a pointer's fragment carries, what a `Source:` line names, what a
tool passes when it marks or lands a part — is the bare `9` or `9.1`; `Item` is the head's
display word and no part of the key. So `carves/x.md#9`, never `carves/x.md#Item 9`
(DECIDED, Bill, 2026-09-05).

**The word is required where the key carries no dot and forbidden where it does**, so
`Item 9` and `9.1` are heads and a bare `9` or an `Item 9.1` is not. **The separator
between the head and the title is the em dash** — `—`, not an en dash and not a hyphen.
Both were stated only by example until 2026-08-20, and a reader that admitted the
variants made two spellings of one key and three of one separator; measured across two
projects' 13 carves, all 63 keyed part lines already used the mandated forms, so nothing
was migrated to say so.

**A line breaking either is *unkeyed*** — listed by every read path with the shape it must
take, refused by every write path, and its checkbox still counted. That is the treatment
a superseded scheme already gets, for the reason given below: the state is legible even
where the address is not.

**Why this one, and not the queue's `#N`:** a part exists long before it is scheduled, and
most parts are unqueued at any moment, so a key drawn from the queue is undefined exactly
when you need to point at the part. `Item N` always resolves. `Part A`/`Part B` letters
carry no ordering and do not subdivide.

**`Item N` and `#N` are two keys doing two jobs, and both appear in a carve.** `Item N`
identifies *within the document*; `#N` is the queue ID and appears in the marker —
`**OPEN (#8.)**`, `**LANDED (2026-08-16 — `#10`.)**`. They are never
interchangeable, which is what makes a bare `#7` in a carve unambiguously a queue
reference.

**Older documents may use another form. Migrate them.** Two other schemes were in live use
before 2026-08-16 — the bare queue number, and `Part X` letters — and this document is the
target they migrate *to*. Rewriting the keys is mechanical; nothing depends on the old
form, and no backward compatibility is offered.

*The shape is deliberately the one already used for sequence anchors* (`Seq:
seq-foo.md#1.4`): same `file#dotted-fragment` syntax, a different resolver. Context always
says which, since a `Seq:` reference and a part pointer never appear in the same position.

**Trajectory files are private and public carves are not, so the public→private half of
the link is never a markdown link.** A carve carries a bare `#N`; the queue side holds the
recorded pointer back.

### Checkboxes and stamps

**A stamp records what happened; a checkbox records what is left.** Different jobs, and
the format needs both — but only the second is queryable, so only the second cannot be
optional.

- **Every list this format tracks progress on uses `- [ ]` / `- [x]`**, the same notation
  as the gaps list, so one query spans both. A section tracking progress in prose alone is
  invisible to the one view the format is built around: however carefully a human marked
  it up, `grep '^- \[ \]'` does not see it.
- **The interior is `[ ]` or `[x]` and nothing else.** Any other interior — `[X]`, a tick —
  counts toward neither state and is listed with `[x]` named as its target. A reader that
  took *any non-empty interior* as landed made three spellings of one state until
  2026-08-20; no carve part line in either project had ever written one.
- **The checkbox is authoritative** where the markings on a line disagree. A line states
  the same fact three ways — checkbox, strikethrough, marker — because three readers want
  it: a grep, a skimmer, and someone wanting provenance.
- **No checkbox *means* something.** Standing constraints, accepted residue,
  already-in-place preconditions, and split parents are listed without one: there is
  nothing to close. Same distinction an approved gap draws.

### Markers

Shape: `**VERB (attribution)**`, where the attribution is a date and the queue ID for work, or
a name and date for a judgment.

**The vocabulary is open; the shape is what gets checked.** The checkbox carries the
state a machine reads, so the verb never has to be parsed — which frees it to carry
information instead. Ten verbs are in live use and four of them were coined in a single
ten-day stretch; a set closed on any given day would refuse the words the next month
needs. **The table below is a census, not a permitted set.**

**A verb is written in capitals, and it may be several words joined by spaces or hyphens**
— `NOT VERIFIED`, `RE-CUT` — with lowercase appearing only inside the attribution's
parentheses. *Stated 2026-08-21 (Bill), because it never had been:* the reader admitted an
interior space and refused an interior hyphen, so `RE-CUT` was non-conforming from the day
it was first written here and nothing reported it. That is drift running its **unusual**
direction — the reader stricter than the format it enforces — which is why it was silent:
nothing goes red when a document obeys a rule stated nowhere. The reader's leniency is
unchanged and is not a second spelling: a verb typed in another case is still read,
canonicalized, and **rewritten in capitals** by the tool, which is what keeps a mis-cased
marker visible to the writer instead of accumulating a second queue state beside it.

| verb | means |
|---|---|
| `LANDED (date — #N)` | the work is in the repository; the commit is found by `git log --grep '#N'`, which is why every commit names the items it lands |
| `OPEN (#N)` / `OPEN (not queued)` | not started; queued under that ID, or not yet scheduled |
| `REVERTED (#N)` | the queue rolled back over this part's item (`pending revert`); replay returns it to `OPEN (#N)`, an abandoned attempt to `OPEN (not queued)` |
| `SPLIT (name, date)` | issued as subparts; this line keeps no checkbox |
| `MOVED (name, date — <where>)` | the part now lives in another document |
| `MIGRATED (date, commit)` | the decisions here are now specified elsewhere, which is the authority |
| `DISCHARGED (commit, date)` | satisfied without being the work it described — a blocker cleared, a precondition met |
| `SENT (name, date)` | discharged by a cross-project message; owns no commit |
| `DEFERRED (name, date)` | parked on purpose, as distinct from merely unstarted |
| `NOT VERIFIED` | the work happened; nobody has confirmed it does what was asked |
| `ASSESSED (name, date)` | examined and judged, with no change made |

**`OPEN` and `REVERTED` are the two *transient* verbs, and every other verb is a record.** A
transient carries the queue state and is *replaced* by the next transition — a tool writing one
replaces the transient already on the line, deletes any other, and never touches a record; a
record is appended beside whatever is there. That rule is the tool's (its backup spec) and the
distinction is stated here so a reader of the lexicon knows which two words it applies to.
*`REVERTED` joined this table on 2026-09-05, having been written by the tool since August without
the format ever naming it — which is how a reader built from this table came to treat it as a
record.*

**Coin a verb where it carries information.** Forcing `LANDED` onto a part that was sent,
deferred or moved makes the marker a worse record than the prose it replaced.

**State the negative.** A part with no marker is indistinguishable from one nobody has
thought about. `NOT VERIFIED` earns its own words in particular: a repaired *symptom*
reads exactly like a satisfied *requirement*, and that misreading is the most expensive
one this block prevents. No tool can validate it — the marker is visible, the claim is
not.

### Decisions

**`DECIDED (name, date)`**, append-only, in the body. A reader can then tell a settled
call from a musing, and whose it was. This is the single highest-value habit in the
format.

`LANDED (date — #N)` mirrors it deliberately — same shape, same greppability — and the item
number is what makes the claim findable: `git log --grep '#N'` reaches the commit. **No commit
hash** (Bill, 2026-09-15): a hash exists only after the commit, which forced every flip into a
second commit, and an amend or a squash rewrites it away. Older lines carry one; they are
read, and never written again.

**A decision naming exactly one part lives with that part; a decision spanning parts, or
governing the document as a whole, stays in `## Decisions`** — the carve's chronological
section for them. The test is mechanical rather than editorial: *does this decision name
exactly one part?* Chronology survives either way, since every decision already carries its
date. And a decision does not move because it grew important — it moves only because it
names one part.

**The sited shape is a `### Item N's decisions` subsection** under that part's elaboration,
dated where a group was taken in one sitting. This is codified rather than designed: two
carves reached for it independently before any rule said to, which is the cheapest evidence
available that a shape is the right one.

**What the rule buys is that a *partial* read of a carve becomes safe.** A carve keeps a
part's information in several places and only some are keyed by the part number: the status
line and the elaboration are both found by grepping `Item N`, while a decision in
`## Decisions` is keyed by nothing, so grep finds only the phrasings you can guess. Measured
2026-08-16 — everything about one part sat in **four places spanning 1047 lines**, two of them
unkeyed and missed until the document was read end to end, and one of those two had already
**superseded** the keyed elaboration on the day both were written. The stale text read as
current and the work was scoped wrong. Without the rule the standing mitigation is to read
every carve end to end before touching one part of it, which is a tax paid on every item
forever.

**An `@undecided:` sits where its decision will.** One naming a single part goes with that
part, by the same test, because it converts to `DECIDED` **in place** — a rule requiring a
move at the moment of conversion would be skipped exactly when someone is busy settling
something.

**Supersede in place.** When a decision overturns an earlier one, say so at the earlier
one rather than only at the new one. A dated `DECIDED` sitting beside an unmarked
paragraph that contradicts it will be read as current, because nothing about the unmarked
paragraph looks provisional.

### `@undecided:`

Marks a proposal that is not yet a decision — the state between unmarked thinking and
`DECIDED`. Named for the state rather than for an action, so nothing reads it as an
instruction to go and decide.

- **It carries what would settle it.** A marked fork with no stated resolution condition
  is just a hedge.
- **It converts to `DECIDED` when settled, or it is deleted.** It shares a stem with
  `DECIDED`, so one grep for `decid` spans both states.
- **Bare is a use; backquoted is a mention.** Written bare it is a real tag, so what is
  unsettled is findable across a corpus. Written in backquotes it is a mention, skipped by
  a reader's grep by dropping any hit whose next character is a backtick. One character
  does the work a markdown parser would otherwise have to do.
- **In a carve it is ordinary; in a spec it is a defect.** A carve is allowed to be
  unsettled. A spec describes how the system *is*, so an unresolved point inside one is a
  hole in the specification — the lightweight inline form of *blocked*, where no amount of
  effort closes the item and only an answer does.

---

## The tool/agent contract

**The rule is one line: the tool ingests by position.** What it reads as data comes from
fixed positions in the file. Everything else — every other line, every phrase inside one —
is yours to write freely, and no amount of prose can be mistaken for a record.

That is what makes your writing safe *by construction* rather than by inspection. You never
have to ask whether a sentence will be misread, because the tool is not reading sentences.

### What the tool reads

| file | position | what is taken |
|---|---|---|
| pending | `## N.` at column 0 | the item's number |
| pending | the `Source:` line beneath that heading | the document it links to, and the `` `#key` `` after `part` |
| done | `- **` at column 0 | the entry begins here |
| done | between the header's em dash and the colon that opens the title | every `#N` in that slot, and nothing outside it |
| done | a backquoted `` `<doc>#<key>` `` in the entry | the part it discharged |
| current | a level-2 `## Active` heading | the region a reset clears; it ends at the next heading of the same level or higher |
| carve | a level-2 `## Status` heading | the status block begins; it ends at the next heading of the same level or higher |
| carve | a `- [ ]` / `- [x]` list item **inside that block** | one part |
| part line | the checkbox | the state, and it is authoritative over the other two markings |
| part line | the bullet's leading whitespace | the depth — a subpart is an indented sibling |
| part line | `Item N` or `N.M` after any `~~` and `**` | the key — the bare `N` or `N.M`; `Item` is the display word |
| part line | `#N` inside a `**VERB (…)**` marker | the queue ID |

**Everything not in that table is prose.** A done entry's body may quote another item freely
— measured, five body lines in one ledger would contribute a queue ID if bodies were read.
A carve may hold examples of the format it is written in. Both are safe because neither sits
in a listed position.

### Why by position, and not by pattern

Because the alternative was tried, five times, and returned a confident wrong answer every
time. The most instructive is the fifth: an extractor sweeping `**VERB (…)**` across a whole
status block read a prose sentence quoting another project's shape — inside backticks, a
mention rather than a use — as a live citation, and reported a dangling `#121` in a
repository that never had one. It was written by someone who had read the record of the
other four while building the check that replaces them.

So the rule is not enforced by remembering it. It is enforced by the reader: a citation is
read from a part line's marker or not at all, and a code span is a code span.

### The companion rule: say what you could not ingest

Ingesting by position has a cost, and stating it is the other half of the contract. **A
shape-based check is blind by construction to anything outside the shape** — "does this
entry carry an identifier slot?" has no answer for a line never recognized as an entry — so
such a check reports *clean* over everything it never saw.

That is not hypothetical. Measured 2026-08-16 in one live ledger: **139 entry-like lines,
54 recognized.** The other 85 wrote the date outside the bold, an older shape. Every
shape-based check reported clean over 61% of the file, and the next-free-ID query answered
from what it read of the remaining 39% with nothing saying so.

**So a tool reading these files reports two things, always: what it found, and what it could
not read.** A count of unrecognized entry-like lines, and a count of references it could not
reach. Both are printed even when nothing else is, because a reader takes silence about
coverage as a claim of completeness.

For you as an author this is the useful half: if the tool says it could not read something,
that is a migration to perform, not a complaint to argue with. It names the target shape.

## What this file deliberately does not carry

These are judgment, and no tool and no format can hold them:

- **Why a carve exists**, and the coherence-versus-schedulability argument behind it.
- **Promotion to a carve** — the maintainer's call, explicitly not the agent's.
- **Priority ordering** — by intent, expressed as position rather than number.
- **Whether `NOT VERIFIED` is still true.** The marker is visible; the claim is not.
- **Migrate-on-landing** — where a part's decisions go once it lands, which is about
  authority rather than about file shape.

All of it is in `SKILL.md`.
