# Installing Apps from GitHub

**Language:** Go
**Environment:** a small package in the frictionless binary, with two asynchronous host
operations defined on every session beside the checkpoint operations

## Purpose

App-console downloads apps from GitHub: it investigates a repository directory so the user
can review its files, then installs the app. Before this, both steps ran shell commands
(`curl`, `unzip`, `mv`, `rm`) synchronously on the session, so the whole session stalled
while the network worked, and none of it ran on Windows. Now Go does the network and file
work off the session, and app-console learns the result through a host operation's
fields and callback (ui-engine's `specs/host-ops.md`).

## The URL

A GitHub tree URL, `https://github.com/{user}/{repo}/tree/{branch}/{path}`, names one
directory in one branch. The app's name is the last segment of `{path}`. A URL of any
other shape is an error.

## Investigating

```lua
ui.githubInvestigate(target, field, url, cb)
```

Fetches what the review needs, in one call:

- The listing of the app's directory, and the listing of its `viewdefs/` directory, each a
  list of `{name, type}` entries where type is `file` or `dir`.
- The text of every file the review can show: `requirements.md`, `design.md`, every `.lua`
  file in the directory, and every `.html` file in `viewdefs/`, keyed by its path relative
  to the app directory (`app.lua`, `viewdefs/App.DEFAULT.html`).

It writes `target[field] = {files = ..., viewdefs = ..., contents = ...}`. A file that
cannot be fetched is absent from `contents`; that alone is not an error, since the
review decides what is required. A URL that does not parse, or a directory listing that
cannot be fetched, is an error written to `target[field .. "Error"]`. The deadline is 60
seconds.

## Installing

```lua
ui.githubInstall(target, field, url, name, cb)
```

Downloads the branch as one zip archive and extracts the app's directory into
`{base_dir}/apps/{name}`:

- A name that is empty, or contains a path separator or `..`, is an error and nothing is
  changed, so the app always lands directly under `apps/`.
- If `apps/{name}` already exists, it is an error and nothing is changed.
- Only entries under the app's directory in the archive are extracted (the archive's one
  top-level directory, then `{path}`), with that prefix removed.
- An entry whose path would land outside the app's directory is an error, and the
  install fails.
- An archive larger than 100 MB, or one whose files total more than 100 MB, is refused.
- Files are extracted into a temporary directory under `apps/` whose name begins with a
  dot, so it is never taken for an app, and moved into place only when everything has
  been extracted. Any failure removes the temporary directory; `apps/{name}` is either
  the complete app or absent.
- On success it writes `target[field] = true`; any failure is written to
  `target[field .. "Error"]`. The deadline is 120 seconds.

## Both operations

- They return at once and never run another program.
- They use GitHub's public endpoints without credentials, as the shell commands did.
