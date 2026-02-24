---
name: ui-variables
description: understand ui variables for performance tuning and diagnosing problems; how long it took to compute their values, where they are bound in the DOM
---

Be sure you have loaded /ui-basics.

# Getting the current variables

`.ui/mcp variables` produces the current variables as a JSON array of variable objects.

Variable object properties and their meanings:
- id: the variable's id
- parentId: 570,
- goType: the Go type of the actual value
- path: the variable's server-side path from the parent
- value: the value of the variable, which could be a wrapper on the baseValue. This is what the UI uses
- type: The type of the value
- baseValue: the actual value of the variable. Different from value only if there is a wrapper
- childIds: an array of the ids of this variable's children
- computeTime: time it took to compute value, with units as a string, like "610ns"
- maxComputeTime: the maximum value of this variable's computeTime since its creation
- active: whether the variable is actively checking for changes
- access: access type
- changeCount: how many times this variable has changed since creation
- depth: how far down in the variable tree this variable is
- elementId: the id of the DOM element that binds this variable
- viewdef: the viewdef this variable represents (e.g. "JobTracker.DEFAULT"). Only present on view-level variables — for children, walk up the variable hierarchy via parentId to find the nearest ancestor with a viewdef property.
- properties: a map of the variable's properties (path, type, elementId, viewdef, and access come from here)

# Human use of variables
The variable browser in the MCP shell (toggled by the `{}` icon) shows variables in a sortable table with a Viewdef column. Clicking any cell:
- Highlights the **view** with a translucent overlay (matching the element's border-radius from the current theme)
- Highlights the **element** with a box-shadow ring
- Draws a **dashed connector line** from the click point to the target center
- Shows a **toast** with the resolved viewdef name
- Copies to clipboard (viewdef name for the Viewdef column, full variable JSON for others)

If the element is hidden but its view is visible, only the view overlay shows. If both are hidden, the toast shows "(not visible)". All highlights fade after 3 seconds (1s solid + 2s fade).

# Finding the viewdef for a variable

View-level variables have a `viewdef` property directly (e.g. "JobTracker.DEFAULT"). For child variables, walk up via `parentId` until you find an ancestor with a `viewdef` property.

# Agent-driven diagnostics

You can draw the user's attention to any element in the UI via `mcp.code`. This replicates the variable browser's highlight technique and works even when the panel is closed.

**Important:** `mcp.code` uses change detection — re-assigning the same value is a no-op. Append a nonce (e.g. `.. os.time()`) to force re-execution.

**Mouse position:** There is no synchronous Web API to query mouse coordinates — they must be captured from events. Install a persistent tracker early (ideally at session start), then read `window._mousePos` when highlighting. The user's normal interaction (typing in chat, clicking) keeps it current.

**Step 1: Install mouse tracker** (once per session, before any highlighting):
```lua
mcp.code = [=[
(function() {
  if (!window._mousePos) {
    window._mousePos = {x: 0, y: 0};
    document.addEventListener("mousemove", function(e) {
      window._mousePos.x = e.clientX;
      window._mousePos.y = e.clientY;
    });
  }
})()
// setup: ]=] .. os.time()
```

**Step 2: Highlight with overlay, ring, and connector line:**
```lua
-- Highlight element "ui-42" with a line from the user's mouse position
mcp.code = [=[
(function() {
  var mp = window._mousePos || {x: 0, y: 0};
  var el = document.getElementById('ui-42');
  if (!el || !el.getBoundingClientRect().width) return;
  var r = el.getBoundingClientRect();
  var br = getComputedStyle(el).borderRadius;
  // Box-shadow ring (follows border-radius, unlike outline)
  el.style.boxShadow = '0 0 0 3px #E07A47, 0 0 12px rgba(224,122,71,0.6)';
  el.style.transition = 'box-shadow 2s ease-out';
  setTimeout(function() { el.style.boxShadow = '0 0 0 3px transparent'; }, 2000);
  setTimeout(function() { el.style.boxShadow = ''; el.style.transition = ''; }, 4000);
  // Translucent overlay
  var ov = document.createElement('div');
  ov.style.cssText = 'position:fixed;pointer-events:none;z-index:9998;background:rgba(224,122,71,0.15);border:2px solid rgba(224,122,71,0.5);border-radius:' + br + ';transition:opacity 2s ease-out;';
  ov.style.left = r.left + 'px'; ov.style.top = r.top + 'px';
  ov.style.width = r.width + 'px'; ov.style.height = r.height + 'px';
  document.body.appendChild(ov);
  setTimeout(function() { ov.style.opacity = '0'; }, 1000);
  setTimeout(function() { ov.remove(); }, 3000);
  // SVG dashed connector from mouse to element center
  var cx = r.left + r.width / 2, cy = r.top + r.height / 2;
  var svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
  svg.style.cssText = 'position:fixed;top:0;left:0;width:100vw;height:100vh;pointer-events:none;z-index:9999;';
  var line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
  line.setAttribute('x1', mp.x); line.setAttribute('y1', mp.y);
  line.setAttribute('x2', cx); line.setAttribute('y2', cy);
  line.setAttribute('stroke', '#E07A47'); line.setAttribute('stroke-width', '2');
  line.setAttribute('stroke-dasharray', '6 4');
  var dot = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
  dot.setAttribute('cx', cx); dot.setAttribute('cy', cy);
  dot.setAttribute('r', '4'); dot.setAttribute('fill', '#E07A47');
  svg.appendChild(line); svg.appendChild(dot);
  document.body.appendChild(svg);
  setTimeout(function() { svg.style.transition = 'opacity 2s ease-out'; svg.style.opacity = '0'; }, 1000);
  setTimeout(function() { svg.remove(); }, 3000);
})()
// nonce: ]=] .. os.time()
```

Workflow for pointing to a UI element:

1. Install the mouse tracker (step 1) early — ideally when you first start interacting with the UI. It's idempotent and survives until page reload.
2. Fetch variables with `.ui/mcp variables`
3. Find the target variable (by name, type, computeTime, etc.) and note its `elementId`
4. Use `mcp.code` to highlight the element (step 2 above)
5. Use `mcp:addAgentMessage()` to explain what you're pointing at

The connector line draws from the user's last known mouse position to the target, giving a natural "pointing at" effect. Since the user interacts with the chat panel via mouse, the position is typically near the chat area.
