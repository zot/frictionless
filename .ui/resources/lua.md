# Lua API & Patterns Guide

The backend logic is written in Lua (GopherLua). This guide covers the idiomatic patterns for building UI logic.

## Defining Classes (Hot-Loadable)

Use `session:prototype()` to define classes that support hot-loading.

### Prototype Naming Convention

The `session:prototype(name, init)` function accepts arbitrary prototype names and does not consult global variables. The `name` becomes the prototype's `type` field, which is used for viewdef resolution (e.g., `MyForm` → `MyForm.DEFAULT.html`). Each app creates two globals with minimal namespace pollution:

- **Name** (PascalCase) — The app prototype, which also serves as a namespace for related prototypes
- **name** (camelCase) — The instance that ui-mcp uses to display the app

| App Directory | Prototype/Namespace | Instance Variable |
|---------------|---------------------|-------------------|
| `contacts`    | `Contacts`          | `contacts`        |
| `my-form`     | `MyForm`            | `myForm`          |

### Complete App Example

```lua
-- 1. Declare app prototype (serves as namespace)
-- init declares instance fields — only these are tracked for mutation
MyForm = session:prototype("MyForm", {
    userInput = "",
    error = EMPTY,  -- EMPTY: starts nil, but tracked for mutation
})

-- Prototype-level variables (shared across instances, not in init)
MyForm.submitCount = MyForm.submitCount or 0

-- 2. Nested prototype with dotted name
MyForm.FormEntry = session:prototype('MyForm.FormEntry', {
    value = "",
    timestamp = "",
})
local FormEntry = MyForm.FormEntry  -- Local shortcut

function FormEntry:new(data)
    return session:create(FormEntry, data)
end

-- 3. Override :new() only when you need custom initialization
-- Default :new() just calls session:create() automatically
function MyForm:new(instance)
    instance = session:create(MyForm, instance)
    instance.id = MyForm.submitCount
    MyForm.submitCount = MyForm.submitCount + 1
    return instance
end

-- 4. Define methods
function MyForm:submit()
    mcp.pushState({ app = "my-form", event = "submit", value = self.userInput })
end

-- 5. Guard instance creation (idempotent)
if not session.reloading then
    myForm = MyForm:new()
end
```

The agent then calls `ui_display("my-form")` to show it in the browser.

**Key points:**
- `session:prototype(name, init)` — accepts arbitrary names, preserves identity on reload
- `session:create(prototype, instance)` — tracks instances for hot-reload
- `Name.NestedType = session:prototype('Name.NestedType', ...)` — nested prototypes with dotted names
- `local NestedType = Name.NestedType` — local shortcut for cleaner code
- `EMPTY` — declare optional fields that start nil but are tracked for mutation
- `if not session.reloading` — guard prevents re-creating instance on hot-reload
- Only two globals per app: `Name` (prototype/namespace) and `name` (instance)

## Global Objects

### 1. `session`
Provides access to session-level services and hot-loading support.
- `session:prototype(name, init)`: Define a hot-loadable class
- `session:create(prototype, instance)`: Create a tracked instance
- `session.reloading`: `true` during hot-reload, `false` otherwise

### 2. `mcp` (AI Agents Only)
Provides display and communication for AI Agents.
- `mcp.pushState(event)`: Queue an event for the agent (polled via `/wait` endpoint)

The agent uses the `ui_display("varName")` tool to show objects in the browser.

## Schema Migrations

When you modify prototype fields, use `mutate()` for data migrations:

**Adding fields:**
```lua
MyForm = session:prototype("MyForm", {
    userInput = "",
    timestamp = "",  -- NEW: inherited automatically via metatable
})

-- Optional: compute initial values for existing instances
function MyForm:mutate()
    self.timestamp = self.timestamp or os.date()
end
```

**Removing fields:** Just remove from init. Framework nils out the field on all instances automatically.

**Renaming fields:**
```lua
function MyForm:mutate()
    if self.oldName then
        self.newName = self.newName or self.oldName
        self.oldName = nil
    end
end
```

## Change Detection

The platform uses **Automatic Change Detection**. You do not need to call `update()` or `notify()` when you change properties on a Lua table. The system detects modifications after every message batch and pushes changes to the frontend.

```lua
function MyForm:clear()
    -- These changes are automatically detected and sent to the browser
    self.userInput = ""
    self.error = nil
end
```

## Tips for AI Agents

- **Modules:** Use `require` to load standard libraries or other files.
- **Error Handling:** Errors in Lua code will be reported back through the `ui_run` tool.
- **Persistence:** Use `mcp:status().base_dir` to get the base directory for reading/writing local files.
