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
| `ui-action` | Bind click to Lua method | `<sl-button ui-action="save()">` |
| `ui-event-*` | Bind any event to Lua method | `<sl-select ui-event-sl-change="onSelect()">` |
| `ui-view` | Render child object with its viewdef | `<div ui-view="selectedItem">` |
| `ui-viewlist` | Render array as list | `<div ui-viewlist="items">` |
| `ui-attr-*` | Bind HTML attribute to Lua path | `<sl-alert ui-attr-open="hasError">` |
| `ui-class-*` | Toggle CSS class on boolean | `<div ui-class-active="isSelected">` |
| `ui-style-*` | Bind CSS style to Lua path | `<div ui-style-color="themeColor">` |

**Note:** `ui-value` works for both input elements (sets `.value`) and display elements like `<span>` (sets text content).

## Path Syntax

Paths are resolved relative to the current object on the server.

```
property           → self.property
nested.path        → self.nested.path
method()           → self:method()
method(arg)        → self:method(arg)
items[0]           → self.items[1] (Lua is 1-indexed)
items[index]       → self.items[self.index + 1]
```

### Path Parameters

Add parameters after `?`:

```html
<!-- Trigger on every keystroke -->
<sl-input ui-value="searchQuery?keypress">

<!-- Wrap value in a presenter -->
<div ui-view="contact?wrapper=ContactPresenter">
```

## Lists and Collections

Use `ui-viewlist` to render arrays:

```html
<!-- Basic list -->
<div ui-viewlist="contacts"></div>

<!-- With item wrapper -->
<div ui-viewlist="contacts?item=ContactPresenter"></div>
```

### How ViewList Works

1. You have an array: `contacts = [{obj: 1}, {obj: 2}, {obj: 3}]`
2. `ui-viewlist` creates a ViewList wrapper
3. ViewList creates ViewItem for each element
4. Each ViewItem renders using `lua.ViewListItem` viewdef
5. ViewItem has `item` property pointing to your object (or wrapped presenter)

### ViewListItem Viewdef

Create a viewdef for `lua.ViewListItem` with your desired namespace:

```html
<!-- lua.ViewListItem.contact-row viewdef -->
<template>
  <div class="contact-row" ui-action="select()">
    <span ui-value="item.fullName()"></span>
    <span ui-value="item.email"></span>
    <sl-icon-button name="trash" ui-action="delete()"></sl-icon-button>
  </div>
</template>
```

Use it:

```html
<div ui-viewlist="contacts?item=ContactPresenter" ui-namespace="contact-row"></div>
```

### ViewItem Properties

Inside a ViewListItem viewdef, you have access to:

| Property | Description |
|----------|-------------|
| `item` | The wrapped object (or presenter if `?item=` specified) |
| `baseItem` | The original unwrapped object |
| `index` | Position in the array (0-based) |
| `list` | Reference to the ViewList |

## Nested Views

Use `ui-view` to render child objects:

```html
<div class="contact-manager">
  <!-- List of contacts -->
  <div ui-viewlist="contacts?item=ContactPresenter" ui-namespace="list-item"></div>

  <!-- Selected contact detail -->
  <div ui-view="selectedContact"></div>
</div>
```

The child object renders using its own viewdef based on its `type` property.

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
    <sl-input label="Name" ui-value="name" ui-attr-invalid="errors.name"></sl-input>
    <div class="error" ui-class-visible="errors.name" ui-value="errors.name"></div>

    <sl-button ui-action="submit()" ui-attr-disabled="!isValid">Submit</sl-button>
  </form>
</template>
```

### Conditional Display

```html
<template>
  <div>
    <div ui-class-hidden="!isLoading">Loading...</div>
    <div ui-class-hidden="isLoading" ui-view="content"></div>
  </div>
</template>
```

### Action with Arguments

```html
<template>
  <sl-button ui-action="setPage(1)">First</sl-button>
  <sl-button ui-action="setPage(currentPage - 1)">Prev</sl-button>
  <sl-button ui-action="setPage(currentPage + 1)">Next</sl-button>
</template>
```
