---
description: How JavaScript works inside ui-engine viewdefs. Covers script execution, element lookup, timers, animations, and the critical gotchas (no getElementById, no element capture in closures). Read this before writing any JS in viewdefs.
---

# Viewdef JavaScript Guide

How to write JavaScript inside ui-engine `<template>` viewdefs.

## Critical Rules

### 1. No `getElementById` — ui-engine overrides IDs

ui-engine assigns its own `id` attributes (`ui-1`, `ui-2`, etc.) to elements. Any `id` you set in HTML is replaced. **`document.getElementById('my-id')` will never find your element.**

```html
<!-- WRONG: id is overwritten by ui-engine -->
<div id="my-shell" class="my-shell">...</div>
<script>
  document.getElementById('my-shell'); // null!
</script>

<!-- RIGHT: use querySelector with a class -->
<div class="my-shell">...</div>
<script>
  document.querySelector('.my-shell'); // works
</script>
```

**Exception:** IDs on elements used with `window.uiApp.updateValue()` (the JS-to-Lua bridge) DO work because `updateValue` uses ui-engine's internal ID mapping. But you still can't use `getElementById` to find those elements.

### 2. Never capture DOM elements in closures

DOM elements captured in closures create uncollectable garbage when the viewdef is hot-reloaded (the old element is gone but the closure still references it). Always look up elements fresh.

```javascript
// WRONG: captures element in closure
(function() {
  const shell = document.querySelector('.my-shell');
  setInterval(function() {
    shell.querySelector('.counter').textContent = '...'; // stale reference after hot-reload
  }, 1000);
})();

// RIGHT: look up fresh each time
(function() {
  setInterval(function() {
    var shell = document.querySelector('.my-shell');
    if (!shell) return;
    shell.querySelector('.counter').textContent = '...';
  }, 1000);
})();
```

### 3. Prefer data bridge over `ui-code`

`ui-code` executes JS when a Lua property changes. It works but has drawbacks:
- Race conditions on page load (code fires before scripts run)
- Requires counter suffixes to force re-evaluation
- Tightly couples Lua to JS execution

**Instead, use the Lua-to-JS data bridge**: Lua writes state to a hidden span via `ui-value`, JS polls the span with `setInterval`. This is simpler, survives hot-reloads, and has no race conditions.

## Script Placement

Scripts go inside `<template>` but **after** the HTML content they reference:

```html
<template>
  <style>/* CSS here */</style>

  <div class="my-app">
    <!-- App HTML here -->
    <span class="my-bridge" style="display:none" ui-value="myState()"></span>
  </div>

  <!-- Scripts AFTER the HTML they reference -->
  <script>
    (function() {
      // ...
    })();
  </script>
</template>
```

Scripts execute when the viewdef is rendered into the DOM. On hot-reload, the template is re-rendered and scripts re-execute.

## Backend State Persists Across Page Reloads

**The Lua session survives browser page reloads.** When the user reloads the page:

1. The browser gets a fresh JS environment (all scripts re-execute, all closures reset)
2. A new websocket connects to the **same Lua session** — all Lua state is intact
3. All `ui-value`, `ui-class-*`, `ui-attr-*` bindings re-render from current Lua values
4. All `ui-code` bindings re-fire with their current Lua property values

This means JS-side state (variables, timers, closures) resets, but Lua-side state does not. Design accordingly:

- **Timers and animations** must handle restarting cleanly (the `setInterval` pattern does this naturally)
- **`ui-code` properties** with stale values will re-fire (see "`ui-code` Re-fires on Page Reload" below)
- **`session.reloading`** is only true during hot-reload (code file changes), NOT on page reload

## Pattern: State-Driven JS Animation

Use the Lua-to-JS data bridge to drive client-side animations from Lua state.

### 1. Lua bridge method

Return a simple flag that JS can poll:

```lua
function MyApp:animationActive()
    return self.showAnimation and "1" or "0"
end
```

### 2. Hidden span in viewdef

```html
<span class="animation-bridge" style="display:none" ui-value="animationActive()"></span>
```

### 3. JS polls and animates

