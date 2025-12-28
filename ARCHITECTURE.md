# UI Platform Design Approaches

## Core Architecture: Presenters and Domain Objects

The UI platform uses a clear separation between **domain objects** (data/business logic) and **presenter objects** (UI state/behavior).

### Domain Objects
- Hold app data and core behavior
- Example: `Contact` with firstName, lastName, email, phone
- Can have methods like `fullName()` for computed properties
- Should NOT have UI-specific methods like `delete()` (a contact can be in many lists)

### Presenter Objects
- Wrap domain objects and add UI state/behavior
- Hold references to parent presenters for context-aware actions
- Example: `ContactPresenter` wraps `Contact` and adds:
  - `delete()` - removes from the parent manager's list
  - `edit()` - sets editing state
  - `isExpanded`, `validationErrors` - UI state

### App Presenter
- Central control point, like a "server view of the browser"
- Everything visible on screen is reachable from it
- Holds the current object/presenter "on display"
- Backend creates this as variable 1, kicking off the UI

## Variable Creation Flow

**Critical insight: The frontend creates most variables, not the backend.**

1. Backend creates only **variable 1** (app presenter with domain data)
2. Backend sends viewdefs for presenter types
3. Frontend renders viewdefs which contain paths like `ui-view="selectedContact"`
4. These paths create variables that "reach into" backend objects
5. Path resolution may call methods, returning presenters
6. The backend indirectly instructs itself to create variables through viewdefs

This is the key quote: *"The backend creating variables when instructed by the frontend which, in turn, gets those from the viewdefs it parses from the backend."*

### Path Resolution: Always Server-Side

