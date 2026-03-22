---
name: Shell Child Wrapper
description: Fix layout collapse when app has multiple direct child divs
---

# Shell Child Wrapper

The MCP shell forces `height: 100% !important` and `max-height: 100% !important` on all direct child divs of `.mcp-app-container` (see `apps/mcp/css/shell.css`). This works fine when the app has a single root div, but **breaks flex layouts with multiple children** — each child gets `height: 100%`, causing some to collapse to 0px.

## Symptom

App content is invisible. The status bar (or other secondary div) expands to fill the entire container while the main layout collapses.

## Fix

Wrap all children in a single intermediate div:

```html
<template>
  <div class="my-app">
    <div class="my-app-inner">
      <div class="my-layout">...</div>
      <div class="my-status-bar">...</div>
    </div>
  </div>
</template>
```

```css
.my-app-inner {
  display: flex;
  flex-direction: column;
  height: 100%;
}
```

The shell's rule hits `.my-app-inner` (the only direct child), giving it `height: 100%` — which is what we want. Inside it, normal flex layout works without interference.
