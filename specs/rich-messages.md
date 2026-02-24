# Rich Messages

Language: Go, Lua
Environment: MCP server, ui-engine browser app

## Overview

Chat messages in the MCP shell support rich HTML content — markdown rendering for all messages and interactive highlight links for agent messages.

## Markdown Rendering

The MCP server provides a Lua-callable `mcp:renderMarkdown(text)` method that converts markdown text to an HTML fragment using goldmark (already a project dependency). This is the same renderer used for `/api/resource/` and static `.md` file serving, but returns a fragment (no page wrapper).

All chat messages (user and agent) render their `text` field as markdown into an `html` field at creation time. The raw `text` is preserved as source. The viewdef displays `html` when present, falling back to plain `text`.

## Rich Agent Messages

Agents can provide pre-built HTML directly via `mcp:addRichMessage(html)`. This bypasses markdown rendering and injects raw HTML into the chat — used for interactive content like highlight links that can't be expressed in markdown.

A helper `mcp:highlightLink(elementId, label)` constructs an anchor tag that calls `window.uiApp.highlight()` on click, producing inline "click here to see it" links in agent messages.

## Security

Goldmark escapes raw HTML in markdown input by default (no `html.WithUnsafe()`). Agent-provided raw HTML via `addRichMessage` is trusted — the agent is the code author.
