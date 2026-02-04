---
description: Use full buffering when writing many small chunks to files in Lua. Critical for performance - can be 10x+ faster.
---

# Lua File Buffering

When writing many small pieces of data to a file (byte-by-byte, line-by-line), enable full buffering for dramatically better performance.

## The Problem

Lua's default file buffering may flush to disk frequently, causing massive slowdowns when writing many small chunks:

```lua
-- SLOW: Each write may trigger disk I/O
local handle = io.open(path, "wb")
for i = 1, 100000 do
    handle:write(string.char(data[i]))  -- Potentially 100k disk writes!
end
handle:close()
```

## The Solution

Use `handle:setvbuf('full', bufferSize)` immediately after opening:

```lua
-- FAST: Writes accumulate in memory, flush in large batches
local handle = io.open(path, "wb")
handle:setvbuf('full', 4096)  -- 4KB buffer (or 8192, 16384 for larger files)
for i = 1, 100000 do
    handle:write(string.char(data[i]))  -- Buffered in memory
end
handle:close()  -- Final flush
```

## Buffer Modes

| Mode | Behavior | Use Case |
|------|----------|----------|
| `'no'` | No buffering, immediate writes | Real-time logging |
| `'line'` | Flush on newlines | Text output, logs |
| `'full'` | Flush only when buffer full | Binary data, bulk writes |

## Recommended Buffer Sizes

| File Size | Buffer Size |
|-----------|-------------|
| < 100KB | 4096 (4KB) |
| 100KB - 1MB | 8192 (8KB) |
| > 1MB | 16384 (16KB) or higher |

## Complete Example: Base64 Decode to File

```lua
local function decodeBase64ToFile(data, handle)
    handle:setvbuf('full', 4096)  -- CRITICAL: Enable full buffering

    local eq = string.byte('=')
    local i = 1
    while i <= #data do
        local b1, b2, b3, b4 = string.byte(data, i, i+3)
        local c1 = b64decode[b1] or 0
        local c2 = b64decode[b2] or 0
        local c3 = b64decode[b3] or 0
        local c4 = b64decode[b4] or 0

        local n = c1 * 262144 + c2 * 4096 + c3 * 64 + c4

        handle:write(string.char(math.floor(n / 65536) % 256))
        if b3 ~= eq then
            handle:write(string.char(math.floor(n / 256) % 256))
        end
        if b4 ~= eq then
            handle:write(string.char(n % 256))
        end
        i = i + 4
    end
    -- Buffer auto-flushes on close
end

-- Usage
local handle = io.open(path, "wb")
if handle then
    decodeBase64ToFile(base64data, handle)
    handle:close()
end
```

## Key Points

- **Call `setvbuf` immediately after `io.open`** - before any writes
- **Use `'full'` mode for binary/bulk data** - accumulates writes in memory
- **Buffer flushes automatically on `close()`** - no manual flush needed
- **10x+ performance improvement** for many small writes
- **Memory tradeoff** - larger buffers use more RAM but are faster
