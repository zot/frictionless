# Minispec Configuration Reference

## Config File

**Path:** `.minispec.yaml` in the project root (next to the `design/` directory).
**Format:** YAML. All fields are optional — omitted fields keep built-in defaults.

## Schema

```yaml
design_dir: design              # Path to design directory (relative to project root)
src_dir: src                    # Path to source directory (relative to project root)
code_extensions:                # File extensions to scan for traceability comments
  - .go
  - .ts
  - .pas
comment_patterns:               # Comment-prefix regex per file extension
  .go: "//\\s*"
  .pas: "\\{\\s*"
comment_closers:                # Closing delimiter for block-comment languages
  .pas: " }"
  .css: " */"
  .md: " -->"
  .html: " -->"
```

## Comment Patterns

Each entry maps a file extension to a **regex matching the comment prefix only**. The tool appends `CRC:\s*...` to build the full traceability regex.

**Built-in defaults:**

| Extension | Pattern | Style |
|-----------|---------|-------|
| `.go` | `//\s*` | C-style single-line |
| `.js` | `//\s*` | C-style single-line |
| `.ts` | `//\s*` | C-style single-line |
| `.c` | `//\s*` | C-style single-line |
| `.h` | `//\s*` | C-style single-line |
| `.cpp` | `//\s*` | C-style single-line |
| `.py` | `#\s*` | Hash |
| `.lua` | `--\s*` | Double-dash |
| `.sh` | `#\s*` | Hash |
| `.bash` | `#\s*` | Hash |
| `.md` | `<!--\s*` | HTML comment |
| `.html` | `<!--\s*` | HTML comment |
| `.css` | `/\*\s*` | Block comment |

**Adding a new language** — put an entry in `comment_patterns`. User entries override defaults for matching extensions; unmatched extensions keep built-in patterns.

## Comment Closers

**CRITICAL for block-comment languages.** Languages like Pascal (`{...}`), CSS (`/*...*/`), and HTML (`<!--...-->`) require a closing delimiter. An unclosed comment silently swallows all subsequent code until the next accidental closer — this is catastrophic and extremely hard to diagnose.

`comment_closers` maps file extensions to the string that **must** be appended when writing traceability comments. The tool's parser already strips these terminators when reading; this field tells the AI what to emit when writing.

**Built-in languages with line-terminating comments** (Go, Python, JS, Lua, shell) do not need closers — the comment ends at the newline.

**Languages that REQUIRE closers:**

| Extension | Closer | Comment form |
|-----------|--------|--------------|
| `.pas` | ` }` | `{ CRC: ... }` |
| `.dpr` | ` }` | `{ CRC: ... }` |
| `.css` | ` */` | `/* CRC: ... */` |
| `.md` | ` -->` | `<!-- CRC: ... -->` |
| `.html` | ` -->` | `<!-- CRC: ... -->` |

When writing traceability comments, **always check `comment_closers`** for the file extension. If a closer exists, append it. If it doesn't, the comment is line-terminated and needs no closer.

### Example: Delphi/Pascal

Config:
```yaml
comment_patterns:
  .pas: "\\{\\s*"
  .dpr: "\\{\\s*"
comment_closers:
  .pas: " }"
  .dpr: " }"
```

Correct:
```pascal
{ CRC: crc-Parser.md | Seq: seq-compile.md }
procedure Parse;
```

Wrong — **will comment out everything below:**
```pascal
{ CRC: crc-Parser.md | Seq: seq-compile.md
procedure Parse;
```

## Precedence

1. **CLI flags** (`--design-dir`, `--src-dir`) override everything
2. **`.minispec.yaml`** overrides built-in defaults
3. **Built-in defaults** apply when no override exists

## Verifying

After creating or updating `.minispec.yaml`, run:

```bash
~/.claude/bin/minispec query comment-patterns
```

Custom patterns appear alongside built-ins in the output.
