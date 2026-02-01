---
description: Send DOM/browser information to Lua using a hidden input with ui-value binding. JS sets the input value and dispatches an input event, which updates the bound Lua property.
---

# JS-to-Lua Data Bridge

Pass data from JavaScript (browser state, localStorage, DOM measurements) to Lua.

## When to Use

- Reading localStorage values on page load
- Passing browser-only state (current theme, window size, etc.) to Lua
- Any case where JS needs to update Lua state

## Pattern

### 1. Lua Property

Define a property to receive the JS value:

```lua
local MyApp = session:prototype("MyApp", {
    _browserValue = ""  -- Set by JS via hidden input
})

function MyApp:onBrowserValueChanged()
    -- Called when _browserValue updates
    if self._browserValue ~= "" then
        -- Use the value
    end
end
```

### 2. Hidden Input in Viewdef

Create a hidden input bound to the Lua property:

```html
<!-- Hidden input for JS→Lua bridge -->
<input type="hidden" class="my-bridge" ui-value="_browserValue">

<!-- Optional: trigger a method when value changes -->
<span style="display:none" ui-value="onBrowserValueChanged()"></span>
```

The `ui-value` binding on an input is **bidirectional** - changes from either side propagate.

### 3. JavaScript Sets the Value

JS finds the input, sets its value, and dispatches an input event:

```html
<script>
(function() {
    const bridge = document.querySelector('.my-bridge');
    if (bridge) {
        // Read browser state
        const value = localStorage.getItem('myKey') || 'default';

        // Push to Lua
        bridge.value = value;
        bridge.dispatchEvent(new Event('input', { bubbles: true }));
    }
})();
</script>
```

## Complete Example: Theme Sync

Read the current theme from localStorage and sync to Lua on page load.

**Lua:**
```lua
local Prefs = session:prototype("Prefs", {
    _currentTheme = "lcars",
    _browserTheme = ""  -- Set by JS on load
})

function Prefs:syncFromBrowser()
    if self._browserTheme ~= "" then
        self._currentTheme = self._browserTheme
    end
end
```

**Viewdef:**
```html
<template>
  <div class="prefs-panel">
    <!-- UI content here -->
  </div>

  <!-- Hidden input for JS→Lua theme sync -->
  <input type="hidden" class="browser-theme-bridge" ui-value="_browserTheme">
  <span style="display:none" ui-value="syncFromBrowser()"></span>

  <script>
    (function() {
      const bridge = document.querySelector('.browser-theme-bridge');
      if (bridge) {
        const theme = localStorage.getItem('theme') || 'lcars';
        bridge.value = theme;
        bridge.dispatchEvent(new Event('input', { bubbles: true }));
      }
    })();
  </script>
</template>
```

## Key Points

- **Bidirectional binding**: `ui-value` on inputs works both ways
- **Event required**: Must dispatch `input` event for ui-engine to detect the change
- **Hidden input**: Use `type="hidden"` to keep the bridge element invisible
- **Trigger method**: Use a hidden span with `ui-value="methodName()"` to react to changes
- **One-time or continuous**: Can run on load (IIFE) or set up observers for ongoing updates

## Comparison with Lua-to-JS Bridge

| Direction | Pattern | Binding | Trigger |
|-----------|---------|---------|---------|
| Lua → JS | Hidden span with `ui-value` | Read-only | JS polls DOM |
| JS → Lua | Hidden input with `ui-value` | Bidirectional | JS dispatches `input` event |

See also: `lua-js-data-bridge.md` for the Lua→JS direction.
