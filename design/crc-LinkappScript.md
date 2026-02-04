# LinkappScript

**Requirements:** R80, R81, R82, R83, R84, R85, R86, R87

## Knows

- `BASE_DIR`: Script directory (from realpath)
- `APPS_DIR`: `{BASE_DIR}/apps`
- `LUA_DIR`: `{BASE_DIR}/lua`
- `VIEWDEFS_DIR`: `{BASE_DIR}/viewdefs`

## Does

- **usage**: Display help and exit
- **check_app_exists**: Verify `apps/{app}` directory exists
- **add_app**: Create directories, link app.lua, link app directory, link viewdefs
- **remove_app**: Remove lua file symlink, remove lua directory symlink, scan/remove viewdef symlinks
- **list_apps**: Scan lua/ for .lua symlinks, report names

## Collaborators

- **MCPScript**: Invokes linkapp via delegation
- **ln**: Creates symbolic links
- **readlink**: Reads symlink targets for remove
