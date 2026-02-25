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

# Finding elements and views from variables

Every variable with a DOM binding has an `elementId` property — the id of the element it's bound to (e.g. `"ui-42"`).

To find which **viewdef** a variable belongs to, look for a `viewdef` property on the variable itself. Only view-level variables have one (e.g. `"JobTracker.DEFAULT"`). For child variables, walk up via `parentId` until you find an ancestor with a `viewdef` property.

**Example:** find the element for a slow-computing variable:
1. Fetch variables with `.ui/mcp variables`
2. Sort/filter by `computeTime` or `maxComputeTime`
3. Read the target's `elementId` — that's the DOM element
4. Walk `parentId` chain to find the enclosing `viewdef`

Once you have an `elementId`, you can highlight it for the user — see `/ui-highlight`.