```html
<script>
(function() {
  var wasActive = false;
  var startTime = 0;

  function update() {
    var bridge = document.querySelector('.animation-bridge');
    if (!bridge) return;

    var active = bridge.textContent === '1';

    // Detect activation edge
    if (active && !wasActive) {
      startTime = Date.now();
    }
    wasActive = active;

    if (!active) {
      // Clean up animation state
      document.querySelectorAll('.animated-item.active')
        .forEach(function(el) { el.classList.remove('active'); });
      return;
    }

    // Use elapsed time for cycling (e.g., 3 seconds per item)
    var elapsed = Date.now() - startTime;
    var items = document.querySelectorAll('.animated-item');
    var idx = Math.floor(elapsed / 3000) % items.length;

    items.forEach(function(el, i) {
      if (i === idx) el.classList.add('active');
      else el.classList.remove('active');
    });
  }

  setInterval(update, 200);
})();
</script>
```

### Why this works

- **No `ui-code`**: JS reads state from the DOM, no Lua-to-JS execution needed
- **No IDs**: Uses `querySelector` with classes
- **No captured elements**: Looks up fresh on every tick
- **Survives hot-reload**: New script re-creates the interval; new bridge span is found by class
- **Activation edge detection**: `wasActive` tracks transitions so animation resets on re-activation
- **Time-based cycling**: `Date.now()` math instead of counters avoids drift

## Pattern: One-Shot JS Actions

For JS that runs once when a condition becomes true (e.g., scroll into view, focus an input):

```lua
function MyApp:shouldFocusInput()
    return self._needsFocus and "1" or "0"
end

function MyApp:clearFocus()
    self._needsFocus = false
end
```

```html
<span class="focus-bridge" style="display:none" ui-value="shouldFocusInput()"></span>
<span style="display:none" ui-value="clearFocus()?priority=high"></span>

<script>
(function() {
  var lastState = '0';
  setInterval(function() {
    var bridge = document.querySelector('.focus-bridge');
    if (!bridge) return;
    var state = bridge.textContent;
    if (state === '1' && lastState !== '1') {
      var input = document.querySelector('.my-input');
      if (input) input.focus();
    }
    lastState = state;
  }, 200);
})();
</script>
```

## `ui-code` Re-fires on Page Reload

**Critical behavior:** `ui-code` bindings fire once per websocket session when the variable value is first sent to the browser. On page reload, a new websocket session starts, so `ui-code` re-fires with whatever value is stored in the Lua property — even if that value was set long ago.

This means **stale `ui-code` values have side effects on every page reload**. If a `ui-code` property contains `"doSomething()"` and the user reloads the page, `doSomething()` runs again.

**Guard pattern:** If the JS action should only run when a feature is active, check a visibility flag:

```javascript
window.myAction = function(selector, step) {
  var container = document.querySelector('.my-overlay');
  if (!container || container.classList.contains('hidden')) return;
  // ... safe to proceed
};
```

**Clear after use:** Set the property to `""` when the action is no longer needed:

```lua
function MyApp:finish()
    self.active = false
    self.codeProperty = ""  -- prevents re-fire on page reload
end
```

## When `ui-code` IS appropriate

`ui-code` is still useful for cases where:
- The JS must run with **exact values** from Lua (e.g., computed CSS positions, dynamic selectors)
- The action is **fire-and-forget** with no ongoing state (e.g., scroll to a position)
- There's no activation/deactivation cycle to manage

Example (tutorial card repositioning):
```lua
self.repositionCode = string.format([[
    var el = document.querySelector(%q);
    if (el) el.scrollIntoView({block: 'center'});
    // %d
]], selector, self.counter)
```

```html
<div style="display:none" ui-code="repositionCode"></div>
```

The counter suffix forces re-evaluation even if the selector is the same.

## Summary

| Need | Approach |
|------|----------|
| Find elements | `document.querySelector('.class')` — never `getElementById` |
| Ongoing animation/timer | Lua-to-JS data bridge (hidden span + polling) |
| One-shot action from Lua | Data bridge with edge detection, or `ui-code` |
| Fire-and-forget with Lua values | `ui-code` with counter suffix |
| DOM element reference | Look up fresh each time — never capture in closure |
