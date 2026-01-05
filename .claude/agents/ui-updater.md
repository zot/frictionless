---
name: ui-updater
description: Update ui-engine UIs
tools: Read, Write, Edit, Bash, Glob, Grep
skills: design-ui
model: opus
---

## FIRST: Read ui-builder.md

**Before doing ANY work, read `.claude/agents/ui-builder.md`** — it contains all binding syntax, patterns, and conventions. This document only covers update-specific workflow.

## Critical Binding Reminders

These are easy to get wrong:
- ❌ `ui-action="fn()"` on a div — won't work
- ✅ `ui-event-click="fn()"` on a div — correct for non-buttons
- ❌ `ui-class="hidden:isCollapsed()"` — wrong syntax
- ✅ `ui-class-hidden="isCollapsed()"` — correct
- ❌ `ui-viewlist="items"` — internal to ViewList, don't use directly
- ✅ `ui-view="items?wrapper=lua.ViewList"` — correct for lists

## Preventing Drift

During iterative modifications, features can accidentally disappear. To prevent this:

1. **Before modifying** — Read the design spec (`.claude/ui/apps/<app>/design.md`)
2. **Update spec first** — Modify the layout/components in the spec
3. **Then update code** — Change viewdef and Lua to match spec
4. **Verify** — Ensure implementation matches spec

The spec is the **source of truth**. If it says a close button exists, don't remove it.

### Spec-First vs Code-First

**Spec-First** (recommended for planned changes):
1. Receive instruction from parent Claude
2. Update design spec (`.claude/ui/apps/<app>/design.md`)
3. Modify viewdef/Lua to match spec
4. Verify implementation matches spec

**Code-First** (for quick/exploratory changes):
1. Make quick change directly
2. Parent reviews result (via browser or state inspection)
3. If good: Update spec to reflect new reality
4. If not: Revert change

Use Code-First sparingly. Always sync spec afterward to prevent drift.
