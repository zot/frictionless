---
description: Use JSON encoding/decoding in Lua via the bundled rxi/json.lua library
---

# Lua JSON

Parse and generate JSON in Lua using the bundled `mcp/json.lua` module.

## Setup

```lua
local json = require("mcp.json")
```

This returns the json table with `encode` and `decode` methods. The module does NOT create a global — you must assign the return value.

## Encoding (Lua → JSON string)

```lua
local str = json.encode({name = "Alice", items = {1, 2, 3}})
-- '{"name":"Alice","items":[1,2,3]}'
```

## Decoding (JSON string → Lua)

```lua
local data = json.decode('{"name":"Alice","age":30}')
print(data.name)  -- "Alice"
print(data.age)   -- 30
```

## Safe Decoding

`json.decode` throws on invalid input. Use `pcall` for untrusted data:

```lua
local ok, data = pcall(json.decode, inputStr)
if not ok or not data then
    -- Handle error
    return
end
-- Use data safely
```

## Key Points

- Arrays encode as JSON arrays, tables with string keys encode as objects
- `nil` values are omitted from encoded output
- Nested tables work as expected
- The library is at `apps/mcp/json.lua` (rxi/json.lua)
