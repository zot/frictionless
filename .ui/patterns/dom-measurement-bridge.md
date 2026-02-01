---
name: DOM Measurement Bridge
description: Let JavaScript measure rendered positions from Lua-generated content, then create UI elements based on actual pixel positions
---

# DOM Measurement Bridge

Pattern for JavaScript to access positioning data by measuring Lua-generated DOM elements.

## Problem

JavaScript needs pixel-accurate positioning data (scroll metrics, element bounds) that only the browser can calculate, but Lua generates the content. Direct Lua→JS data passing isn't available in ui-engine.

## Solution

1. **Lua marks elements** with classes during HTML generation
2. **JS finds elements** by querying for those classes
3. **JS measures positions** using browser APIs (`offsetTop`, `offsetHeight`, etc.)
4. **JS creates/updates UI** based on measurements

## Example: Scrollbar Trough Markers

Lua wraps warning lines with class-marked spans:
```lua
escaped = '<span class="pushstate-highlight-line">' .. escaped .. '</span>'
```

JS finds spans, measures positions, creates markers:
```javascript
const warningSpans = pre.querySelectorAll('.pushstate-highlight-line, .osexecute-highlight-line');

warningSpans.forEach(span => {
  const offsetTop = span.offsetTop;
  const height = span.offsetHeight;

  // Calculate scroll-aligned position (aligns with scrollbar thumb)
  const topPercent = Math.min((offsetTop / maxScroll) * 100, 100);
  const heightPercent = (height / maxScroll) * 100;

  // Create marker element
  const marker = document.createElement('div');
  marker.className = 'trough-marker trough-marker-warning';
  marker.style.top = topPercent.toFixed(2) + '%';
  marker.style.height = Math.max(heightPercent, 0.3).toFixed(2) + '%';
  trough.appendChild(marker);
});
```

## Triggering Updates

Use MutationObserver to detect when Lua updates content:
```javascript
const wrapper = document.querySelector('.content-wrapper');
if (wrapper) {
  const observer = new MutationObserver(() => setTimeout(positionMarkers, 50));
  observer.observe(wrapper, { childList: true, subtree: true });
}
```

## Scroll Position vs Document Position

For elements that should align with the scrollbar thumb:
- **Document position**: `offsetTop / scrollHeight` — where in the document
- **Scroll position**: `offsetTop / (scrollHeight - clientHeight)` — aligns with thumb

Use scroll position when markers should match thumb position when content is at viewport top.

## When to Use

- Minimap/overview indicators needing pixel-accurate positioning
- Features depending on rendered layout (text wrap, overflow)
- Scroll position correlation (trough markers, progress indicators)
- Any UI that must reflect actual rendered positions, not line counts
