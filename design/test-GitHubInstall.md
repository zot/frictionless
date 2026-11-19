# Test Design: GitHubInstall
**Source:** crc-GitHubInstall.md, crc-GitHubOps.md

Investigating and installing used to be `curl`, `unzip`, `mv` and `rm` run from Lua on the
session's executor. These tests point the package's endpoints at a local `httptest` server
that serves a contents-API listing, raw files and an in-memory zip archive, so nothing
reaches GitHub.

## Test: a tree URL parses, and nothing else does
**Purpose:** the URL names user, repo, branch and path, and the app is the last segment (R236)
**Input:** `https://github.com/u/r/tree/main/apps/demo`; then a blob URL, a URL with no path, and a non-GitHub URL
**Expected:** u, r, main, apps/demo, name demo; each other URL is an error
**Refs:** crc-GitHubInstall.md
**Code:** githubinstall/githubinstall_test.go
**Alarm:** 1
**Fire alarm:** make ParseURL accept any URL with at least four path segments; the blob URL parses
**Inject:** githubinstall/githubinstall.go:ParseURL
@alarm-hash: bbff25d5e6760718
**Pulled:** 2026-11-19 — rang: `ParseURL("https://github.com/u/r/blob/main/apps/demo") succeeded, want an error`

## Test: investigate lists and fetches what the review shows
**Purpose:** both listings come back, and the text of every reviewable file is keyed by its relative path (R237, R238)
**Input:** a served app directory with requirements.md, design.md, app.lua, extra.lua, a README.md, and viewdefs/ holding App.DEFAULT.html and App.Item.list-item.html
**Expected:** Files lists every entry; Viewdefs lists both HTML files; Contents has requirements.md, design.md, app.lua, extra.lua and both `viewdefs/...` paths, and not README.md
**Refs:** crc-GitHubInstall.md
**Code:** githubinstall/githubinstall_test.go
**Alarm:** 2
**Fire alarm:** skip fetching `viewdefs/*.html` in Investigate; Contents lacks both viewdef paths
**Inject:** githubinstall/githubinstall.go:Investigate
@alarm-hash: 036b7e47eb0cd6fc
**Pulled:** 2026-11-19 — rang: `Contents["viewdefs/App.DEFAULT.html"] = "", want its text`

## Test: a missing file is absent, a missing listing is an error
**Purpose:** a file that cannot be fetched is left out without an error; a directory listing that cannot be fetched is the call's error (R239)
**Input:** a listing that names design.md while the raw server answers 404 for it; then a URL whose listing answers 404
**Expected:** the first investigation succeeds without design.md in Contents; the second returns an error
**Refs:** crc-GitHubInstall.md
**Code:** githubinstall/githubinstall_test.go
**Alarm:** 3
**Fire alarm:** make Investigate return an error when a file fetch fails; the first investigation errors
**Inject:** githubinstall/githubinstall.go:Investigate
@alarm-hash: 036b7e47eb0cd6fc
**Pulled:** 2026-11-19 — rang: `an unfetchable file failed the investigation: fetching design.md`

## Test: install extracts only the app, with its prefix removed
**Purpose:** entries under the archive's top directory and the app path land in apps/{name}; others do not (R241, R244)
**Input:** an archive `r-main/apps/demo/app.lua`, `r-main/apps/demo/viewdefs/D.html`, `r-main/README.md`, `r-main/apps/other/app.lua`; Install as demo
**Expected:** apps/demo/app.lua and apps/demo/viewdefs/D.html exist with their content; nothing else under apps/ but demo; apps/demo is mode 0755 like any app directory
**Refs:** crc-GitHubInstall.md
**Code:** githubinstall/githubinstall_test.go
**Alarm:** 4
**Fire alarm:** extract every entry with only the top directory stripped; apps/demo holds README.md and apps/other
**Inject:** githubinstall/githubinstall.go:extract
@alarm-hash: 72d4c828ee887daf
**Pulled:** 2026-11-19 — rang: `apps/demo holds [README.md apps], want [app.lua viewdefs]`

## Test: an existing app is never overwritten
**Purpose:** install refuses when apps/{name} exists, changing nothing (R240)
**Input:** apps/demo exists holding keep.txt; Install as demo
**Expected:** an error; the server received no archive request; keep.txt is unchanged; no temporary directory remains
**Refs:** crc-GitHubInstall.md
**Code:** githubinstall/githubinstall_test.go
**Alarm:** 5
**Fire alarm:** drop the existence check; the archive is downloaded before the install fails. Checking keep.txt alone would not ring: renaming onto a non-empty directory fails anyway, so the app survives either way
**Inject:** githubinstall/githubinstall.go:Install
@alarm-hash: 65b013133f5de1e9
**Pulled:** 2026-11-19 — rang: `the archive was requested 1 times, want 0: the existence check must come first` (re-pulled: Install gained the 0755 chmod) *Earlier —* 2026-11-19 — rang: `the archive was requested 1 times, want 0: the existence check must come first`
## Test: an entry escaping the app fails the install
**Purpose:** zip-slip is refused and nothing is left behind (R242, R244)
**Input:** an archive with `r-main/apps/demo/app.lua` and `r-main/apps/demo/../../escaped.txt`; Install as demo
**Expected:** an error; no escaped.txt anywhere under the base; no apps/demo; no temporary directory
**Refs:** crc-GitHubInstall.md
**Code:** githubinstall/githubinstall_test.go
**Alarm:** 6
**Fire alarm:** drop the escape check in the extraction; escaped.txt is written outside the app
**Inject:** githubinstall/githubinstall.go:extract
@alarm-hash: 72d4c828ee887daf
**Pulled:** 2026-11-19 — rang: `an escaping entry installed, want an error`

## Test: an oversize archive is refused
**Purpose:** the download and the extracted total are capped (R243)
**Input:** MaxBytes lowered to 1 KB; an archive whose app file holds 4 KB
**Expected:** an error; no apps/demo; no temporary directory
**Refs:** crc-GitHubInstall.md
**Code:** githubinstall/githubinstall_test.go
**Alarm:** 7
**Fire alarm:** drop the extracted-bytes check; the 4 KB file is installed
**Inject:** githubinstall/githubinstall.go:extract
@alarm-hash: 72d4c828ee887daf
**Pulled:** 2026-11-19 — rang: `a 4 KB app installed under a 1 KB cap, want an error`

## Test: the host operations deliver into their fields
**Purpose:** `ui.githubInvestigate` writes `{files, viewdefs, contents}` and `ui.githubInstall` writes true, each then running its callback (R239, R245)
**Input:** a session whose endpoints point at the local server; Lua calls both operations with callbacks that record their arguments
**Expected:** the investigation's `contents["app.lua"]` is the served text; the install field is true and `apps/demo` exists; both callbacks ran with no error
**Refs:** crc-GitHubOps.md
**Code:** internal/mcp/github_ops_test.go

## Not tested, deliberately
**No other program runs (R246), and no credentials are sent (R247).** Both are absences:
the package imports no `os/exec` and sets no authorization header. A reader checks them by
looking; a test could only restate the import list.

**App-console's callback flow.** It is Lua driving a live form; it is checked live, the
way Item 5's checkpoint status was.
