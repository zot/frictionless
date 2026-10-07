---
name: ui-diagnostics
description: Diagnose UI engine problems - variable leaks, stale views, change detection failures, wrapper issues. Load when the UI doesn't update correctly.
---

# UI Diagnostics

Load this skill when the UI isn't behaving correctly: stale data showing, lists not updating, elements not appearing/disappearing, or variables accumulating without cleanup.

## On Skill Load

Read `UNDERSTANDING-UI-ENGINE.md` in the ui-engine project directory for the full architecture. This skill assumes you've internalized those concepts.

Then identify the problem category and follow the appropriate section below.

## Diagnostic Tools

### Variable Browser
Open `/{session-uuid}/variables` in the browser or fetch `/{session-uuid}/variables.json` for machine-readable data. Shows the full variable tree with values, properties, and parent-child relationships.

### Playwright Frontend Inspection
Access frontend state via `uiApp` in the browser console (or Playwright evaluate):

```javascript
// Variable store
uiApp.store.variables.get(varId)        // Get a variable
uiApp.store.variables.size              // Total variable count

// Widgets and Views
uiApp.binding.widgets.get(elementId)    // Get widget for element
widget.views                            // Array of Views on this widget

// Find a ViewList by walking the view tree
view.viewLists                          // ViewList instances on this view
view.childViews                         // Child View instances
viewList.itemViews                      // Item Views in a ViewList
viewList.variableId                     // The variable this ViewList watches
```

### Backend Lua State
Use the project's UI command to query Lua directly:
```bash
# Check array lengths
<ui-command> run 'return #myApp.items'

# Inspect object state
<ui-command> run 'local r = {}; for i, item in ipairs(myApp.items) do r[i] = {name=item.name}; end; return r'
```

### Server Logs
Run the server with `-vvvv` for full diagnostic output. Key log patterns:
- `[IN] CREATE/UPDATE/DESTROY` — protocol messages from frontend
- `AfterBatch: variable N changed` — detected changes
- `ViewList` — wrapper creation and sync activity

## Problem: List Shows Wrong Number of Items

**Symptom**: A `ui-view="items?wrapper=lua.ViewList"` shows more or fewer items than the backend has.

**Diagnosis chain**:

1. **Compare Lua vs backend vs frontend**:
   - Lua: `<ui-command> run 'return #myObject.items'`
   - Backend: Check `variables.json` for the ViewList's `items` child variable — count the array elements
   - Frontend: `uiApp.store.variables.get(itemsVarId).value.length`

2. **If Lua and backend disagree**: The change tracker isn't detecting the array change. Check:
   - Is `ConvertToValueJSON` producing proper `ObjectRef` values or nils? (See "Value JSON shows all nils" below)
   - Is the wrapper's `Update` being called? Check for ViewList log messages at v4

3. **If backend and frontend disagree**: The update isn't reaching the frontend. Check:
   - Is the variable active? (`active: true` in variables.json)
   - Is there a watcher? The variable must be watched for updates to be sent
   - Check `AfterBatch` log — is the variable listed as changed?

4. **If frontend store is correct but DOM is wrong**: The frontend ViewList isn't syncing. Check:
   - `viewList.itemViews.length` vs store value length
   - Whether the ViewList's `update()` method is being called (watcher may be missing)

## Problem: Value JSON Shows All Nils

**Symptom**: In diagnostics, a wrapped array variable shows `valueJSON=[<nil> <nil> ...]` instead of `[{obj:N} {obj:M} ...]`.

**Root cause**: `ConvertToValueJSON` in the resolver must NOT call `tracker.ToValueJSON()` on array elements. The outer `ToValueJSON` already iterates the returned slice and processes each element. If `ConvertToValueJSON` calls `ToValueJSON` on elements, `ObjectRef` structs get double-processed — the struct handler returns nil for structs.

**Fix**: `ConvertToValueJSON` should convert the container (Lua table to Go slice) but return raw Go values as elements. Let the caller handle element conversion.

