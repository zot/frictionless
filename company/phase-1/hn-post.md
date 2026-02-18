# HN Post Draft

## Title

Show HN: A personal software ecosystem where Claude is literally the runtime

## URL

https://github.com/zot/frictionless

## Text (if self-post, otherwise this goes in first comment)

Frictionless is an open-source app platform for Claude Code. You chat with Claude to build apps and then Claude stays inside them as part of how they work.

There's no API layer, no app-specific frontend JavaScript, no sync wiring on the backend. Just Lua backend logic and HTML templates with declarative bindings. Claude writes the logic and skips the boilerplate.

The apps are fully hot-loadable, frontend and backend. Not just "swap a component," but structural changes to your data model while the app keeps running. Rename a field on a prototype, all instances update live.

And here why I built this: Claude doesn't just build the app and leave. It runs inside it. Your buttons can call Claude. Claude can update your app state. When you paste a LinkedIn job URL into the job tracker, Claude scrapes the posting, fills the form, web-searches for salary data and HQ address when they're not listed. All happening inside the running app. Claude can see what you have selected, make its own selections, and the app responds just as if you clicked.

**What it looks like in practice:**

- `/ui show` opens the app console in your system browser or in Playwright, if you prefer
- Click [+], name an app, describe what you want
- Claude builds it: requirements, design, Lua code, HTML templates
- The app appears. You use it. You chat to improve it. It hot-reloads.
- Ask Claude to change something. It reshapes the app live while it keeps running.
- Or paste a GitHub URL to download a community app. The app-console code scans this for security risks:
  - Shows counts per file and highlights dangerous calls like shell access
  - Presents them to you before Claude "sees" the files

**The architecture is intentionally weird:**

- Backend-only state. No frontend code required (but you can put JS in your templates if you need to).
- Implicit change tracking via inspection. No sync code needed.
  - Uses github/zot/change-tracker to detect changes
  - No decorators, no observers, no special setters
- Prototype-based OOP with structural hotload and instance mutation. Claude manages the proper mutation code.
- No API layer. Backend data *is* the UI state.
  - Presentation objects can represent UI state without polluting the domain
  - Template bindings can wrap domain objects with implicitly created presentation objects

These choices mean Claude writes dramatically less code per app, which means fewer tokens, faster builds, and fewer things to break.

**Available now:**

- Install: tell Claude to `Install using github zot/frictionless readme`
- Or manual install: download Frictionless release binary, `claude mcp add frictionless -- PATH/TO/FRICTIONLESS mcp`, `PATH/TO/FRICTIONLESS install`
- Ships with an app console for managing everything
  - a job tracker app is available in the same repository
    - in the app console, click the GitHub button and download `https://github.com/zot/frictionless/tree/main/apps/job-tracker`
- Open source, MIT license

Video demo: https://youtu.be/Wd5n5fXoCuU
Built on [ui-engine](https://github.com/zot/ui-engine).

I'm a solo developer who's been building this for a while. Happy to answer questions about the architecture, the Lua/HTML approach, or why I chose this over the obvious alternatives.

---

## First Comment (post immediately)

Hey HN, I'm the author. A few things that might come up:

**"Why Lua?"** Lua is tiny, embeddable, and fast. But the real reason: Lua tables with metatables give you prototype-based OOP that's perfect for hot-loadable state. You can mutate a prototype's structure and all instances update. Try doing that with a class-based language.

**"Why not just generate React/Next.js apps?"** Because then Claude has to write API endpoints, serialization, state management, component lifecycle, data fetching... all of which is boilerplate that burns tokens and introduces bugs. Strip all that out and Claude only writes the interesting parts.

**"How does Claude 'run inside' the app?"** The app can push events to Claude via the MCP protocol. Claude handles them and pushes state changes back. From the user's perspective, clicking a button can trigger Claude to do something smart (scrape a URL, analyze data, whatever) and the result appears in the app.

**"Why system browser by default?"** Privacy. With Playwright, Claude can see into the page. That's great for development, but it also means Claude has visibility into your browser. The system browser keeps Claude blind to the UI. Claude only interacts through the backend state, which is all it needs. Your other tabs stay private.

**"Is this just for Claude Code?"** Currently yes, it's an MCP server that connects to Claude Code. But the underlying engine (ui-engine) is standalone. You can run `frictionless serve` independently.
