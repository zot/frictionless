---
name: Scrollable View List
description: Make a ui-view list scrollable with max-height constraint
---

# Scrollable View List

When using `ui-view` with `wrapper=lua.ViewList`, the binding creates a `.view-list` wrapper element. To make this scrollable, target the `.view-list` class directly rather than adding a wrapper div.

## Problem

The `ui-view` binding replaces its host element with a `.view-list` container:

```html
<!-- This div gets replaced -->
<div class="my-messages" ui-view="messages?wrapper=lua.ViewList"></div>

<!-- Becomes this in the DOM -->
<div class="view-list" id="ui-XXX">
  <!-- list items here -->
</div>
```

CSS targeting `.my-messages` won't work because that element no longer exists.

## Solution

Target `.view-list` as a child of the parent container:

```html
<div class="chat-output">
  <div ui-view="messages?wrapper=lua.ViewList&scrollOnOutput"></div>
</div>
```

```css
.chat-output > .view-list {
  max-height: 200px;
  overflow-y: auto;
}
```

## With Auto-Scroll

Add `scrollOnOutput` to the binding path to auto-scroll when new items are added:

```html
<div ui-view="messages?wrapper=lua.ViewList&scrollOnOutput"></div>
```

## Example from app-console

```css
.chat-lua-column > .view-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  background: var(--term-bg);
  border: 1px solid var(--term-border);
}
```
