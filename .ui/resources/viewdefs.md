# Viewdef Syntax Reference

View definitions (viewdefs) are HTML templates that define how Lua objects are rendered. They use `ui-*` attributes to bind UI elements to Lua state.

## Template Structure

```html
<template>
  <div class="my-component">
    <!-- UI bindings here -->
  </div>
</template>
```

## Core Binding Attributes

| Attribute | Description | Example |
|:----------|:------------|:--------|
| `ui-value` | Bind element value/text to Lua path | `<sl-input ui-value="name">` or `<span ui-value="fullName()">` |
| `ui-action` | Bind click to Lua method (buttons only) | `<sl-button ui-action="save()">` |
| `ui-event-*` | Bind any event to Lua method | `<div ui-event-click="toggle()">` |
| `ui-event-keypress-*` | Bind specific key press | `<sl-input ui-event-keypress-enter="submit()">` |
| `ui-view` | Render child object with its viewdef | `<div ui-view="selectedItem">` |
| `ui-attr-*` | Bind HTML attribute to Lua path | `<sl-alert ui-attr-open="hasError">` |
| `ui-class-*` | Toggle CSS class on boolean | `<div ui-class-active="isSelected">` |
| `ui-style-*` | Bind CSS style to Lua path | `<div ui-style-color="themeColor">` |
| `ui-code` | Execute JavaScript when value changes | `<div ui-code="jsCode">` |
| `ui-namespace` | Set viewdef namespace for children | `<div ui-namespace="COMPACT">` |

**Note:** `ui-value` works for both input elements (sets `.value`) and display elements like `<span>` (sets text content).

## Keypress Bindings

`ui-event-keypress-*` fires only when the specified key is pressed:

**Basic keys:**
- `ui-event-keypress-enter` - Enter/Return key
- `ui-event-keypress-escape` - Escape key
- `ui-event-keypress-left/right/up/down` - Arrow keys
- `ui-event-keypress-tab` - Tab key
- `ui-event-keypress-space` - Space bar
- `ui-event-keypress-{letter}` - Any single letter (e.g., `ui-event-keypress-a`)

**With modifier keys:**
- `ui-event-keypress-ctrl-enter` - Ctrl+Enter
- `ui-event-keypress-shift-a` - Shift+A
- `ui-event-keypress-ctrl-shift-s` - Ctrl+Shift+S
- `ui-event-keypress-alt-left` - Alt+Left arrow
- Modifiers: `ctrl`, `shift`, `alt`, `meta` (can be combined in any order before the key)

**Modifier matching is exact:** If modifiers are specified, they must all be pressed and no additional modifiers should be pressed.

## ui-code Binding

Execute JavaScript when a variable's value changes. The code has access to:
- `element` - The bound DOM element
- `value` - The new value from the variable
- `variable` - The variable object
- `store` - The VariableStore

```html
<!-- Close browser when closeWindow becomes truthy -->
<div ui-code="closeWindow" style="display:none;"></div>
```

```lua
-- In Lua: trigger the code
app.closeWindow = "window.close()"
```

Use cases: auto-close window, trigger downloads, custom DOM manipulation, browser APIs.

## Path Syntax

Paths are resolved relative to the current object on the server.

```
property           → self.property
nested.path        → self.nested.path
method()           → self:method()
method(_)          → self:method(value) -- passes update value as arg
items.0            → self.items[1] (Lua is 1-indexed)
..                 → parent object
```

**IMPORTANT:** No operators in paths! `!`, `==`, `&&`, `+`, etc. are NOT valid. For negation, create a method (e.g., `isCollapsed()` returning `not self.expanded`).

### Path Parameters

Add parameters after `?`:

```html
<!-- Trigger on every keystroke -->
<sl-input ui-value="searchQuery?keypress">

<!-- Auto-scroll to bottom on content change -->
<div class="log-viewer" ui-value="log?scrollOnOutput"></div>

<!-- Wrap array with ViewList -->
<div ui-view="contacts?wrapper=lua.ViewList">
```

| Property | Description |
|----------|-------------|
| `keypress` | Live update on every keystroke (for search boxes) |
| `scrollOnOutput` | Auto-scroll container to bottom when content changes |
| `wrapper=lua.ViewList` | Wrap array with ViewList for list rendering |
| `itemWrapper=PresenterType` | Specify presenter type for list items |

