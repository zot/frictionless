# Variable Browser

**Source Spec:** specs/debug-pages.md
**Requirements:** R131, R132, R133

## Overview

Static HTML page (`install/html/variables.html`) that provides an interactive variable inspector. Originally extracted from an embedded Go string in ui-engine, now maintained as a standalone file in frictionless.

Served on the UI port at `/variables` (with session cookie). Fetches JSON from `/{sessionId}/variables.json` (ui-engine route backed by `getDebugVariables()` in `resources.go`).

## Click-to-Copy

Clicking any table cell **except** the Path column copies that row's full variable JSON to the clipboard.

- Path is excluded because it handles tree expand/collapse click interaction
- Uses `navigator.clipboard.writeText()` with the variable's full JSON object
- Shows a brief CSS toast notification ("Copied") that auto-fades

### Implementation

- Each data row stores its variable JSON object during render
- Click handler on `<tbody>` delegates to cell, checks column, builds JSON from the row's stored variable, copies to clipboard
- Toast element is a fixed-position div, toggled with a CSS class
