# Sequence: Checkpoint

**Source Spec:** checkpoint.md
**Requirements:** R79, R192, R193, R194, R203, R204, R205, R209, R210, R211, R212, R213, R214, R215, R216

Every step that runs fossil is checked (R210). "verify" is the binding check that stops the
command when fossil would act on any checkout but the app's own (R211).

## Sequence: a command

```
1.1. frictionless checkpoint CMD APP [ARG] --dir DIR
  1.1.1. Command(args, DIR/apps, stdout, onChange)
  1.1.2. Find -> fossil, or the install instructions as an error
  1.1.3. App{DIR/apps/APP, fossil}; missing directory -> error
  1.1.4. run the verb; print its result to out
  1.1.5. changing verb succeeded -> onChange: POST the refresh Lua to DIR's server, ignoring every failure
1.2. error -> the binary prints it on stderr and exits 1; success exits 0
```

## Sequence: baseline

```
2.1. no repository -> init, commit baseline, done (2.9)
2.2. ensureCheckout; verify
2.3. branch list, parsed to exact names; keep those named updates or local
2.4. mkdir aside (Dir/.#baseline.XXXXXX)
2.5. each kept branch -> bundle export to aside/BRANCH.bundle; failure -> remove aside, error naming the branch
2.6. verify, immediately before anything moves
2.7. rename checkpoint.fossil and .fslckout into aside
2.8. init (open --nested --keep, verify, settings, add .); commit baseline
2.9. each bundle -> bundle import --force --publish; then branch list must hold the branch by exact name
2.10. success -> remove aside; print "Baseline set" (with "preserved branches restored" when any)
2.11. any failure from 2.7 on -> remove the new repository and .fslckout, rename the old ones back, remove aside, error
```

## Sequence: branch save

```
3.1. ensureCheckout; verify
3.2. checkout BRANCH --keep (the working tree is untouched)
  3.2.1. "cannot resolve name": addremove; commit --branch BRANCH; checkout trunk --keep; done
3.3. addremove; commit
  3.3.1. "closed leaf": tag cancel --raw closed BRANCH; commit once more
3.4. checkout trunk --keep
```

## Sequence: changed since install

```
4.1. no original.fossil -> false
4.2. tarball trunk into a temp file, -R original.fossil
4.3. read the tarball: path -> content, its top directory stripped
4.4. walk Dir: skip fossil's own files and ignoreGlob matches
4.5. any file differing, present on one side only -> true
```