## Lists and Collections

**Standard pattern (recommended):**

```html
<!-- Basic list -->
<div ui-view="contacts?wrapper=lua.ViewList"></div>

<!-- With item wrapper -->
<div ui-view="contacts?wrapper=lua.ViewList&itemWrapper=ContactPresenter"></div>

<!-- With custom namespace -->
<div ui-view="contacts?wrapper=lua.ViewList" ui-namespace="customer-item"></div>
```

### How ViewList Works

1. You have an array: `contacts = [{obj: 1}, {obj: 2}, {obj: 3}]`
2. `wrapper=lua.ViewList` creates a ViewList wrapper
3. ViewList creates ViewListItem for each element
4. Each ViewListItem renders using `lua.ViewListItem.list-item` viewdef
5. ViewListItem has `item` property pointing to your object

### ViewListItem Viewdef

Create a viewdef for `lua.ViewListItem` with your desired namespace:

```html
<!-- lua.ViewListItem.list-item.html viewdef -->
<template>
  <div style="display: flex; align-items: center;">
    <div ui-view="item" ui-namespace="list-item" style="flex: 1;"></div>
    <sl-icon-button ui-action="remove()" name="x" label="Remove"></sl-icon-button>
  </div>
</template>
```

### ViewListItem Properties

Inside a ViewListItem viewdef, you have access to:

| Property | Description |
|----------|-------------|
| `item` | The wrapped object (or presenter if `itemWrapper=` specified) |
| `baseItem` | The original unwrapped object |
| `index` | Position in the array (0-based) |
| `list` | Reference to the ViewList |

## Namespace Resolution

Namespaces control which viewdef variant is used for rendering:

1. If variable has `namespace` property and `TYPE.{namespace}` viewdef exists, use it
2. Otherwise, if variable has `fallbackNamespace` property and `TYPE.{fallbackNamespace}` viewdef exists, use it
3. Otherwise, use `TYPE.DEFAULT`

**Setting namespace:**

```html
<div ui-namespace="COMPACT">
  <div ui-view="contact"></div>  <!-- Uses Contact.COMPACT if it exists -->
</div>
```

## Nested Views

Use `ui-view` to render child objects:

```html
<div class="contact-manager">
  <!-- List of contacts -->
  <div ui-view="contacts?wrapper=lua.ViewList"></div>

  <!-- Selected contact detail -->
  <div ui-view="selectedContact"></div>
</div>
```

The child object renders using its own viewdef based on its `type` property.

## Select Views

Use `ui-viewlist` to populate `<sl-select>` options. Note that using ui-viewlist can be a bit esoteric:

```html
<sl-select ui-value="selectedContact" ui-viewlist="contacts" ui-namespace="OPTION">
  <sl-option></sl-option>
</sl-select>
```

With a Contact.OPTION.html viewdef:

```html
<template>
    <span ui-value="name"></span>
</template>
```

## Component Libraries

The platform includes **Shoelace** web components:

```html
<sl-input label="Email" type="email" ui-value="email">
  <sl-icon name="envelope" slot="prefix"></sl-icon>
</sl-input>

<sl-button variant="primary" ui-action="save()">Save</sl-button>

<sl-rating ui-value="rating"></sl-rating>

<sl-select ui-value="status">
  <sl-option value="active">Active</sl-option>
  <sl-option value="inactive">Inactive</sl-option>
</sl-select>
```

## Common Patterns

### Form with Validation

```html
<template>
  <form class="my-form">
    <sl-input label="Name" ui-value="name" ui-attr-invalid="hasNameError()"></sl-input>
    <div class="error" ui-class-hidden="isNameValid()" ui-value="nameError"></div>

    <sl-button ui-action="submit()" ui-attr-disabled="isInvalid()">Submit</sl-button>
  </form>
</template>
```

### Conditional Display

```html
<template>
  <div>
    <div ui-class-hidden="isNotLoading()">Loading...</div>
    <div ui-class-hidden="isLoading()" ui-view="content"></div>
  </div>
</template>
```

### Scrolling Output

```html
<template>
  <div class="chat-messages" ui-view="messages?wrapper=lua.ViewList&scrollOnOutput"></div>
</template>
```
