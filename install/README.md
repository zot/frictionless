# Frictionless

**Version: 0.17.3**

**An app ecosystem for Claude. Share the love—or steal it.**

Chat with Claude, get **fully hot-loadable apps**. Lua backend, HTML templates. No recompiles. No restarts.

And your apps can even **integrate with Claude**: your apps poke Claude and Claude pokes back. Right in the state.

Apps for

- **Quality of life** — tame complex tasks with forms and buttons
  - **Claude Code life** — point, click, Claude makes it so
  - **UNIX life** — UIs for UNIX tools
- **Life beyond Claude Code** — expenses, habits, projects, whatever
- **Dashboards** — surface info at a glance
- **Prototypes** — functional wireframes at a fraction of the tokens

What does **"fully" hot-loadable** mean?

- Both front-end changes and backend changes are hot-loadable.
- All your state is in the backend and hotloading preserves it.
- You rename a field of a prototype, all its instances' fields get renamed.
- Yeah even structural changes to your data. *That's* what **fully hot-loadable** means.

[![IMAGE ALT TEXT HERE](https://img.youtube.com/vi/Wd5n5fXoCuU/0.jpg)](https://youtu.be/Wd5n5fXoCuU)

## How It Works

Built on [ui-engine](https://github.com/zot/ui-engine). Less complexity → fewer tokens:

- **No API layer** — no endpoints, no serialization, no DTOs
- **No frontend code** — just HTML templates with declarative bindings
- **No sync wiring** — change backend data, UI updates automatically—no code to detect or push changes

Claude writes your app logic and skips everything else. See [overview](docs/OVERVIEW.md) for details.

## Usage

Once installed, use `/ui show` to show the Frictionless console. You can build apps in the console.

Frictionless uses your project's `.ui` directory for its apps and content.

### Building UIs in the CLI

Ask Claude to build a UI:

```
/ui-thorough make a contacts app with search and inline editing
```

Or display an existing app:

```
/ui show contacts
```

### Using the App Console

The app-console is your home base for managing Frictionless apps. Use `/ui show` to open it.

**Downloading apps from GitHub:**

Click the GitHub icon in the header to download apps directly from GitHub repositories.

![Download from GitHub](docs/images/download-from-github.jpg)

**Viewing app details:**

Select an app to see its requirements, open it, test it, or analyze it with Claude.

![App view](docs/images/app-view.jpg)

The bottom panel has two tabs:
- **Chat** — talk to Claude about the selected app
- **Lua** — run Lua code directly in your app's environment

### Standalone Mode

Run frictionless independently for development or testing changes to Frictionless itself:

```bash
frictionless serve --port 8000 --mcp-port 8001
frictionless serve --port 8000 --mcp-port 8001 --dir /path/to/ui-dir
```

The `--dir` option specifies the working directory for Lua scripts, viewdefs, and apps. Defaults to `.ui`.
The `--mcp-port` is only needed if you want to connect it to Claude.

### Bundling

Create custom binaries with your site embedded:

```bash
frictionless bundle site/ -o my-ui-dir   # Create bundled binary
frictionless ls                          # List bundled files
frictionless cat index.html              # Show file contents
frictionless cp '*.lua' scripts/         # Copy matching files
frictionless extract output/             # Extract all bundled files to current directory
```

## Available Apps

Download these from the app-console's GitHub panel:

| App | Description |
|-----|-------------|
| [Job&nbsp;Tracker](https://github.com/zot/frictionless/tree/main/apps/job-tracker) | Track job applications through the hiring pipeline. Paste a URL and Claude scrapes the details. |

## Documentation (in .ui by default)

- **[Intro](install/resources/intro.md)** — Introduction and overview
- **[Platform Reference](install/resources/reference.md)** — Architecture, tools, and quick start guide
- **[Viewdef Syntax](install/resources/viewdefs.md)** — HTML template bindings (`ui-*` attributes)
- **[Lua API](install/resources/lua.md)** — Class patterns and globals
- **[Agent Workflow](install/resources/mcp.md)** — Best practices for AI agents
- **[Themes](install/resources/themes.md)** — Theme switching and customization

## Installation

Tell Claude:

```
Install using github zot/frictionless readme
```

To install manually:

```bash
# Download (replace OS/ARCH: linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64)
mkdir -p {home}/.claude/bin
curl -L https://github.com/zot/frictionless/releases/latest/download/frictionless-linux-amd64 -o {home}/.claude/bin/frictionless
chmod +x {home}/.claude/bin/frictionless

cd {your-project}

# Add Frictionless to your project
claude mcp add frictionless -- {home}/.claude/bin/frictionless} mcp

# Initialize the project
{home}/.claude/bin/frictionless} install

# Add the playwright MCP server to your project for better debugging
claude mcp add playwright -- npx @playwright/mcp@latest
```

## Future Directions

### App Permissions

Lua apps have filesystem access via `io.open`, which allows reading files without Claude Code permission prompts. While convenient, this could be misused to explore directories the user hasn't explicitly shared.

**Planned guardrails:**
- Apps declare required permissions in a manifest (e.g., `app.json`)
- Permissions scoped to specific paths: `.claude/`, project root, etc.
- User approves permissions on first run
- Sandbox enforcement in the Lua environment
