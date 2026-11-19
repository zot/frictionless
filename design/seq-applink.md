# Sequence: App Activation

**Source Spec:** applink.md
**Requirements:** R184, R185, R228, R230, R231, R232, R233, R234

## Participants

- Caller: `frictionless linkapp`, the `mcp` script, or Lua through AppLinkOps
- AppLink: the applink package
- FS: the file system
- Server: the MCP server

## Sequence: activate

```
1.1. Caller -> AppLink.Activate(baseDir, app)
  1.1.1. FS stat apps/{app}: not a directory -> error naming the path
  1.1.2. FS mkdir lua/
  1.1.3. lua/{app} is our link -> done
  1.1.4. lua/{app} is another link -> FS remove it
  1.1.5. makeLink: Unix symlink ../apps/{app}; Windows junction to the absolute path
  1.1.6. removeOldLinks
```

## Sequence: deactivate and delete

```
2.1. Caller -> AppLink.Deactivate(baseDir, app)
  2.1.1. FS lstat lua/{app}: absent -> done; a real directory -> error
  2.1.2. FS remove lua/{app}; removeOldLinks
2.2. Caller -> AppLink.Delete(baseDir, app)
  2.2.1. validName: empty, separator or ".." -> error, nothing touched
  2.2.2. Deactivate
  2.2.3. FS lstat apps/{app}: not a real directory -> error
  2.2.4. FS remove apps/{app} recursively
```

## Sequence: repair at start

```
3.1. Server.Start -> AppLink.Repair(baseDir, log)
  3.1.1. for each link in lua/ that does not resolve
    3.1.1.1. apps/{name} is a directory -> makeLink again
    3.1.1.2. otherwise -> log it, leave it
```

## Sequence: from Lua

```
4.1. Lua -> ui.linkApp(name) | ui.unlinkApp(name) | ui.deleteApp(name)
  4.1.1. AppLinkOps -> AppLink.Activate | Deactivate | Delete
  4.1.2. error -> raise it as a Lua error; else return nothing
```
