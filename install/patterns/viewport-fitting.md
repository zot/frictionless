---
name: Viewport Fitting
description: Make app fill viewport with scrollable content area
---

# Viewport Fitting

Make an app fill the entire viewport with a scrollable content area.

## CSS Implementation

```css
html, body {
  margin: 0;
  padding: 0;
  overflow: hidden;
}

.my-app {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

.scrollable-area {
  flex: 1;
  min-height: 0;  /* CRITICAL */
  overflow-y: auto;
}
```

## Key Points

- `height: 100vh` on the app container fills the viewport
- `overflow: hidden` on body prevents double scrollbars
- **CRITICAL:** `min-height: 0` on flex children allows them to shrink below content size
- `flex: 1` makes the scrollable area take remaining space
- `overflow-y: auto` enables scrolling only when needed

## Common Mistake

Forgetting `min-height: 0` causes the flex child to expand to fit its content, breaking the scroll behavior.
