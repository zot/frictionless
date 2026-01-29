# Project Instructions

Use the `/ui` skill to run frictionless UIs.
Use the `/ui-builder` skill whenever you need to explore, change, or audit specs, design, or code for **frictionless UIs**.
### mini-spec
Whenever you need to explore specs, design, or code, use `design/design.md` with mini-spec.
Use mini-spec and it's phased approach when creating or altering specs, design, or code.
`design/design.md ` contains a map of the project's non-ui code.

### Testing with the bundled binary
Always use `make build` before testing features that depend on bundled files (install, skills, agents, resources). The unbundled binary (`go build`) won't find these files.

```bash
make build                    # Creates build/frictionless with bundled files
./build/frictionless install        # Test install command
./build/frictionless mcp -vvvv      # Run MCP server
```

### Running the demo
From the project directory, this command runs the mcp `./build/frictionless mcp --port 8000 --dir .ui -vvvv`
You can use the playwright browser to connect to it.

## When committing
1. Check git status and diff to analyze changes
2. Ask about any new files to ensure test/temp files aren't added accidentally
3. Add all changes (or only staged files if you specify "staged only")
4. Generate a clear commit message with terse bullet points
5. Create the commit and verify success

## Versioning and Releasing

Release versions use semantic versioning in `README.md` (the `**Version: X.Y.Z**` line near the top).

**To create a release:**
1. Check if [ui-engine](https://github.com/zot/ui-engine) has a newer version and update `go.mod` if needed (`go get github.com/zot/ui-engine@latest`)
2. Update `**Version: X.Y.Z**` in both `README.md` and `install/README.md`
3. Commit: `git commit -am "Release vX.Y.Z"`
4. Tag: `git tag vX.Y.Z`
5. Build: `make release` (creates binaries in `release/` for Linux, macOS, Windows)
6. Push: `git push && git push --tags`
7. Create GitHub release: `gh release create vX.Y.Z release/* --title "vX.Y.Z" --notes "Release notes here"`
