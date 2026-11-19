# Test Design: AppLink
**Source:** crc-AppLink.md, crc-AppLinkOps.md

App activation was a bash script, `install/linkapp`, tested by running it
(`test-AppActivation.md`'s former alarm 6). Its behaviour moves into the `applink` package;
these tests drive the package, its command, and the Lua bindings, each over a temporary
base directory.

## Test: activate makes one relative link and cleans up the old scheme
**Purpose:** activation links `lua/{app}` to `../apps/{app}`, creates `lua/`, and removes old-scheme links (R184, R187, R188, R226)
**Input:** a base with `apps/demo/viewdefs/Demo.DEFAULT.html`, no `lua/`, an old `lua/demo.lua` link, and a `viewdefs/Demo.DEFAULT.html` link into the app; Activate demo
**Expected:** `lua/demo` is a link whose target reads exactly `../apps/demo`; both old links are gone
**Refs:** crc-AppLink.md
**Code:** applink/applink_test.go
**Alarm:** 1
**Fire alarm:** skip removeOldLinks in Activate; the old `lua/demo.lua` link survives
**Inject:** applink/applink.go:Activate
@alarm-hash: 2b980aad63c90705
**Pulled:** 2026-11-19 — rang: removeOldLinks skipped in Activate; "old-scheme link lua/demo.lua survived Activate (R188)"

## Test: the Unix link target is relative
**Purpose:** a moved base directory keeps its links resolving (R226)
**Input:** activate demo in a base, then rename the base directory
**Expected:** `lua/demo` in the renamed base still resolves to its `apps/demo`
**Refs:** crc-AppLink.md
**Code:** applink/applink_test.go
**Alarm:** 2
**Fire alarm:** make the Unix makeLink use the absolute app path; after the rename the link no longer resolves
**Inject:** applink/link_unix.go:makeLink
@alarm-hash: 9c67c313d86a77dd
**Pulled:** 2026-11-19 — rang: the Unix makeLink given the absolute app path; "after moving the base, lua/demo no longer resolves ... (R226)"

## Test: activate is idempotent and replaces a foreign link
**Purpose:** activating twice changes nothing, and a link elsewhere is replaced (R230)
**Input:** activate demo twice; then point `lua/demo` at another directory and activate again
**Expected:** no error each time; finally `lua/demo` targets `../apps/demo`
**Refs:** crc-AppLink.md
**Code:** applink/applink_test.go
**Alarm:** 3
**Fire alarm:** in Activate, treat any existing link as ours; the foreign link is left in place
**Inject:** applink/applink.go:Activate
@alarm-hash: 2b980aad63c90705
**Pulled:** 2026-11-19 — rang: Activate treats any existing link as ours; `foreign link not replaced: lua/demo -> "../apps/other" (R230)`

## Test: activating a missing app is an error
**Purpose:** an app must exist under `apps/` (R230)
**Input:** Activate ghost, with no `apps/ghost`
**Expected:** an error naming `apps/ghost`; no `lua/ghost`
**Refs:** crc-AppLink.md
**Code:** applink/applink_test.go

## Test: deactivate refuses a real directory
**Purpose:** a real directory at `lua/{app}` is someone's data (R185, R231)
**Input:** an empty real directory at `lua/demo`; Deactivate demo. Then activate demo and Deactivate it; then Deactivate it again
**Expected:** first an error and the directory remains; then the link is removed; the third call does nothing and returns no error
**Refs:** crc-AppLink.md
**Code:** applink/applink_test.go
**Alarm:** 4
**Fire alarm:** drop the real-directory check from Deactivate; the empty directory is removed
**Inject:** applink/applink.go:Deactivate
@alarm-hash: aca022c07b9ae7d7
**Pulled:** 2026-11-19 — rang (second attempt; the first removed the whole case and broke the build, which is not a ring): the real-directory case disabled; "Deactivate removed a real directory without error (R231)"

## Test: list reports only directory links
**Purpose:** active apps are the links in `lua/`, not files or real directories (R186)
**Input:** activate demo; also put a file `lua/main.lua` and a real directory `lua/plain` in `lua/`
**Expected:** List returns exactly `demo`
**Refs:** crc-AppLink.md
**Code:** applink/applink_test.go
**Alarm:** 5
**Fire alarm:** make List report every entry in `lua/`; it reports main.lua and plain as well
**Inject:** applink/applink.go:List
@alarm-hash: 3063d8f3567f0281
**Pulled:** 2026-11-19 — rang: List reports every entry in lua/; "List = [demo main.lua plain], want [demo] (R186)"

## Test: delete guards its name and its target
**Purpose:** delete removes only a real app directory named plainly (R232)
**Input:** names "", "a/b", "..", "../victim", with a directory `victim` beside `apps/`; a link at `apps/linked`; then a real activated `apps/demo`
**Expected:** each bad name is an error and `victim` still exists; `apps/linked` is an error and its target survives; demo is deactivated and `apps/demo` is gone
**Refs:** crc-AppLink.md
**Code:** applink/applink_test.go
**Alarm:** 6
**Fire alarm:** make validName accept every name; deleting "../victim" removes the directory beside `apps/`
**Inject:** applink/applink.go:validName
@alarm-hash: 1930280bee97638c
**Pulled:** 2026-11-19 — rang on the empty name rather than on ../victim: with validName disabled, `Delete("")` removed all of apps/; ../victim was stopped anyway, because Deactivate refused lua/../victim as a real directory

## Test: repair remakes only stale links
**Purpose:** a link that no longer resolves is remade when its app exists, left when it does not, and no link is created (R228)
**Input:** `lua/demo` a dangling link with `apps/demo` present; `lua/gone` dangling with no `apps/gone`; `apps/idle` present with no link; Repair
**Expected:** `lua/demo` resolves to `apps/demo`; `lua/gone` is still dangling and was logged; `lua/idle` does not exist
**Refs:** crc-AppLink.md
**Code:** applink/applink_test.go
**Alarm:** 7
**Fire alarm:** make Repair link every directory under `apps/`; `lua/idle` is created
**Inject:** applink/applink.go:Repair
@alarm-hash: 036c15d4420fb438
**Pulled:** 2026-11-19 — rang: Repair links every directory under apps/; "Repair created a link for an app that was not active (R228)"

## Test: a junction counts as a link
**Purpose:** isLink treats Windows' `ModeIrregular` as a link (R229)
**Input:** isLink on file infos reporting `ModeSymlink`, `ModeIrregular`, `ModeDir`, and a plain file
**Expected:** true, true, false, false
**Refs:** crc-AppLink.md
**Code:** applink/applink_test.go
**Alarm:** 8
**Fire alarm:** make isLink test only `ModeSymlink`; the `ModeIrregular` case reports false
**Inject:** applink/applink.go:isLink
@alarm-hash: cfbc7ced0da07fd3
**Pulled:** 2026-11-19 — rang: isLink tests only ModeSymlink; "isLink(?---------) = false, want true (R229)"

## Test: the command runs each verb
**Purpose:** `frictionless linkapp add|remove|list` drives the package (R233)
**Input:** Command with `add demo`, `list`, `remove demo`, and with no arguments
**Expected:** the link appears, list prints demo, the link goes; no arguments is a usage error
**Refs:** crc-AppLink.md
**Code:** applink/applink_test.go

## Test: the Lua bindings succeed quietly and raise on error
**Purpose:** `ui.linkApp`, `ui.unlinkApp` and `ui.deleteApp` return nothing on success and raise otherwise (R234)
**Input:** in a session over a temporary base: `ui.linkApp("demo")`; `pcall(ui.deleteApp, "../x")`
**Expected:** the link exists and the call returned no value; the pcall reports false with an error mentioning the name
**Refs:** crc-AppLinkOps.md
**Code:** internal/mcp/applink_ops_test.go
**Alarm:** 9
**Fire alarm:** make the bindings ignore the operation's error and return normally; the pcall reports true
**Inject:** internal/mcp/applink_ops.go:Server.defineAppLinkOps
@alarm-hash: c462a74de2f3b31e
**Pulled:** 2026-11-19 — rang: the bindings ignore the operation error; `pcall(ui.deleteApp, "../x") = "true|nil"` (R234)

## Not tested, deliberately
**The Windows junction (R227).** It is made by a Windows-only file through
FSCTL_SET_REPARSE_POINT and cannot run on the Linux machine these tests run on. The code is
ark's, which measured it on Windows (build 26200); its link test (R229) runs everywhere.
