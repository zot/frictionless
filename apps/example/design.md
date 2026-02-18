# Example (Todo List) - Design

## Intent

Minimal todo list demonstrating text input, persistent storage, bookmarklet page capture, and Claude-assisted URL summarization. Serves as the example app for the tutorial.

## Layout

```
+--------------------------------------------------+
|  Todo List  (bookmarklet)            panel-header |
+--------------------------------------------------+
|  Drag this to your bookmarks bar: [Add Todo]     | <- bookmarklet section (hidden by default)
+--------------------------------------------------+
|  [ Enter a todo or paste a URL...        ] [Add] | <- textarea (4 rows, resizable)
+--------------------------------------------------+
|  [x] Buy groceries                          [x]  | <- todo item (done)
|  [ ] Read the Lua manual                    [x]  | <- todo item
|  ================================================ |
|  [ ] Summary of pasted page — example.com   [x]  | <- unsaved draft (highlighted)
|      [Save] [Cancel]                              |
+--------------------------------------------------+
```

## Data Model

### Example

| Field | Type | Description |
|-------|------|-------------|
| _todos | TodoItem[] | List of saved todo items |
| _draft | TodoItem/nil | Unsaved draft from URL/bookmarklet (nil when none) |
| inputText | string | Current text in the input field |
| showBookmarklet | boolean | Whether bookmarklet section is visible |

### TodoItem

| Field | Type | Description |
|-------|------|-------------|
| text | string | Todo description |
| done | boolean | Completion state |
| url | string | Source URL if created from link (empty otherwise) |
| _parent | ref | Reference to parent Example for callbacks |

## Methods

### Example

| Method | Description |
|--------|-------------|
| todos() | Returns _todos for binding |
| draft() | Returns _draft for binding |
| addTodo() | Create TodoItem from inputText, clear input, save |
| addDraftTodo(text, url) | Create unsaved _draft TodoItem with text and url |
| saveDraft() | Move _draft into _todos, clear _draft, save |
| cancelDraft() | Clear _draft |
| noDraft() | Returns _draft == nil |
| removeTodo(todo) | Remove from _todos, save |
| save() | Write _todos to storage JSON |
| load() | Read _todos from storage JSON |
| toggleBookmarklet() | Toggle showBookmarklet |
| isBookmarkletHidden() | Returns not showBookmarklet |
| storagePath() | Returns .ui/storage/example/todos.json |

### TodoItem

| Method | Description |
|--------|-------------|
| toggle() | Flip done, call parent save |
| remove() | Call parent removeTodo(self) |
| label() | Returns text, appending " — domain" if url is present |

## ViewDefs

| File | Type | Purpose |
|------|------|---------|
| Example.DEFAULT.html | Example | Main panel: header, bookmarklet, input, todo list, draft |
| Example.TodoItem.list-item.html | TodoItem | Single todo row: checkbox, label, delete button |

## Events

### From UI to Claude

```json
{"app": "example", "event": "chat", "text": "https://example.com/article"}
{"app": "example", "event": "page_received", "url": "https://...", "title": "Page Title", "text": "...rendered page text..."}
```

### Claude Event Handling

| Event | Action |
|-------|--------|
| `chat` | If text is a URL, fetch the page content and create an unsaved draft via `example:addDraftTodo(summary, url)`. If fetch fails, call `mcp:notify("Couldn't fetch that page — try the bookmarklet instead", "warning")`. Non-URL chat is ignored. |
| `page_received` | Summarize the `text` field (already clean innerText, no fetching needed) into a short description. Create unsaved draft via `example:addDraftTodo(summary, url)`. |

For both events, the summary should be a concise one-line description of the page content suitable as a todo item.

## File I/O

### Storage

```
.ui/storage/example/
└── todos.json    # Array of {text, done, url}
```

Format:
```json
[
  {"text": "Buy groceries", "done": true, "url": ""},
  {"text": "Read Lua docs", "done": false, "url": ""},
  {"text": "Review article on CSS grid", "done": false, "url": "https://example.com/css-grid"}
]
```

On load, deserialize into TodoItem instances. On save, serialize _todos to JSON and write.

## Styling Notes

The app uses a `.todo-inner` wrapper (same pattern as prefs) to handle the MCP shell's `> div { height: 100% !important }` rule. The inner wrapper gets the flex column layout.

Draft items are visually distinct: highlighted background with accent border, and Save/Cancel buttons below the description.

The input is a `sl-textarea` (4 rows, vertically resizable). Todo item labels use `white-space: pre-wrap` to preserve newlines. Items align to `flex-start` for multi-line content.
