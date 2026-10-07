# Understanding UI-Engine

This document explains the core architecture for anyone (human or AI)
working with the codebase. It focuses on the concepts that matter
most and the relationships between them.

## The Big Idea

Most UI frameworks assume you control the model and design it for
your UI. UI-engine inverts this: **the model is given, and the UI
wraps it without contaminating it.** You can build UIs on other
people's domains who weren't expecting it.

Objects present themselves. A `Contact` object doesn't know about
pages or routes -- it has viewdefs (`Contact.DEFAULT.html`,
`Contact.list-item.html`) that define how it appears in different
contexts. The same object renders differently based on namespace.

## Variables: The Core Abstraction

Everything flows through **variables**. A variable is a node in a
tree that:

- Has an ID, a parent, and a path
- Caches a value (resolved via the path from its parent)
- Carries properties (metadata as string key-value pairs)
- Is watched for changes

**The backend creates only variable 1** (the app object). The
frontend creates all other variables as the UI renders. When a
viewdef contains `ui-view="contacts"`, the frontend creates a child
variable with path `contacts` under the current context variable.
The backend resolves the path, computes the value, and sends it
back.

This means the variable tree grows dynamically as the UI renders,
driven by the frontend, with the backend supplying values.

## Value JSON and Object References

Values are serialized in a specific format:

- Primitives: as-is (`"hello"`, `42`, `true`)
- Objects (Go pointers, Lua tables with `type`): `{"obj": ID}`
- Arrays: `[{"obj": 1}, {"obj": 2}]` (element-by-element)

Object references enable identity tracking. The same Lua table
appearing in multiple places gets the same obj ID. Change detection
compares these serialized forms -- if the array of obj refs changes,
the variable has changed.

**Critical detail**: Lua array tables (no `type`) are converted to
Go slices by `ConvertToValueJSON` in the resolver. The outer
`ToValueJSON` then processes each element. The resolver must NOT
call `ToValueJSON` on elements itself -- that causes
double-processing where `ObjectRef` structs (which are Go structs)
get converted to nil by the struct handler.

## Wrappers: Presentation Without Pollution

Wrappers let you present domain objects with UI-specific state
without modifying them. When a variable has a `wrapper` property,
the resolver's `CreateWrapper` is called to produce a wrapper
object. Child variables then navigate through the wrapper instead
of the raw value.

```
Variable: contacts (wrapper=lua.ViewList)
  Raw value: Lua table [Contact#1, Contact#2, Contact#3]
  WrapperValue: *ViewList (Go object)
  NavigationValue(): returns the ViewList, not the table

  Child variable: items (path="items" on the ViewList)
    Value: ViewList.Items = [*ViewListItem, *ViewListItem, ...]
```

**Wrapper lifecycle**:
1. Created when variable has `wrapper` property and value is non-nil
2. On value change: `CreateWrapper` called again
   - Same pointer returned: wrapper updated in place (state preserved)
   - Different pointer: old wrapper unregistered, new one takes over
3. Destroyed when variable is destroyed or wrapper property cleared

The wrapper's external representation (`WrapperJSON`) is always an
object reference `{"obj": N}` -- it's the wrapper's *children* that
expose the interesting data (like the Items array).

## ViewList: The Array Wrapper

`ViewList` is the key wrapper type. It transforms a Lua array of
domain objects into an array of `ViewListItem` objects, each
carrying:

- `item`: the domain object (or a further wrapper if `itemWrapper` is set)
- `baseItem`: always the original domain object
- `index`: position in the list
- `list`: back-reference to the ViewList

**The chain for a list of contacts**:
```
Variable: contacts (wrapper=lua.ViewList)
  └── items (child, resolves ViewList.Items)
       ├── 0 (ViewListItem)
       │    └── item (Contact#1 or ContactPresenter wrapping Contact#1)
       ├── 1 (ViewListItem)
       │    └── item (Contact#2 or ContactPresenter wrapping Contact#2)
       ...
```

Each ViewListItem gets its own viewdef (`lua.ViewListItem.list-item.html`)
which typically contains `ui-view="item"` -- creating a child
variable that resolves to the presenter, which then renders with the
appropriate viewdef.

**SyncViewItems** (viewlist.go) is called when the wrapper updates:
- Grows: appends new ViewListItems
- Shrinks: destroys excess ViewListItems from the end
- Updates: checks each item for identity change, replaces if needed

