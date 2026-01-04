---
name: init-crc-project
description: Initialize CRC modeling in a project. Use when setting up the three-tier CRC system (specs, design, implementation) in a new or existing project.
---

# init-crc-project

Initialize CRC (Class-Responsibility-Collaboration) modeling in a project.

## What this skill does

This skill sets up the three-tier CRC modeling system in any project:
1. Creates directory structure (specs/, design/, .claude/agents/, etc.)
2. Updates CLAUDE.md with CRC workflow documentation
3. Validates required files (designer agent, plantuml setup)
4. Shows welcome message with next steps

## Approval Required

This command requires user approval before running (modifies project structure and CLAUDE.md).

Review the changes before approving.

## Commands

### init

Initialize CRC modeling in the current project.

**Usage:**
```bash
python3 ./.claude/scripts/init-crc-project.py
```

**What it does:**
- Creates `specs/` directory for Level 1 human-written specs
- Creates `design/` directory for Level 2 CRC cards, sequences, UI specs, architecture mapping
- Checks for designer agent (`.claude/agents/designer.md`)
- Checks for PlantUML setup (script, skill, jar)
- Updates `CLAUDE.md` with CRC workflow sections (if not present)
  - Includes architecture.md as design entry point
  - Includes diagnostic usage guidance
- Displays welcome message and documentation link

**Safe to run multiple times** - idempotent, won't duplicate content.

## Examples

### Initialize CRC in new project
```bash
python3 ./.claude/scripts/init-crc-project.py
```

Output:
```
🎯 Initializing CRC Modeling System...

✓ Created specs/ directory
✓ Created design/ directory
✓ Found designer agent
✓ Found PlantUML setup
✓ CLAUDE.md already has CRC sections

🎉 CRC Modeling initialized!

📚 Documentation: .claude/doc/crc.md
🚀 Next steps:
   1. Write Level 1 specs in specs/*.md
   2. Generate Level 2 designs: Task(subagent_type="designer", ...)
   3. Implement Level 3 code with traceability comments
```

## Requirements

To use this command, your project needs:
- `.claude/agents/designer.md` - Core CRC agent
- `.claude/scripts/plantuml.py` - Sequence diagram generator
- `.claude/skills/plantuml.md` - PlantUML skill
- `CLAUDE.md` - Project instructions file

## Documentation

See `.claude/doc/crc.md` for complete CRC modeling documentation.

## Implementation

See `.claude/scripts/init-crc-project.py` for the implementation script.
