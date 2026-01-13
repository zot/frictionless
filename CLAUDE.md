# Project Instructions

### Testing with the bundled binary
Always use `make build` before testing features that depend on bundled files (install, skills, agents, resources). The unbundled binary (`go build`) won't find these files.

```bash
make build                    # Creates build/ui-mcp with bundled files
./build/ui-mcp install        # Test install command
./build/ui-mcp mcp -vvvv      # Run MCP server
```

### Running the demo
From the project directory, this command runs the mcp `./build/ui-mcp mcp --port 8000 --dir .claude/ui -vvvv`
You can use the playwright browser to connect to it.

## 🎯 Core Principles
- Use **SOLID principles** in all implementations
- Create comprehensive **unit tests** for all components
- code and specs are as MINIMAL as POSSIBLE
- Before using a callback, see if a collaborator reference would be simpler

## When committing
1. Check git status and diff to analyze changes
2. Ask about any new files to ensure test/temp files aren't added accidentally
3. Add all changes (or only staged files if you specify "staged only")
4. Generate a clear commit message with terse bullet points
5. Create the commit and verify success

## Design Workflow

Use the mini-spec skill for all design and implementation work.

**3-level architecture:**
- `specs/` - Human specs (WHAT & WHY)
- `design/` - Design docs (HOW - architecture)
- `src/` - Implementation (code)

**Commands:**
- "design this" - generates design docs only
- "implement this" - writes code, updates Artifacts checkboxes
- After code changes - unchecks Artifacts, asks about design updates

**Design Entry Point:**
- `design/design.md` serves as the central tracking document
- Lists all CRC cards, sequences, and test designs with implementation status
- Tracks gaps between spec, design, and implementation

See `.claude/skills/mini-spec/SKILL.md` for the full methodology.

## Versioning and Releasing

Release versions use semantic versioning in `README.md` (the `**Version: X.Y.Z**` line near the top).

**To create a release:**
1. Update `**Version: X.Y.Z**` in both `README.md` and `install/README.md`
2. Commit: `git commit -am "Release vX.Y.Z"`
3. Tag: `git tag vX.Y.Z`
4. Build: `make release` (creates binaries in `release/` for Linux, macOS, Windows)
5. Push: `git push && git push --tags`
6. Create GitHub release: `gh release create vX.Y.Z release/* --title "vX.Y.Z" --notes "Release notes here"`
