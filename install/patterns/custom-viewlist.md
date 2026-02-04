---
name: Custom ViewLists
description: Styling default ViewLists and creating custom viewdefs with namespace pattern
---

# Custom ViewList Patterns

Two approaches for customizing ViewList behavior: styling the default ViewList, or creating custom viewdefs with namespaces.

## Approach 1: Styling Default ViewLists

When using `ui-view` with `wrapper=lua.ViewList`, the binding creates a `.view-list` wrapper element. Target this class to style it.

### The Problem

The `ui-view` binding replaces its host element:

```html
<!-- This div gets replaced -->
<div class="my-messages" ui-view="messages?wrapper=lua.ViewList"></div>

<!-- Becomes this in the DOM -->
<div class="view-list" id="ui-XXX">
  <!-- list items here -->
</div>
```

CSS targeting `.my-messages` won't work because that element no longer exists.

### Solution: Target .view-list

Wrap in a parent and target `.view-list` as a child:

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

### Auto-Scroll

Add `scrollOnOutput` to auto-scroll when new items are added:

```html
<div ui-view="messages?wrapper=lua.ViewList&scrollOnOutput"></div>
```

### Example: Flexbox Scrollable List

```css
.chat-lua-column > .view-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  background: var(--term-bg);
  border: 1px solid var(--term-border);
}
```

---

## Approach 2: Custom Viewdefs with Namespaces

When you need different HTML structure (e.g., `<span>` instead of `<div>`), create custom viewdefs using the namespace pattern.

### The Problem

The default `lua.ViewList` and `lua.ViewListItem` viewdefs use `<div>` containers. Sometimes you need inline elements like `<span>` for proper layout (e.g., buttons in a horizontal scrolling container).

### Solution: Namespace Suffix Pattern

1. **Create custom viewdefs with namespace suffix**:

   ```
   .ui/apps/<app>/viewdefs/lua.ViewList.github.html
   .ui/apps/<app>/viewdefs/lua.ViewListItem.github.html
   .ui/apps/<app>/viewdefs/AppConsole.GitHubTab.github.html
   ```

2. **ViewList viewdef** (`lua.ViewList.github.html`):
   ```html
   <template>
     <span ui-viewlist="items" style="white-space: nowrap"></span>
   </template>
   ```

   **Important**: Add `white-space: nowrap` here to prevent child elements from wrapping. This is the container that controls layout flow.

3. **ViewListItem viewdef** (`lua.ViewListItem.github.html`):
   ```html
   <template>
     <span ui-view="item"></span>
   </template>
   ```

4. **Item viewdef** (`AppConsole.GitHubTab.github.html`):
   ```html
   <template>
     <sl-button size="small" ui-attr-variant="buttonVariant()" ui-action="selectMe()">
       <span ui-value="filename"></span>
     </sl-button>
   </template>
   ```

5. **Use in parent viewdef** with `ui-namespace` attribute:
   ```html
   <span style="display: contents"
         ui-view="githubTabs?wrapper=lua.ViewList"
         ui-namespace="github"></span>
   ```

6. **Parent container CSS** for horizontal scrolling:
   ```css
   .github-tabs {
     display: flex;
     overflow-x: auto;  /* Horizontal scroll when content overflows */
   }
   ```

7. **Link the viewdefs** so Frictionless can find them:
   ```bash
   .ui/mcp linkapp add <app>
   ```

### Key Points

- Namespace suffix pattern: `{Prototype}.{namespace}.html`
- The `ui-namespace="github"` attribute applies to the entire subtree
- Must run `linkapp add` after creating new viewdefs in app directories
- All three viewdefs (ViewList, ViewListItem, and the item prototype) need the namespace suffix
- **Put `white-space: nowrap` on the ViewList span**, not on child elements - the ViewList is the layout container
- Parent container needs `overflow-x: auto` for the scrollbar to appear
