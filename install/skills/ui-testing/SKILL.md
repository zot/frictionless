---
name: ui-testing
description: use when **testing frictionless apps** in Playwright against requirements
---

# UI Testing Skill

Test frictionless apps in Playwright browser against their requirements.md.

**Goal:** Test and record results only. Do NOT fix bugs during testing — document them in TESTING.md for later resolution.

## Quick Start

Always make sure you have `base_dir` from `ui_status` first. All paths below use `{base_dir}` as a placeholder for this value.
```
1. Find requirements: {base_dir}/apps/<app>/requirements.md
2. Check server: ui_status → get base_dir and url
3. Load app: ui_display("<app>")
4. Open browser: ui_open_browser OR navigate Playwright to {url}/?conserve=true
5. Create TESTING.md from requirements
6. Test each feature, update checklist
7. Document bugs in Known Issues section
```

## Testing Workflow

### 1. Setup

```bash
# Get server info (auto-started)
ui_status  # → base_dir, url

# Load the app
ui_display("<app-name>")

# Open in Playwright (use url from ui_status)
ui_open_browser  # or navigate to {url}/?conserve=true
```

### 2. Create TESTING.md

Create `{base_dir}/apps/<app>/TESTING.md` based on requirements.md:

```markdown
# <App Name> Testing Checklist

## Gaps
(Design/code mismatches from ui_audit - investigate before testing)

## <Feature Category 1>
- [ ] Requirement from requirements.md
- [ ] Another requirement

## <Feature Category 2>
- [ ] More requirements...

## Known Issues
(Bugs found during testing)
```

**Section order:** `## Gaps` must be first (from ui-builder audit), then feature categories, then `## Known Issues` last.

**Checkbox conventions:**
- `[ ]` — untested
- `[✓]` — passed
- `[✗]` — failed

### 3. Test Features

Use Playwright tools to interact with the UI:

| Action | Tool |
|--------|------|
| See current state | `browser_snapshot` |
| Click element | `browser_click(element, ref)` |
| Fill text | `browser_type(element, ref, text)` |
| Fill form | `browser_fill_form(fields)` |
| Check visual | `browser_take_screenshot` |

### 4. Document Results

Update TESTING.md as you test:

```markdown
- [✓] Feature works correctly
- [✗] Feature broken - **BUG: description**
- [ ] Feature not tested yet
```

### 5. Document Bugs

For each bug, document in Known Issues:

```markdown
### N. Bug Title
**Location:** `file.lua:line` or component
**Error:** Error message if any
**Steps to reproduce:**
1. Step one
2. Step two
**Expected:** What should happen
**Actual:** What actually happens
**Root cause:** Analysis of why (if known)
**Impact:** What features are blocked
```

## Playwright Tips

### Getting Element References

Use `browser_snapshot` to get the accessibility tree with refs:
```yaml
- button "+ Add" [ref=e18] [cursor=pointer]
- textbox "Name" [ref=e46]
```

Use the `ref` values in click/type calls.

### Testing Real-time Updates

For features like search filtering, use `slowly: true` in `browser_type` to trigger character-by-character updates:
```
browser_type(element, ref, text, slowly=true)
```

### Checking State vs Visual

- `browser_snapshot` - accessibility tree (state, refs for interaction)
- `browser_take_screenshot` - visual appearance (colors, layout)

Use snapshot for interaction, screenshot to verify visual changes (themes, styling).

## Example TESTING.md

See `{base_dir}/apps/contacts/TESTING.md` for a complete example with:
- Feature checklists organized by category
- Bug documentation with root cause analysis
- Clear marking of blocked features

## Common Bug Patterns

| Symptom | Likely Cause |
|---------|--------------|
| "attempt to index a non-table object" | Method called with wrong args |
| Changes save when Cancel clicked | Form bound to original, not clone |
| Toggle visual but state unchanged | Property binding not two-way |
| Click handler error | Missing method or wrong `self` reference |
