---
description: Send DOM/browser information to Lua using window.uiApp.updateValue() or a hidden input with ui-value binding. The updateValue API is preferred for reliability.
---

# JS-to-Lua Data Bridge

Pass data from JavaScript (browser state, localStorage, DOM measurements) to Lua.

## When to Use

- Reading localStorage values on page load
- Passing browser-only state (current theme, window size, etc.) to Lua
- File uploads (FileReader → base64 → Lua)
- Any case where JS needs to update Lua state

## Recommended: updateValue API

Use `window.uiApp.updateValue(elementId, value?)` - the most reliable approach.

### 1. Lua Property

```lua
local MyApp = session:prototype("MyApp", {
    browserValue = ""  -- Set by JS
})

function MyApp:processBrowserValue()
    if self.browserValue ~= "" then
        -- Use the value
        self.browserValue = ""  -- Clear after processing
    end
end
```

### 2. Hidden Span with ID in Viewdef

```html
<!-- Hidden span for JS→Lua bridge - MUST have an id -->
<span id="my-bridge" style="display:none" ui-value="browserValue?access=rw"></span>
<!-- Trigger processing when value changes - MUST use priority=low -->
<span style="display:none" ui-value="processBrowserValue()?priority=low"></span>
```

**Critical:** The trigger method MUST use `?priority=low` to ensure it fires AFTER the input value has been set. Without this, you'll get race conditions where the trigger fires before the data is available.

The `?access=rw` makes the span writable (spans are read-only by default).

### 3. JavaScript Uses updateValue

```html
<script>
(function() {
    const value = localStorage.getItem('myKey') || 'default';
    // Use updateValue API - waits for uiApp to be ready
    const tryUpdate = () => {
        if (window.uiApp) {
            window.uiApp.updateValue('my-bridge', value);
        } else {
            setTimeout(tryUpdate, 50);
        }
    };
    tryUpdate();
})();
</script>
```

## Complete Example: File Upload Bridge

Handle file uploads by reading files as base64 and sending to Lua.

**Lua:**
```lua
local MyApp = session:prototype("MyApp", {
    fileUploadData = ""  -- filename:base64content
})

function MyApp:processFileUpload()
    if self.fileUploadData == "" then return end

    local colonPos = self.fileUploadData:find(":")
    if not colonPos then
        self.fileUploadData = ""
        return
    end

    local filename = self.fileUploadData:sub(1, colonPos - 1)
    local base64 = self.fileUploadData:sub(colonPos + 1)
    self.fileUploadData = ""  -- Clear immediately

    -- Decode and save file...
end
```

**Viewdef:**
```html
<template>
  <input type="file" class="file-input" multiple
         onchange="window.handleFiles(this.files); this.value='';" />
  <span id="file-bridge" style="display:none" ui-value="fileUploadData?access=rw"></span>
  <!-- CRITICAL: priority=low ensures trigger fires AFTER value is set -->
  <span style="display:none" ui-value="processFileUpload()?priority=low"></span>

  <script>
    window.handleFiles = async function(files) {
        for (const file of files) {
            const reader = new FileReader();
            reader.onload = function(e) {
                const base64 = e.target.result.split(',')[1];
                const payload = file.name + ':' + base64;
                if (window.uiApp) {
                    window.uiApp.updateValue('file-bridge', payload);
                }
            };
            reader.readAsDataURL(file);
        }
    };
  </script>
</template>
```

## Alternative: Event Dispatch (Legacy)

For simple cases, you can still dispatch input events:

```javascript
const bridge = document.querySelector('#my-bridge');
bridge.value = 'new value';
bridge.dispatchEvent(new Event('input', { bubbles: true }));
```

**Note:** This approach may not work reliably with all input types (especially hidden inputs). Prefer `updateValue` for reliability.

## Key Points

- **Use `updateValue` API**: `window.uiApp.updateValue(elementId, value)` is most reliable
- **Element needs ID**: The target element must have an `id` attribute
- **Wait for uiApp**: Check `window.uiApp` exists before calling
- **Trigger with low priority**: Use `ui-value="methodName()?priority=low"` to ensure trigger fires AFTER the value is set
- **Clear after processing**: Set the bridge value to "" after handling to allow repeat updates

## Comparison with Lua-to-JS Bridge

| Direction | Pattern | API |
|-----------|---------|-----|
| Lua → JS | Hidden span with `ui-value` | JS reads `element.textContent` |
| JS → Lua | Input with `ui-value` + ID | `window.uiApp.updateValue(id, value)` |

See also: `lua-js-data-bridge.md` for the Lua→JS direction.
