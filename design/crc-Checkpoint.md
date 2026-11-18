# Checkpoint
**Requirements:** R69, R70, R71, R72, R73, R74, R75, R78, R79, R125, R126, R127, R192, R193, R194, R202, R203, R204, R205, R206, R207, R208, R209, R210, R211, R212, R213, R214, R215, R216, R217

The `checkpoint` package: per-app fossil checkpoints, imported by the frictionless binary
and by ark. Lifted from ark's `cmd/ark/checkpoint.go`, with Frictionless's baseline safety
merged in, so each version's hardening is in one place.

## Knows
- App: `Dir` (the app directory) and `Fossil` (the binary); the repository is
  `Dir/checkpoint.fossil` (R78)
- ErrNoFossil, and the install instructions it carries (R206, R207)
- baselineMsg `=== BASELINE ===`; the preserved branches, `updates` and `local` (R127)
- ignoreGlob `.#*,.*~,*~,checkpoint.fossil,.fslckout`, set on every open
- aside: `Dir/.#baseline.XXXXXX` during a baseline, holding the old repository, its
  checkout database and one bundle per preserved branch; dot-prefixed so `add` never
  captures it, and inside `Dir` so moving the repository is a rename (R192)

## Does
- Find: `~/.claude/bin/fossil` (or `fossil.exe` on Windows), then `exec.LookPath("fossil")`;
  ErrNoFossil otherwise (R206, R207).
- step: run one fossil command in Dir; a failure returns an error naming the command and
  carrying fossil's output (R210).
- verify: ask fossil (`info`) for the repository and checkout root it would use from Dir;
  error unless both resolve to the app's own (R211).
- ensureCheckout: with no checkout open in Dir, open one `--nested --keep`; verify; apply
  settings. A checkout bound to another repository errors (R212).
- init: `init`, open `--nested --keep`, verify, settings, `add .` (R78, R212).
- settings: `mtime-changes off`, ignoreGlob (R214).
- Save, List, Rollback, Diff, Count: ark's verbs, returning values (R69-R72, R75, R208).
- Update, Local: branchSave: retarget the checkout with `checkout BRANCH --keep`, create
  the branch when missing, reopen a closed leaf, commit, `checkout trunk --keep`; never
  `update` (R125, R126, R213).
- Baseline (and clear): see seq-checkpoint.md#2; branches matched by exact name; the old
  repository set aside, restored byte for byte on any failure (R73, R74, R127, R192-R194).
- Changed: false without `original.fossil` (R217); else export trunk's tip from it as one
  tarball and compare against the files in Dir, ignoring fossil's own files and
  ignoreGlob (R215, R216).
- Command: parse `CMD APP [ARG]`; Find; resolve `appsDir/APP`; run the verb; print its
  result to out; call onChange after save, rollback, clear, baseline, update and local
  (R79, R204, R209).

## Collaborators
- fossil: the binary, through step
- frictionless binary: `frictionless checkpoint` calls Command with an onChange that
  posts the app-console refresh to the running server, best effort (R203, R205)
- MCPScript: hands `mcp checkpoint` off to the binary
- ark: imports the package (R202)

## Sequences
- seq-checkpoint.md
