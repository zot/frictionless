---
name: ViewList Item Presenters (itemWrapper)
description: Wrap each list item with a presenter object so it can carry methods and per-item UI state, separately from the domain data
---

# ViewList Item Presenters

`lua.ViewList` has a second wrapper hook — `itemWrapper` — that creates a **presenter** object per list item. The presenter holds UI-side state and methods; the **domain object** (whatever is in the underlying list) stays untouched.

Use this when:

- List items need action methods (`select()`, `edit()`, `dismiss()`).
- List items carry per-item reactive UI state (`_open`, `_loading`, `_suggestions`).
- The domain object is a plain table (e.g. a host-owned mirror) and you want a Lua-side prototype layered over it.

Don't use this when the items already are full Lua prototypes with the methods you need — bind directly.

## The binding

```html
<div ui-view="items?wrapper=lua.ViewList&itemWrapper=Type"></div>
```

Two wrappers stack:

- `?wrapper=lua.ViewList` — the outer wrapper that turns the array into a renderable list.
- `&itemWrapper=Type` — the inner wrapper applied per element. Frictionless calls `Type:new(listItem)` for each item and uses `Type` to resolve the list-item viewdef.

`items` can be an array property, a method that returns an array (`items()`), or any path that resolves to a sequence. Each element of that array is what the presenter wraps.

## The presenter

```lua
ContactPresenter = session:prototype("ContactPresenter", {
    viewItem = EMPTY,
    contact  = EMPTY,
})

function ContactPresenter:new(listItem)
    return session:create(ContactPresenter, {
        viewItem = listItem,
        contact  = listItem.baseItem,  -- the domain item
    })
end

function ContactPresenter:edit()
    contactApp:editContact(self.contact)
end

function ContactPresenter:noPhone()
    return not self.contact.phone or self.contact.phone == ""
end
```

The argument to `:new` is the **listItem** (a Go-side `ViewListItem`), not the domain object directly. It exposes:

- `listItem.baseItem` — the actual element from the source array (the domain object).
- `listItem.list`     — the parent `ViewList`.
- `listItem.index`    — the position (Go-side index).

Most presenters just stash `listItem` and pull `baseItem` out for convenience.

Per-item reactive UI state goes on the presenter prototype as `_`-prefixed fields. They survive across renders as long as the presenter does (see "Identity preservation" below).

## The list-item viewdef

The viewdef filename uses the **presenter** type name, not the domain type:

```
ContactPresenter.list-item.html
```

Inside the viewdef, **the presenter is the context** — method calls and field reads resolve against it. Reach the domain via the field name the presenter exposes:

```html
<template>
  <div class="contact-card">
    <div ui-value="contact.fullName()"></div>     <!-- via presenter.contact -->
    <div ui-value="contact.email"></div>
    <div ui-attr-hidden="noPhone"
         ui-value="contact.phone"></div>          <!-- presenter method, presenter field -->
    <sl-button ui-action="edit()"></sl-button>    <!-- presenter method -->
    <sl-button ui-action="delete()"></sl-button>
  </div>
</template>
```

`noPhone` (no parens) is a property/method access on the presenter; `contact.phone` traverses into `presenter.contact` and reads `phone` on the domain object.

## Identity preservation

`ViewList` reuses a presenter across renders **only when the base item is reference-equal between syncs** (`view.BaseItem == newItem`). If your source array hands out fresh tables on each access — even with the same contents — `ViewList` treats every item as new and recreates every presenter, wiping their reactive state.

Two ways to keep identity stable:

- **Stable domain objects.** The source array holds the same Lua references across reads. New items get new tables; existing items reuse their tables (in-place mutation only). This is the natural shape for Lua-side state.
- **Host-mirrored arrays:** if a Go side rebuilds the Lua mirror table on each mutation, it has to cache and reuse entry tables by stable key (e.g. an ID) instead of allocating fresh ones. Otherwise every host-side write looks like "every item replaced."

When you can't guarantee either, fall back to storing reactive state outside the presenter — keyed by domain ID in a side-map on the parent object — so presenter recreation doesn't lose it. But that defeats most of the point; prefer fixing identity at the source.

## Quick checklist

When adding `itemWrapper=Type` to a list binding:

1. Define `Type = session:prototype("Type", { viewItem = EMPTY, ... })`.
2. Implement `Type:new(listItem)` that stores `viewItem` and any convenience aliases.
3. Add `Type.list-item.html` viewdef. Inside, bind to presenter methods and traverse domain fields through whatever name the presenter uses.
4. Add presenter methods for each action the list-item viewdef references.
5. Confirm the source array preserves item identity across renders (or accept presenter recreation).

## Reference

Working example: `/home/deck/work/ui-engine/demo/lua/contact.lua` (ContactPresenter) plus `/home/deck/work/ui-engine/demo/viewdefs/ContactPresenter.list-item.html`.

Implementation: `/home/deck/work/ui-engine/internal/lua/viewlist.go` — see `Sync()` for the identity-preservation rule (`view.BaseItem != item` triggers recreation).
