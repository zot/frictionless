# Todo List

A minimal todo list app. Demonstrates text input, URL pasting, bookmarklet page capture, and persistent storage.

## Features

### Todo Items

1. **Add text todos** — type a description in the resizable textarea (4 rows) and click Add. Newlines are preserved.
2. **Toggle done** — click a checkbox to mark complete/incomplete
3. **Delete** — remove a todo with a delete button
4. **Persistent storage** — saved to `.ui/storage/example/todos.json`, loaded on startup

### URL Paste

1. **Paste a URL** into the input field
2. Claude detects it's a URL and fetches/summarizes the page
3. The todo appears **unsaved** (highlighted, with Save/Cancel buttons)
4. User can edit the description, then Save to keep it or Cancel to discard

### Bookmarklet

1. **Collapsible bookmarklet section** in the header — toggle via a small "bookmarklet" link
2. When expanded, shows a draggable "Add Todo" link for the bookmarks bar
3. Bookmarklet sends the current page's URL, title, and text to the publisher endpoint
4. Claude receives a `page_received` event, summarizes the content, and shows an unsaved todo (same as URL paste flow)

## Events

| Event | Payload | Action |
|-------|---------|--------|
| `chat` | `{text}` | If text is a URL, fetch and summarize into unsaved todo. If fetch fails, notify the user to try the bookmarklet instead. Otherwise ignore. |
| `page_received` | `{url, title, text}` | Summarize content into unsaved todo (no fetching needed — text is pre-rendered innerText). |
