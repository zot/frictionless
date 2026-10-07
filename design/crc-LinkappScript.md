# LinkappScript

**Requirements:** R85, R184, R185, R186, R187, R188

## Knows

- `BASE_DIR`: Script directory (from realpath)
- `APPS_DIR`: `{BASE_DIR}/apps`
- `LUA_DIR`: `{BASE_DIR}/lua`
- `VIEWDEFS_DIR`: `{BASE_DIR}/viewdefs` (only to clean up old-scheme links)

## Does

- **usage**: Display help and exit
- **check_app_exists**: Verify `apps/{app}` directory exists
- **add_app**: Create `lua/` if missing, link the app directory to `lua/{app}` (the activation link), then remove old-scheme links
- **remove_app**: Remove the `lua/{app}` link, then remove old-scheme links
- **remove_old_links**: Remove `lua/{app}.lua` and any `viewdefs/` symlink pointing into `apps/{app}/viewdefs/`
- **list_apps**: Scan `lua/` for directory links, report names

## Collaborators

- **MCPScript**: Invokes linkapp via delegation
- **ln**: Creates symbolic links
- **readlink**: Reads symlink targets for remove
