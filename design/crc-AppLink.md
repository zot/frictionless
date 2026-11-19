# AppLink
**Requirements:** R184, R185, R186, R187, R188, R226, R227, R228, R229, R230, R231, R232, R233, R235

The `applink` package (`github.com/zot/frictionless/applink`): app activation for the
frictionless binary, the `mcp` script and Lua. Its link code is lifted from ark's
`dirlink.go`, which solved Windows first; Frictionless keeps a relative target on Unix.

## Knows
- baseDir layout: `apps/{app}` (the app), `lua/{app}` (its activation link), `viewdefs/`
  (only for old-scheme cleanup)
- the link's target: `../apps/{app}` on Unix (R226); the absolute `apps/{app}` path on
  Windows, where a junction cannot be relative (R227)

## Does
- Activate(baseDir, app): `apps/{app}` must be a directory, else an error naming the path
  (R230); make `lua/` (R187); when `lua/{app}` is already our link, do nothing; a link
  there pointing elsewhere is removed; makeLink; removeOldLinks (R184, R188, R230).
- Deactivate(baseDir, app): nothing at `lua/{app}` -> nothing to do; a real directory there
  -> error, nothing removed; else remove the link; removeOldLinks (R185, R188, R231).
- List(baseDir): the names of the entries in `lua/` that are links to directories (R186).
- Delete(baseDir, app): validName; Deactivate; `apps/{app}` must be a real directory (Lstat:
  a directory and not a link), else an error; then remove it and everything in it (R232).
- validName(app): reject empty, a path separator, or `..` anywhere in the name (R232).
- Repair(baseDir, logf): for each link in `lua/` that no longer resolves: when
  `apps/{name}` is a directory, make the link again; otherwise log it and leave it. It
  only walks links that exist, so it never creates one (R228).
- makeLink(linkPath, app): Unix `os.Symlink("../apps/"+app, linkPath)` (R226); Windows a
  junction to the absolute app directory, written as a mount-point reparse point with
  FSCTL_SET_REPARSE_POINT, no other program run (R227). Per-platform files.
- isLink(fi): `ModeSymlink` or `ModeIrregular`; a Windows junction reads as the second
  under Lstat (R229).
- isOurLink(linkPath, app): the link resolves to `apps/{app}` (compared as absolute paths).
- removeOldLinks(baseDir, app): remove `lua/{app}.lua` when it is a link, and each link in
  `viewdefs/` whose target lies in the app's `viewdefs/` (R188).
- Command(args, baseDir, out): parse `add|remove|list [APP]`, run the operation, print a
  one-line result; a usage or operation error is returned to the caller (R233).

## Collaborators
- frictionless binary: `frictionless linkapp` calls Command (R233)
- MCPScript: `mcp linkapp` hands off to the binary (R233)
- AppLinkOps: the Lua bindings call Activate, Deactivate and Delete
- Server: Start calls Repair (R228)
- Install: no longer installs a `linkapp` script; an old one left in a base directory is
  no longer called by anything (R235)

## Sequences
- seq-applink.md
