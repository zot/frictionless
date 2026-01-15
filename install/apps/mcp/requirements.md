# MCP

The outer shell for all ui-mcp apps. Displays the current app and provides navigation between apps.

## Architecture

**Uses the global `mcp` object directly.** This app does not create its own prototype — it renders the `mcp` object provided by the server. The `mcp` object has:
- `value` - the currently displayed app (set via `mcp:display(appName)`)
- `code` - JavaScript code to execute (for browser control)

## Layout

The shell is minimal chrome around the current app:
- Full viewport display of `mcp.value` (the current app)
- A 9-dot menu button in the top-right corner for app switching
- A hidden element with `ui-code="code"` for JavaScript execution

## App Switcher Menu

The 9-dot button (grid icon) in the top-right corner:
- Positioned absolutely, floating over app content
- Opens a dropdown/popover menu listing available apps
- Clicking an app name calls `mcp:display(appName)` to switch
- Menu closes after selection

## Processing Indicator

Show a spinner covering the dots button when the agent event loop is not connected to the `/wait` endpoint:
- Indicates that Claude is processing events
- Spinner replaces/overlays the 9-dot icon
- When connected to `/wait`, the normal dots icon is shown

## Available Apps

The menu should list apps discovered from the apps directory. Since mcp is Lua-driven, it should scan for available apps on load using the same filesystem pattern as the `apps` app.

## JavaScript Execution

A hidden element binds to `mcp.code` via `ui-code`:
- When `mcp.code` changes, the JavaScript is executed
- Used for browser control (close window, open URLs, etc.)
- The element is invisible (display: none or similar)

## Events

This app does not send events to Claude. App switching is handled entirely in Lua via `mcp:display()`.

## Styling

- No padding or margins around the app content
- The menu button should have subtle styling (semi-transparent, hover effect)
- Menu button should not interfere with app content interaction
