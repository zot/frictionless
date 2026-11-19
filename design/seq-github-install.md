# Sequence: Installing an App from GitHub

**Source Spec:** github-install.md
**Requirements:** R236, R239, R240, R241, R242, R243, R244, R245

## Participants

- AppConsole: the GitHub downloader in app-console (Lua, on the session's executor)
- GitHubOps: the host operations
- GitHubInstall: the package, running off the executor
- GitHub: the API, raw-file and archive endpoints

## Sequence: investigate

```
1.1. AppConsole.investigate() -> _investigating = true; ui.githubInvestigate(self, "_investigation", url, cb)
  1.1.1. GitHubOps op: ParseURL; bad URL -> delivered as an error at once
1.2. off the executor: GitHubInstall.Investigate
  1.2.1. GitHub list app dir; list viewdefs/
  1.2.2. GitHub fetch requirements.md, design.md, *.lua, viewdefs/*.html
1.3. delivery stitch: _investigation = {files, viewdefs, contents} (or the error); cb runs
  1.3.1. cb: _investigating = false; validate the listing; build tabs from contents
```

## Sequence: install

```
2.1. AppConsole.approve() -> _installing = true; ui.githubInstall(self, "_install", url, name, cb)
2.2. off the executor: GitHubInstall.Install
  2.2.1. apps/{name} exists -> error
  2.2.2. GitHub download the archive (capped)
  2.2.3. extract the app's entries into apps/.#install.XXXXXX (escape and size checks)
  2.2.4. rename to apps/{name}; any failure removes the temporary directory
2.3. delivery: cb runs
  2.3.1. error -> _installing = false; show it; stop
  2.3.2. ui.linkApp(name); write source.txt; ui.checkpointInstall(self, "_installed", name, cb2)
2.4. cb2: _installing = false; close the form; refresh the app list; select the new app
```