```go
// WRONG — double-processing
result[i-1] = tracker.ToValueJSON(r.luaElementToGo(elem))

// RIGHT — let caller handle elements
result[i-1] = r.luaElementToGo(elem)
```

## Problem: Variables Accumulating (Memory Leak)

**Symptom**: `uiApp.store.variables.size` grows over time, never shrinks. Clicking through UI states creates variables that are never cleaned up.

**Diagnosis**:

1. **Check if destroy messages flow**: Monkey-patch in Playwright:
   ```javascript
   const orig = uiApp.store.destroy.bind(uiApp.store);
   uiApp.store.destroy = function(varId) {
     console.log('DESTROY sent:', varId);
     return orig(varId);
   };
   ```

2. **Check if destroy confirmations return**: The backend must send a `destroy` message back for each destroyed variable (including descendants, children before parents). If destroys are sent but no confirmations return, check the protocol handler's `handleDestroy`.

3. **Check backend variable count**: Compare `variables.json` entry count with frontend `store.variables.size`. If backend also accumulates, the destroy isn't reaching the backend.

## Problem: Wrapper Update Not Firing

**Symptom**: A wrapped variable's value changes in Lua but the wrapper's `Update` method never runs.

**How wrappers update**:
1. `DetectChanges` calls `checkSingleVariable` for the wrapped variable
2. `GetValue()` re-resolves the path, getting the current value
3. `ToValueJSON(currentValue)` produces the current JSON
4. If `ValueJSON` differs from cached: `updateWrapper()` is called
5. `updateWrapper` calls `CreateWrapper` which calls `NewViewList` → `Update`

**Common causes**:
- **Same object identity**: If the Lua table pointer hasn't changed (in-place mutation), but elements inside it changed, the change must be detected through the VALUE JSON (element-by-element comparison), not object identity
- **Nil ValueJSON**: If all elements serialize to nil (see "Value JSON Shows All Nils"), every state looks identical
- **Variable not active**: Inactive variables are skipped during change detection
- **Priority ordering**: Parent must be checked before child. If parent's change isn't detected, children won't see updated values

## Problem: Stale DOM Elements After Re-render

**Symptom**: Old elements remain visible after a view re-renders, or duplicate elements appear.

**Check**:
- No-flash buffering: Old elements should have `ui-obsolete-view` class, new elements `ui-new-view`. After 100ms timer, obsolete are removed, new are revealed.
- **Duplicate IDs**: If obsolete elements keep their `id` attributes, `getElementById` may return the wrong element. Fix: remove `id` from obsolete elements.
- **Buffer root detection**: `parent.closest('.ui-new-view')` determines if we're inside an ancestor's buffer. If this check fails, elements may be incorrectly managed.

## Variable Path Notation

For diagnostics, the `DiagPath()` method on change-tracker Variables produces paths like:
```
1/value/currentView()/columns/items/0/item/items
```

This shows the full navigation chain from root (variable 1) through each path segment. Useful for identifying which variable in a deep tree is misbehaving.

## The Variable Tree

Understanding the tree structure for a ViewList:

```
Variable: columns (wrapper=lua.ViewList)
  └── items (child of wrapper, resolves ViewList.Items)
       ├── 0 (ViewListItem, created by frontend ViewList)
       │    └── item (child View, resolves to domain object or presenter)
       │         └── [domain object's children...]
       ├── 1 (ViewListItem)
       │    └── item
       ...
```

Key insight: the `items` variable is a child of the **wrapper variable** (not the wrapper object). Its value comes from resolving `items` on the ViewList's `NavigationValue()` (the `*ViewList` Go object), which returns `vl.Items` (the `[]*ViewListItem` slice). This is a Go slice, so `ToValueJSON` processes it directly — no Lua table conversion needed.

The domain objects inside each ViewListItem ARE Lua tables, so those go through `ConvertToValueJSON` when their child variables resolve paths.
