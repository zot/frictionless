# CRC Scripts

This directory contains utility scripts for the CRC modeling workflow in Python 3.6+.

## Available Scripts

### plantuml.py - Generate ASCII Sequence Diagrams

Generates ASCII sequence diagrams from PlantUML syntax.

```bash
python3 .claude/scripts/plantuml.py sequence "User -> View: click"
```

**Features:**
- Accepts PlantUML source via command-line argument or stdin
- Generates ASCII art sequence diagrams
- Supports all PlantUML sequence diagram syntax
- Auto-detects Java installation
- Cross-platform (Linux, macOS, Windows)

**Requirements:**
- Python 3.6+
- Java runtime
- PlantUML jar (`.claude/bin/plantuml.jar`)

### init-crc-project.py - Initialize CRC Modeling

Initializes a project with CRC modeling infrastructure.

```bash
python3 .claude/scripts/init-crc-project.py
```

**Features:**
- Creates `specs/` and `design/` directories
- Checks for required CRC agents
- Creates/updates `CLAUDE.md` with CRC workflow instructions
- Reports missing components
- Idempotent (safe to run multiple times)

**Checks for:**
- designer agent
- sequence-diagrammer agent
- test-designer agent
- gap-analyzer agent
- documenter agent
- plantuml.py script
- plantuml skill
- plantuml.jar

### trace-verify.py - Verify Traceability

Syncs checkboxes in traceability.md with actual code.

```bash
python3 .claude/scripts/trace-verify.py [phase-number]
```

**Features:**
- Checks for CRC: comments in files
- Automatically updates checkboxes in traceability.md
- Reports missing traceability comments
- Phase-based verification
- Exit code 0 if in sync, 1 if items need fixing

**Requirements:**
- Python 3.6+
- design/traceability.md file

## Usage Examples

### Generate Sequence Diagram

**Simple sequence:**
```bash
python3 .claude/scripts/plantuml.py sequence "User -> System: request"
```

**Multi-line sequence:**
```bash
python3 .claude/scripts/plantuml.py sequence "User -> Editor: save
alt valid
  Editor -> Storage: save()
else invalid
  Editor -> User: show errors
end"
```

**From stdin:**
```bash
echo "User -> System: request" | python3 .claude/scripts/plantuml.py sequence
```

### Initialize Project

**First-time setup:**
```bash
python3 .claude/scripts/init-crc-project.py
```

**Check status:**
Scripts are idempotent - running them again will check status without modifying existing content.

### Verify Traceability

```bash
python3 .claude/scripts/trace-verify.py [phase-number]
```

**Examples:**
```bash
# Verify Phase 1 (default)
python3 .claude/scripts/trace-verify.py

# Verify Phase 2
python3 .claude/scripts/trace-verify.py 2
```

**Features:**
- Syncs checkboxes in traceability.md with actual code
- Checks for CRC: comments in files
- Reports missing traceability comments
- Returns exit code 0 if all items in sync, 1 if items need fixing

## Installation

### PlantUML Jar

Download PlantUML jar file and place it at `.claude/bin/plantuml.jar`:

```bash
mkdir -p .claude/bin
curl -L -o .claude/bin/plantuml.jar https://github.com/plantuml/plantuml/releases/download/v1.2024.0/plantuml-1.2024.0.jar
```

### Make Scripts Executable

```bash
chmod +x .claude/scripts/*.py
```

## Troubleshooting

### "Java is not installed or not in PATH"

Install Java runtime:

```bash
# Ubuntu/Debian
sudo apt-get install default-jre

# macOS
brew install java

# Windows
# Download from: https://www.java.com/download/
```

### "PlantUML jar not found"

Download PlantUML jar as shown in Installation section above.

### Permission denied

Make scripts executable:

```bash
chmod +x .claude/scripts/plantuml.py
chmod +x .claude/scripts/init-crc-project.py
```

## Development

### Adding New Scripts

When adding new scripts:

1. Update this README
2. Add to init-crc-project checks if required
3. Make executable with `chmod +x`

### Testing

Test both versions:

```bash
python3 .claude/scripts/plantuml.py --help
python3 .claude/scripts/init-crc-project.py
```

## See Also

- `.claude/doc/crc.md` - CRC methodology documentation
- `.claude/agents/` - CRC workflow agents
- `CLAUDE.md` - Project-specific CRC instructions