## Change Detection

The change-tracker runs `DetectChanges()` after each message batch:

1. **Collect**: Walk variable tree, group by priority (high/medium/low)
2. **Check**: For each variable, call `GetValue()` to recompute from path
3. **Compare**: `ToValueJSON(currentValue)` vs cached `ValueJSON`
4. **Update wrapper**: If value changed and variable has wrapper, call `CreateWrapper`
5. **Report**: If external representation (`JsonForUpdate()`) changed, record it

Parent-before-child ordering is guaranteed. This matters because a
parent's value change may affect what its children resolve to.

**Priority** controls detection order within a level. High-priority
variables are checked first, enabling dependent calculations.

## Frontend: Views and ViewLists

### View (view.ts)
Manages a `ui-view` element. When the bound variable gets a `type`
property, the View looks up the viewdef (`Type.Namespace.html`),
clones it, replaces its element(s) in the DOM, and processes any
`ui-view` or `ui-viewlist` children.

**No-flash buffering**: Old elements stay visible while new ones are
hidden. After 100ms, old elements are removed and new ones revealed.
This prevents white-flash on re-renders.

### ViewList (viewlist.ts)
Manages a `ui-viewlist` element. Watches an array variable and
maintains one child View per array element. When the array grows or
shrinks, Views are added or removed.

### BindingEngine (binding.ts)
Processes `ui-*` attributes: `ui-value`, `ui-action`, `ui-attr-*`,
`ui-class-*`, `ui-style-*`, `ui-event-*`, `ui-html`, `ui-code`.
Creates child variables from paths and wires up DOM updates.

## Namespace Resolution

Same type, different presentation. A 3-tier cascade:

1. Variable's `namespace` property (set via `ui-namespace` attribute)
2. Variable's `fallbackNamespace` property (inherited from parent ViewList)
3. `DEFAULT`

ViewLists automatically set `fallbackNamespace` to `list-item`,
so items in a list render with `Type.list-item.html` by default.

## Protocol

WebSocket JSON messages between frontend and backend:

- **create**: Frontend creates a variable (sends ID, parentId, properties with path)
- **update**: Value or property changes (bidirectional)
- **destroy**: Remove variable and descendants (frontend sends, backend confirms per-variable)
- **watch/unwatch**: Subscribe to changes
- **error**: Error reporting

The frontend vends its own variable IDs (starting from 2). No
round-trip needed to create a variable -- the frontend creates it
optimistically and the backend fills in the value.

## Lua Resolver

The resolver (`internal/lua/resolver.go`) implements path navigation
for Lua tables and wrapper types:

- **Lua tables**: `GetField` for string paths, `RawGetInt` for numeric
- **ViewList**: `items` path returns the Items slice
- **ViewListItem**: `item`, `index`, `list` properties
- **Method calls**: `name()` syntax calls Lua methods

`ConvertToValueJSON` handles Lua-specific type conversion:
- Array tables (no `type`): converted to `[]any` (elements left as raw Go values for the caller to process)
- Object tables (has `type`): returned as `*lua.LTable` for registration as obj ref
- Non-tables: returned unchanged

## Destroy Protocol

When the frontend destroys a variable (e.g., ViewList shrinks):

1. Frontend sends `destroy(varId)`
2. Backend calls `DestroyVariable` -- removes the variable and all
   descendants from the tracker
3. Backend sends a `destroy` notification back for each destroyed
   variable (children before parents)
4. Frontend removes each variable from its local store

## Key Files

| File | Purpose |
|------|---------|
| `web/src/view.ts` | View rendering and lifecycle |
| `web/src/viewlist.ts` | Frontend array handling |
| `web/src/connection.ts` | WebSocket, VariableStore, destroy handling |
| `web/src/binding.ts` | `ui-*` attribute processing |
| `internal/lua/viewlist.go` | ViewList wrapper (Go side) |
| `internal/lua/viewlistitem.go` | ViewListItem type |
| `internal/lua/resolver.go` | Lua path navigation, ConvertToValueJSON |
| `internal/lua/runtime.go` | Lua session, executor, AfterBatch |
| `internal/protocol/handler.go` | Protocol message routing |
| `change-tracker/tracker.go` | Variable tree, change detection, wrappers |
| `USAGE.md` | Complete feature guide with examples |
