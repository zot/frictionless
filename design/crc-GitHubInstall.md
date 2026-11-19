# GitHubInstall
**Requirements:** R236, R237, R238, R239, R240, R241, R242, R243, R244, R246, R247, R248

The `githubinstall` package (`github.com/zot/frictionless/githubinstall`): investigating
and installing an app from a GitHub tree URL, with Go's HTTP client and `archive/zip`. It
runs no other program (R246).

## Knows
- Source: User, Repo, Branch, Path, Name (the last segment of Path) (R236)
- endpoints, package variables so tests can serve them locally (R247):
  - APIBase `https://api.github.com` (directory listings: `/repos/{u}/{r}/contents/{path}?ref={b}`)
  - RawBase `https://raw.githubusercontent.com` (file text: `/{u}/{r}/{b}/{path}/{file}`)
  - ArchiveBase `https://github.com` (the branch zip: `/{u}/{r}/archive/refs/heads/{b}.zip`)
- MaxBytes: 100 MB, the cap on the archive and on its extracted files (R243); a package
  variable so tests can lower it
- Entry: Name, Type (`file` or `dir`)
- Investigation: Files and Viewdefs ([]Entry), Contents (relative path to text)

## Does
- ParseURL(url): match `https://github.com/{u}/{r}/tree/{b}/{path}`; Name is path's last
  segment; any other shape is an error (R236).
- list(ctx, src, dir): GET the contents API, decode the JSON array into Entries. A non-200
  status or bad JSON is an error.
- fetch(ctx, src, rel): GET the raw file; a non-200 status is "not found" (no error value
  surfaces to the caller, see Investigate).
- Investigate(ctx, src): list the app directory (error on failure) and `viewdefs/` (an
  absent `viewdefs/` gives an empty list); fetch `requirements.md`, `design.md`, every
  `.lua` file and every `viewdefs/*.html`; a fetch that fails leaves that path out of
  Contents (R237, R238, R239).
- Install(ctx, src, appsDir, name): a name that is empty or holds a separator or `..` -> error (R248); `apps/{name}` exists -> error, nothing changed (R240);
  download the archive through a reader capped at MaxBytes (R243); make a temporary
  directory `apps/.#install.XXXXXX` (mode 0755, as any app directory); for each zip entry under `{top}/{Path}/`, strip that
  prefix, refuse a path that escapes the app (absolute, or `..` after cleaning) (R242),
  count extracted bytes against MaxBytes (R243), write it; then rename the temporary
  directory to `apps/{name}`; any failure removes the temporary directory (R241, R244).
- AsLua(Investigation): the plain-value form delivered to Lua, `{files, viewdefs,
  contents}` with each entry `{name, type}` (R239).

## Collaborators
- GitHubOps: the host operations that run Investigate and Install off the executor
- net/http, archive/zip: the network and the archive

## Sequences
- seq-github-install.md
