---
description: Pass Lua state to client-side JS via hidden spans with ui-value bindings. JS polls the DOM for responsive local updates without server round-trips. Use for timers, animations, or any case where JS needs to act on server state more frequently than Lua can push.
---

# Lua-to-JS Data Bridge

Pass data from Lua to client-side JavaScript without round-trips, enabling JS to act on server state locally.

## When to Use

- JS needs to read Lua state (timestamps, counters, configuration)
- JS needs to update UI more frequently than Lua can push (e.g., timers, animations)
- Avoiding server round-trips for responsive client behavior

## Pattern

### 1. Lua Method

Return the data you want JS to access:

```lua
function mcp:waitStartOffset()
    local wt = self:waitTime()
    if wt == 0 then return 0 end
    return math.floor(os.time() - wt)
end
```

### 2. Hidden Span in Viewdef

Bind the Lua method to a hidden span's `ui-value`:

```html
<span class="my-data-bridge" style="display:none;" ui-value="waitStartOffset()"></span>
```

The span updates automatically when the Lua value changes.

### 3. JavaScript Reads from DOM

JS reads the value from the span's text content:

```html
<script>
(function() {
  function update() {
    const dataEl = document.querySelector('.my-data-bridge');
    if (!dataEl) return;

    const value = parseInt(dataEl.textContent, 10);
    // Use the value...
  }

  // Poll for changes
  setInterval(update, 1000);
})();
</script>
```

## Complete Example: Wait Time Counter

Shows elapsed seconds since server wait started, updating every second client-side.

**Lua:**
```lua
function mcp:waitStartOffset()
    local wt = self:waitTime()
    if wt == 0 then return 0 end
    return math.floor(os.time() - wt)
end
```

**Viewdef:**
```html
<template>
  <script>
    (function() {
      function update() {
        const timestampEl = document.querySelector('.wait-timestamp');
        const counterEl = document.querySelector('.wait-counter');
        if (!timestampEl || !counterEl) return;

        const startOffset = parseInt(timestampEl.textContent, 10);
        if (!startOffset) {
          counterEl.classList.add('hidden');
          return;
        }

        const elapsed = Math.floor(Date.now() / 1000) - startOffset;
        if (elapsed > 5) {
          counterEl.textContent = elapsed;
          counterEl.classList.remove('hidden');
        } else {
          counterEl.classList.add('hidden');
        }
      }

      setInterval(update, 1000);
    })();
  </script>

  <!-- Data bridge: Lua timestamp to JS -->
  <span class="wait-timestamp" style="display:none;" ui-value="waitStartOffset()"></span>

  <!-- Display element updated by JS -->
  <span class="wait-counter hidden"></span>
</template>
```

## Key Points

- **One-way flow**: Lua → DOM → JS (JS reads, doesn't write back)
- **Automatic updates**: `ui-value` binding keeps the hidden span in sync with Lua
- **No polling Lua**: JS polls the DOM, not the server
- **Use UNIX timestamps**: For time-based data, use seconds since epoch to avoid timezone issues
