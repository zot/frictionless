# Apps - Testing

## Checklist

- [✓] App list displays all apps with status badges
- [ ] Clicking an app shows its details
- [ ] Build button triggers build_request event
- [ ] Test button triggers test_request event
- [ ] Fix Issues button triggers fix_request event
- [ ] Open button opens embedded view (disabled for apps/mcp)
- [ ] New app form creates directory and requirements.md
- [ ] Chat panel sends messages with quality setting
- [ ] Lua console executes code and shows output
- [✓] Panel tabs switch between Chat and Lua
- [ ] Refresh rescans apps from disk
- [ ] Gaps indicator shows warning icon when TESTING.md has non-empty Gaps section
- [✓] Checkpoint icon shows rocket for apps with checkpoints, gem otherwise
- [ ] Make it thorough button appears when app has checkpoints
- [ ] Make it thorough button triggers consolidate_request event
- [ ] Review Gaps button appears when app has gaps
- [ ] Review Gaps button triggers review_gaps_request event
- [✓] Hammer icon shows for unbuilt apps
- [ ] Checkpoint icon hidden for unbuilt apps
- [✓] Delete App button shows for non-protected apps
- [ ] Delete confirmation dialog appears when Delete App clicked
- [ ] Confirming delete removes app globals, unlinks, and deletes directory
- [ ] Protected apps (app-console, mcp, claude-panel, viewlist) cannot be deleted
- [✓] GitHub download button opens form
- [ ] GitHub URL validation rejects invalid URLs
- [ ] GitHub URL validation rejects directories missing required files
- [ ] GitHub name conflict detection shows warning for existing apps
- [ ] GitHub file tabs show all inspectable files
- [ ] GitHub tab labels show warning counts for Lua files
- [ ] GitHub tab clicking marks tab as viewed
- [ ] GitHub Approve button disabled until all tabs viewed
- [ ] GitHub file content shows pushState highlighting
- [ ] GitHub file content shows os.execute/io.popen highlighting
- [ ] GitHub scrollbar trough shows warning position markers
- [ ] GitHub approve downloads and installs app

## Gaps

### Fast Code

The following were added via `/ui-fast` checkpoints and may need review:

- `_isDownloaded`, `_hasLocalChanges`, `sourceUrl`, `readmePath` fields - downloaded app tracking
- `localChangesIcon()`, `showLocalChangesIcon()`, `hideLocalChangesIcon()` - local changes indicator
- `isDownloaded()`, `hasLocalChanges()`, `noLocalChanges()` - downloaded app state
- `hasSourceUrl()`, `noSourceUrl()`, `hasReadme()`, `noReadme()` - source URL state
- `readmeLinkHtml()` - HTML anchor for MCP readme endpoint
- `openReadme()`, `openSourceUrl()`, `checkLocalChanges()` - actions
- `findReadme()` - constructs readme URL from GitHub source URL
- Source URL row in detail header with link icons
- Pencil icon for apps with local changes
