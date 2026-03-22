# Incremental Search Pattern

Search-as-you-type without timers or debounce. The variable check cycle and Lua's single-threaded blocking provide natural event compression.

## How It Works

1. Bind the search input with `?keypress` so every keystroke updates the variable
2. Bind a method via `ui-value` that fires on every variable check cycle
3. The method compares the current query against the last searched query
4. If changed, 3+ chars, and no search pending: fire the search immediately
5. The search (`io.popen`) blocks Lua — keystrokes queue during the block
6. When the search returns, the next variable check sees the final query value
7. If the query changed while searching, it searches again automatically

No `setTimeout`, no debounce delay, no timer management.

## Why It Works

- **Lua is single-threaded.** No race conditions, no concurrent searches.
- **`io.popen` blocks.** While search runs, the UI server queues incoming keystrokes. When the search finishes, the next cycle processes all queued updates at once and sees only the final value.
- **The variable check cycle is the polling loop.** It runs after every state change, so `onSearchInput()` fires naturally without scheduling.

This is event compression where the "frame" is the search duration itself. Fast searches (FTS) complete between keystrokes. Slow searches (vector) naturally batch multiple keystrokes.

## Template

```lua
MyApp = session:prototype("MyApp", {
    searchQuery = "",
    _lastSearchedQuery = "",
    _searchResults = EMPTY,
})

function MyApp:onSearchInput()
    local q = self.searchQuery
    if q == self._lastSearchedQuery then return "" end
    if q == "" then
        self:clearSearch()
        return ""
    end
    if #q < 3 then return "" end
    self:search()
    return ""
end

function MyApp:search()
    if self.searchQuery == "" then return end
    self._lastSearchedQuery = self.searchQuery
    -- io.popen blocks — keystrokes queue naturally
    local output = run('my-search-command "' .. self.searchQuery .. '"')
    -- parse output into _searchResults ...
end
```

```html
<!-- Hidden span triggers onSearchInput() on every variable check -->
<sl-input ui-value="searchQuery?keypress" ui-event-keypress-enter="search()">
  <sl-icon slot="prefix" name="search"></sl-icon>
</sl-input>
<span class="hidden" ui-value="onSearchInput()"></span>
```

## When to Use

- Search-as-you-type against a local index or fast API
- Any incremental operation where the backend call blocks
- Cases where "latest value wins" — intermediate results are discarded

## When NOT to Use

- Non-blocking async operations (use debounce with `session:setTimeout`)
- Operations where intermediate results matter (use event compression)
- High-frequency continuous events like drag/scroll (use RAF loop — see `event-compression.md`)
