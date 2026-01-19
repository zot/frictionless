# ViewList - Internal Component

ViewLists are **internal to frictionless** - there is no Lua code for them. The ViewList and ViewListItem types are implemented in the ui-engine backend and provide automatic list rendering.

## How ViewLists Work

A **ViewList** renders an array of object references as a list of views.

### Usage

Use `ui-view` with the `wrapper=lua.ViewList` path property:

```html
<!-- Basic list -->
<div ui-view="contacts?wrapper=lua.ViewList"></div>

<!-- With item presenter wrapper -->
<div ui-view="contacts?wrapper=lua.ViewList&itemWrapper=ContactPresenter"></div>
```

### Alternative: ui-viewlist attribute

The `ui-viewlist` attribute is shorthand that implicitly uses `lua.ViewList`:

```html
<div ui-viewlist="contacts"></div>
<!-- Equivalent to: <div ui-view="contacts?wrapper=lua.ViewList&access=r"></div> -->
```

## View Chain

When a ViewList renders, it creates this chain:

```
ViewList (lua.ViewList)
  └── ViewListItem (lua.ViewListItem) - one per array element
        └── Domain Object (e.g., Contact) - via ui-view="item"
```

### ViewListItem Properties

Each ViewListItem automatically has:
- `item` - Pointer to the domain object from the array
- `list` - Pointer to the parent ViewList
- `index` - Position in the list (0-based)

## Namespace Resolution

ViewList uses `list-item` as the fallback namespace:

1. ViewList variable gets `fallbackNamespace: "list-item"` from the backend wrapper
2. ViewListItem inherits this fallback
3. Domain object inherits it via `ui-view="item"`

### Namespace Lookup Order

For each type in the chain:
1. Try `TYPE.{namespace}` if namespace property is set
2. Try `TYPE.{fallbackNamespace}` (typically `list-item`)
3. Fall back to `TYPE.DEFAULT`

### Example with Custom Namespace

```html
<div ui-view="customers?wrapper=lua.ViewList" ui-namespace="customer-item"></div>
```

Resolution:
1. `lua.ViewList.customer-item` → fallback to `lua.ViewList.list-item`
2. `lua.ViewListItem.customer-item` → fallback to `lua.ViewListItem.list-item`
3. `Customer.customer-item` → if exists, use it; else fallback to `Customer.list-item`

## Viewdef Files

This directory contains two viewdefs:

- `lua.ViewList.DEFAULT.html` - Container with `ui-viewlist="items"`
- `lua.ViewListItem.list-item.html` - Renders `ui-view="item"` to delegate to domain object

### lua.ViewList.DEFAULT.html

```html
<template>
  <div class="view-list">
    <div ui-viewlist="items" ui-namespace="list-item"></div>
  </div>
</template>
```

### lua.ViewListItem.list-item.html

```html
<template>
  <div ui-view="item"></div>
</template>
```

The ViewListItem viewdef delegates rendering to the domain object's `list-item` viewdef (e.g., `Contact.list-item.html`).