**⚠️ CRITICAL: See [viewdefs.md - Path Resolution: Server-Side Only](specs/viewdefs.md#path-resolution-server-side-only)**

Variable values are **object references** (`{"obj": 1}`), not actual data. All path-based bindings (`ui-value`, `ui-attr-*`, `ui-class-*`, `ui-style-*-*`) MUST create child variables - the backend resolves paths, not the frontend.

**Why object references?**
- Enables object identity tracking across the UI
- Allows the same object to appear in multiple places
- Supports circular references without infinite serialization
- Change detection compares object identity, not deep equality

## Variable Wrappers: Transforming Values

The `wrapper` property on variables enables value transformation at the backend. When a variable has a wrapper, the backend uses `Wrapper(variable)` to compute the outgoing value instead of sending the raw value directly.

### The Wrapper Property

Any variable can have a `wrapper` property set via path syntax:

```html
<!-- Direct wrapper usage in viewdef -->
<div data-ui-view="selectedContact?wrapper=ContactPresenter">

<!-- Wrapper with additional properties -->
<div data-ui-path="currentUser?wrapper=UserPresenter&editable=true">
```

The wrapper:
- Receives the **variable** (not just the value), enabling it to watch for changes
- Computes the outgoing JSON value sent to the frontend
- Can create/manage additional objects (like presenters)

### Variable Value Architecture

Variables need two distinct values:

1. **Monitored value** - Used to detect changes
   - For arrays: a copy so content changes are detected
   - For other values: tracks the raw value from the path

2. **Outgoing JSON value** - What gets sent to frontend
   - Without wrapper: monitored value in "value JSON" form (objects as `{obj: ID}` refs, not inline)
   - With wrapper: computed by `Wrapper(variable)`
   - Enables transformation (e.g., domain object ref → presenter object ref)

### Wrapper Use Cases

```html
<!-- Wrap a single object in a presenter -->
<div data-ui-view="contact?wrapper=ContactPresenter">

<!-- Wrap with editable form presenter -->
<sl-input data-ui-path="user.email?wrapper=EditableField&validate=email">

<!-- Custom computed value -->
<span data-ui-text="items?wrapper=CountDisplay">  <!-- shows "3 items" -->
```

## ViewList: A High-Level Widget Using Wrappers

`ViewList` is a built-in high-level widget that automatically sets the `wrapper` property. It's not the only way to use wrappers—it's just a convenient pattern for array handling.

### How ViewList Uses Wrappers

When a viewdef uses `data-ui-viewlist="contacts"`, the frontend:
1. Creates a variable for `contacts`
2. Automatically adds `wrapper=ViewList` property
3. Backend uses ViewList to process the array value
4. ViewList manages a parallel list of presenter objects

### Path Syntax for ViewList Configuration

```html
<!-- Basic: uses generic ViewList -->
<div data-ui-viewlist="contacts">

<!-- With custom item presenter -->
<div data-ui-viewlist="contacts?item=ContactPresenter">
```

The `?item=ContactPresenter` is passed to ViewList, telling it which presenter class to wrap each item with.

### How ViewList Works Internally

```
Backend domain data:
  contacts: [{obj: 101}, {obj: 102}, {obj: 103}]  ← Contact object refs

ViewList.computeValue(rawArray) creates ViewItem objects:
  viewItems: [ViewItem1, ViewItem2, ViewItem3]

Each ViewItem:
  ├── baseItem: {obj: 101}    ← reference to domain object (Contact)
  ├── item: {obj: P1}         ← wrapped presenter (if item=ContactPresenter)
  │         or {obj: 101}     ← same as baseItem (if no item property)
  ├── list: ViewList          ← back-reference to list
  └── index: 0                ← position in list

Each ContactPresenter (created via ItemWrapper(viewItem)):
  ├── has access to viewItem.baseItem (the Contact)
  ├── has access to viewItem.list (for delete operations)
  └── methods: delete(), select(), etc.

Outgoing JSON (what frontend receives):
  [{obj: VI1}, {obj: VI2}, {obj: VI3}]  ← ViewItem refs

ViewItem viewdef uses ui-view="item" to display the presenter/domain object
```

### Why This Approach

1. **Backend stays simple** - Just stores domain object arrays, no presenter management
2. **Frontend controls presentation** - Decides which presenter wraps which domain type
3. **Automatic synchronization** - ViewList watches domain array, manages presenters in parallel
4. **Flexible per-view** - Same `contacts` array can have different presenters in different views

### Comparison: Manual vs Wrapper-Based

**Old approach (backend manages presenters):**
```lua
-- Backend creates and manages presenters
local contactPresenters = {}
for _, contact in ipairs(contacts) do
    table.insert(contactPresenters, ContactPresenter:new(contact, self))
end
app.contacts = contactPresenters  -- stores presenter refs
```

**New approach (wrapper manages presenters):**
```lua
-- Backend just stores domain data
app.contacts = contacts  -- stores domain object refs
-- Wrapper (via ViewList) handles presenter creation based on viewdef
```

### Other High-Level Widgets Can Use Wrappers

ViewList isn't special—other widgets can also set wrapper properties:

```html
<!-- Hypothetical DataGrid widget that wraps rows -->
<ui-datagrid data-ui-source="transactions?wrapper=DataGrid&row=TransactionRow">

<!-- Tree widget that wraps nodes -->
<ui-tree data-ui-root="fileSystem?wrapper=TreeView&node=FileNode">
```

The `wrapper` property is the fundamental mechanism; ViewList and other widgets are conveniences built on top of it.

## Presenters in Collections: Design Options

### Option A: Backend Holds Domain Objects + ViewList (Recommended)
```
ContactManagerPresenter (backend)
├── contacts: [Contact1, Contact2, Contact3]  ← domain object refs
├── contactIndex: 0
└── selectedContact() → contacts[contactIndex]

ViewList (created by frontend viewdef)
├── source: contacts variable
├── presenters: [ContactPresenter1, ContactPresenter2, ContactPresenter3]
└── manages: presenter lifecycle, sync with source
```

### Option B: Backend Holds Presenters Directly
```
ContactManagerPresenter
├── contacts: [ContactPresenter1, ContactPresenter2, ContactPresenter3]
├── contactIndex: 0
└── selectedContact → contacts[contactIndex]
```

### Why Option A (ViewList) is Better

1. **Separation of concerns** - Backend handles domain logic, frontend handles presentation
2. **Flexibility** - Same domain list can have different presenters in different views
3. **Simpler backend** - No presenter management code in Lua
4. **Automatic sync** - ViewList keeps presenters in sync with domain array changes

### When to Use Option B (Backend Presenters)

- Complex presenter logic that must run on backend
- Presenters need to persist across frontend reconnections
- Performance-critical scenarios where frontend can't handle presenter creation

## Example: Contact Manager Structure

### Backend (Lua) - Simple Domain Data

```
ContactApp (variable 1)
├── contacts: [{obj: 101}, {obj: 102}, {obj: 103}]  ← domain object refs
├── selectedIndex: null or 0, 1, 2...
├── searchQuery: ""
├── addContact() → creates Contact, adds ref to contacts array
└── removeContact(index) → removes from contacts array

Contact (domain objects, stored by ref)
├── id, firstName, lastName, email, phone, notes
└── fullName() → firstName + " " + lastName
```

### Frontend (ViewList manages presenters)

```
ViewList (created from data-ui-viewlist="contacts?item=ContactPresenter")
├── sourceVariable: contacts variable
├── presenters: [ContactPresenter1, ContactPresenter2, ContactPresenter3]
└── syncs presenters with source array changes

ContactPresenter (created by ViewList for each item)
├── item: {obj: 101}        ← reference to Contact
├── list: ViewList          ← back-reference for delete
├── index: 0                ← position in list
├── isEditing: false        ← UI state
├── delete() → list.removeAt(index)
└── select() → app.selectedIndex = index
```

### Data Flow

1. Backend has `contacts: [{obj: 101}, {obj: 102}]`
2. Viewdef: `data-ui-viewlist="contacts?item=ContactPresenter"`
3. Frontend creates variable with `wrapper=ViewList, item=ContactPresenter`
4. Backend sees wrapper, uses ViewList to process value
5. ViewList creates ContactPresenter for each Contact ref
6. Variable sends `[{obj: P1}, {obj: P2}]` to frontend (presenter refs)
7. Frontend renders each presenter using ContactPresenter viewdef

## Viewdef Example

```html
<!-- ContactApp viewdef -->
<div class="contact-manager">
  <sl-input data-ui-path="searchQuery" placeholder="Search..."/>
  <sl-button data-ui-action="addContact()">Add Contact</sl-button>

  <!-- List with automatic presenter wrapping -->
  <div data-ui-viewlist="contacts?item=ContactPresenter" data-ui-namespace="list-item">
    <!-- Each item renders ContactPresenter.list-item viewdef -->
  </div>

  <!-- Selected contact detail (uses selectedIndex to pick from presenter list) -->
  <div data-ui-view="contacts[selectedIndex]">
    <!-- Renders ContactPresenter.DEFAULT viewdef when selected -->
  </div>
</div>

<!-- ContactPresenter.list-item viewdef (compact row) -->
<div class="contact-row" data-ui-action="select()">
  <span data-ui-path="item.fullName()"/>
  <span data-ui-path="item.email"/>
  <sl-icon-button name="trash" data-ui-action="delete()"/>
</div>

<!-- ContactPresenter.DEFAULT viewdef (full detail) -->
<div class="contact-detail">
  <h2 data-ui-text="item.fullName()"/>
  <sl-input data-ui-path="item.firstName" label="First Name"/>
  <sl-input data-ui-path="item.lastName" label="Last Name"/>
  <sl-input data-ui-path="item.email" label="Email"/>
  <sl-input data-ui-path="item.phone" label="Phone"/>
  <sl-textarea data-ui-path="item.notes" label="Notes"/>
  <sl-button data-ui-action="delete()" variant="danger">Delete</sl-button>
</div>
```

## Key Principles

1. **Backend stores domain objects** - Simple data, no UI concerns
2. **Frontend creates variables via viewdefs** - Paths in viewdefs cause variable creation
3. **ViewList wraps domain arrays** - Automatically creates presenter objects
4. **Presenters add UI state/behavior** - `isEditing`, `delete()`, etc.
5. **Path properties configure wrappers** - `?item=ContactPresenter` sets presenter type
6. **Variables have monitored + outgoing values** - Detect changes, send transformed data

## Spec Changes Required

1. **Variable properties via path syntax** - `path?key=value&key2=value2`
2. **Variable wrapper property** - `wrapper=ViewList` triggers value transformation
3. **Variable dual values** - Monitored value (for change detection) vs outgoing JSON
4. **ViewList as built-in wrapper** - Creates/manages presenter objects for array items
5. **ViewList receives variable, not just value** - Can watch for changes and manage lifecycle
