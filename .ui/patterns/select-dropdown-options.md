# Select Dropdown with Dynamic Options

Use ViewList inside `<sl-select>` to populate options from a Lua array.

```html
<sl-select ui-value="selectedId" label="Pick one">
  <span ui-view="items()?wrapper=lua.ViewList" ui-namespace="my-option"></span>
</sl-select>
```

Viewdef (`lua.ViewListItem.my-option.html`):
```html
<template>
  <sl-option ui-attr-value="index">
    <span ui-value="item.name"></span>
  </sl-option>
</template>
```

Key points:
- Use a `<span>` as the container (it gets replaced by the options)
- The viewdef type is `lua.ViewListItem`, not your domain type
- `index` is 0-based position; `item` is the wrapped object
