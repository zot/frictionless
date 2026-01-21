# Frictionless Project Overview

## What is Frictionless?

**An app ecosystem for Claude.** Build your own Claude apps or download them:

- **Dashboards** — surface information at a glance
- **Command frontends** — tame complex UNIX tools with forms and buttons
- **Workflow tools** — common Claude usage patterns as clickable actions
- **Life beyond code** — expense tracking, habit building, project planning
- **Prototype production apps** — build functional wireframes at a fraction of the tokens

Build and modify apps while they run. No restarts, no rebuilds, no wait.

## Why Frictionless?

### The Problem with Traditional Web Development

Building a typical web app requires massive amounts of non-application code:
- API endpoints and OpenAPI specs
- Data fetching, serialization, DTOs
- State management (Redux, Vuex, etc.)
- Form handling and validation
- Real-time sync infrastructure

Your actual domain logic becomes a fraction of the codebase. This is especially painful with AI-assisted development: the AI spends most of its tokens on plumbing rather than your application.

### The Solution: eliminate the client/server boundary

**Claude writes app logic and skips the rest.**

Frictionless uses [ui-engine](https://github.com/zot/ui-engine) to eliminate complexity that eats tokens:

- **No API layer** — no endpoints, no serialization, no DTOs
- **No frontend code** — just HTML templates with declarative bindings
- **No sync wiring** — change backend data, UI updates automatically—no code to detect or push changes

Claude writes your app logic and skips everything else.

### Why This Matters for AI

Simpler architecture = fewer tokens = faster iteration:
- AI can focus on your application logic, not boilerplate
- Changes are localized to single files (Lua or HTML)
- Hot-reloading means instant feedback without restart cycles
- The entire app state is visible and debuggable in one place

## Design Decisions

### Why `.ui` instead of `.claude/ui`?

File permissions work better outside the `.claude` directory. The `.claude` directory has special handling that can interfere with normal file operations.

### Why Lua on the backend?

- **Dynamic and interactive** - Lua's flexibility allows for rapid iteration and live updates
- **In-place mutation** - You can rename a field in a prototype and it will update all instances automatically. This makes refactoring trivial and keeps the data model consistent without migration steps.
- **Hot-reload friendly** - Code changes apply instantly while preserving app state

### Why ui-engine instead of React?

- **Low cognitive overhead** - Simpler mental model for humans means fewer tokens when working with AI. Less boilerplate, less abstraction to explain.
- **No front-end code required** - UI is defined declaratively in HTML templates with bindings. The framework handles reactivity automatically.
- **Escape hatch available** - Custom JavaScript is supported when you need it, but rarely necessary for typical use cases.

## Desktop UIs Without the Complexity

Frictionless targets desktop/local UIs via stateful WebSocket connections. The presentation model lives on the backend, which enables automatic change detection without observer pattern boilerplate.

Traditional desktop UI approaches (Electron + React/Vue) require:
- Full frontend framework setup and bundling
- Observer pattern implementation on the backend
- API layer between backend and frontend
- State synchronization logic

This complexity exists even in Electron apps where the JS engine is directly connected to the DOM—developers still reach for React/Vue and all their associated boilerplate.

Frictionless eliminates all of this. The backend mutates objects directly; the framework handles synchronization automatically.

## Best For

- Rapid prototyping with AI assistance
- Desktop applications and internal tools
- Admin panels, dashboards, dev tools
- Apps where backend logic dominates
- Teams without frontend expertise

## Less Ideal For

- Internet-scale web apps with millions of concurrent users
- Offline-first applications
- Highly latency-sensitive UIs (games, real-time collaboration)
