---
name: Method Caching
description: Caching patterns for expensive method paths to prevent slowdowns
---

# Method Caching Patterns

Methods bound to viewdefs via `ui-value`, `ui-html`, `ui-attr-*`, etc. are called on every render cycle. Expensive computations must be cached to prevent UI slowdowns.

## Pattern 1: Immediate Caching (Input-Based)

Cache result based on input value. Return cached result if input hasn't changed.

**Use when:** Processing large data that changes infrequently (file content, parsed data).

```lua
function MyApp:expensiveHtml()
    local content = self.sourceContent
    if not content or content == "" then
        self._cachedHtml = ""
        self._cachedInput = ""
        return ""
    end

    -- Return cached result if input unchanged
    if self._cachedInput == content then
        return self._cachedHtml
    end

    -- Expensive processing
    local result = processContent(content)

    -- Cache for next call
    self._cachedInput = content
    self._cachedHtml = result
    return result
end
```

**Key points:**
- Store both the input and the result
- Compare input before recomputing
- Clear cache when source is cleared

## Pattern 2: Batched Caching (Time-Based TTL)

Refresh data for multiple items in one batch operation, with time-based expiration.

**Use when:** Multiple items need the same type of data that's expensive to fetch individually (filesystem checks, API calls).

```lua
-- In the item's method (called per-item):
function AppInfo:hasCheckpoints()
    -- Trigger batch refresh if cache is stale (1 second TTL)
    local now = os.time()
    if not appConsole._checkpointsTime or (now - appConsole._checkpointsTime) >= 1 then
        appConsole:refreshCheckpoints()
    end
    return self._hasCheckpoints or false
end

-- Batch refresh method (refreshes ALL items at once):
function AppConsole:refreshCheckpoints()
    for _, app in ipairs(self._apps) do
        -- Fetch data for each app
        app._hasCheckpoints = checkFilesystem(app.name)
        app._checkpointCount = countCheckpoints(app.name)
    end
    self._checkpointsTime = os.time()  -- Update timestamp
end
```

**Key points:**
- Store timestamp on the parent/container object
- First item to check triggers refresh for ALL items
- Subsequent items in same render cycle get cached data
- TTL prevents constant re-fetching (1 second is typical)

## When to Use Each Pattern

| Pattern | Use Case | Example |
|---------|----------|---------|
| Immediate | Single expensive computation | HTML escaping + highlighting |
| Batched | N items needing same data type | Filesystem status for app list |

## Warning Signs You Need Caching

- UI becomes sluggish when displaying certain data
- Console shows many repeated method calls
- Methods that call shell commands (`io.popen`)
- Methods that process large strings line-by-line
- Methods called for each item in a list
