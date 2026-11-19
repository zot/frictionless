# App Activation

**Language:** Go
**Environment:** the `applink` package (`github.com/zot/frictionless/applink`), imported
by the frictionless binary; ark has its own copy of the link code and may import this one

## Purpose

An app is active when `{base_dir}/lua/{app}` links to its directory; that one link is the
whole on/off bit. ui-engine reads the app's entry file, modules and viewdefs through it
(see ui-engine's app-discovery spec).

Activation lives in one Go package, so it works the same from the command line, from the
`mcp` script and from Lua, and so it works on Windows. Before this it was a bash script,
`linkapp`, which app-console ran through `.ui/mcp`; on Windows neither bash nor `ln`
is there. The link code comes from ark, which solved Windows first.

## Directory Structure

```
{base_dir}/
├── apps/
│   └── {app}/
│       ├── app.lua           # Entry file, loaded as lua/{app}/app.lua
│       ├── json.lua          # Optional module files: require("{app}.json")
│       └── viewdefs/*.html
└── lua/
    └── {app} -> ../apps/{app}   # Activation link
```

## The link

- **On Unix** the link is a symbolic link with a relative target, `../apps/{app}`, so the
  whole `{base_dir}` can be moved and every link still resolves.
- **On Windows** the link is a directory junction. A junction needs no administrator
  privilege and no Developer Mode, and it is made without running any other program. Its
  target is necessarily absolute, so a moved `{base_dir}` leaves its links pointing at the
  old place: the server repairs them when it starts. A link that no longer resolves, and
  whose app directory exists under `apps/`, is made again; one whose app is gone is
  reported and left alone. Repair never creates a link that was not there, so a
  deactivated app stays deactivated.
- **Either way**, a junction counts as a link: on Windows it does not report itself as a
  symbolic link, so telling a link from a real directory checks for both.

## Operations

| Operation | Command | Lua |
|-----------|---------|-----|
| Activate | `frictionless linkapp add APP` | `ui.linkApp(name)` |
| Deactivate | `frictionless linkapp remove APP` | `ui.unlinkApp(name)` |
| List | `frictionless linkapp list` | |
| Delete | | `ui.deleteApp(name)` |

The `mcp` script's `linkapp` command hands off to the binary, as its `checkpoint` command
does (`specs/helper-scripts.md`).

**Activate:**
- The app's directory must exist under `apps/`; otherwise it is an error naming the path.
- Creates `{base_dir}/lua/` if missing.
- Links the app directory to `lua/{app}`. Activating an app that is already linked to its
  own directory does nothing; a link there pointing elsewhere is replaced.
- Removes the app's old-scheme links if present: `lua/{app}.lua`, and any `viewdefs/`
  symlink pointing into the app's `viewdefs/`.

**Deactivate:**
- Removes the `lua/{app}` link, and the same old-scheme links as Activate.
- A real directory at `lua/{app}`, not a link, is someone's data: it is an error and
  nothing is removed. Deactivating an app that is not active does nothing.

**List:**
- Reports the names of the directory links in `lua/`.

**Delete** (for app-console):
- Deactivates the app, then removes `apps/{app}` and everything in it.
- The name must be a plain name: one containing a path separator or `..`, or empty, is an
  error and nothing is touched.
- Only a real directory at `apps/{app}` is removed. A link there, or a file, is an error
  and nothing is removed.

From Lua, each operation returns nothing on success and raises an error naming what went
wrong otherwise. They run synchronously: each is a few file-system calls.
