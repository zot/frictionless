#!/usr/bin/env python3
"""
init-crc-project.py - Initialize CRC modeling in a project
"""
import sys
import os
from pathlib import Path

# ANSI color codes
class Colors:
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    NC = '\033[0m'  # No Color

def get_paths():
    """Get current working directory as project root."""
    project_root = Path.cwd()
    return project_root

def create_directories(project_root):
    """Create specs/ and design/ directories."""
    print("📁 Setting up directory structure...")

    specs_dir = project_root / 'specs'
    design_dir = project_root / 'design'

    if not specs_dir.exists():
        specs_dir.mkdir(parents=True)
        print(f"{Colors.GREEN}✓{Colors.NC} Created specs/ directory")
    else:
        print(f"{Colors.BLUE}→{Colors.NC} specs/ directory already exists")

    if not design_dir.exists():
        design_dir.mkdir(parents=True)
        print(f"{Colors.GREEN}✓{Colors.NC} Created design/ directory")
    else:
        print(f"{Colors.BLUE}→{Colors.NC} design/ directory already exists")

    print()

def check_required_components(project_root):
    """Check for required files and return list of missing components."""
    print("🔍 Checking for required components...")

    missing_components = []

    components = [
        ('.claude/agents/designer.md', 'designer agent', True),
        ('.claude/agents/design-maintainer.md', 'design-maintainer agent', True),
        ('.claude/agents/sequence-diagrammer.md', 'sequence-diagrammer agent', True),
        ('.claude/agents/test-designer.md', 'test-designer agent', True),
        ('.claude/agents/gap-analyzer.md', 'gap-analyzer agent', True),
        ('.claude/agents/documenter.md', 'documenter agent', True),
        ('.claude/scripts/plantuml.py', 'plantuml.py script', True),
        ('.claude/skills/plantuml.md', 'plantuml skill', True),
        ('.claude/bin/plantuml.jar', 'plantuml.jar', False),
    ]

    for path, name, is_agent in components:
        full_path = project_root / path
        if not full_path.is_file():
            if path.endswith('plantuml.jar'):
                print(f"{Colors.YELLOW}⚠{Colors.NC} Missing {name} ({path})")
                print(f"   {Colors.BLUE}→{Colors.NC} Download from: https://plantuml.com/download")
            else:
                print(f"{Colors.YELLOW}⚠{Colors.NC} Missing {name} ({path})")
            missing_components.append(name)
        else:
            if is_agent:
                print(f"{Colors.GREEN}✓{Colors.NC} Found {name}")
            else:
                print(f"{Colors.GREEN}✓{Colors.NC} Found {name}")

    print()
    return missing_components

def check_claude_md(project_root):
    """Check CLAUDE.md and add CRC sections if needed."""
    print("📝 Checking CLAUDE.md...")

    claude_md = project_root / 'CLAUDE.md'

    # Check if CLAUDE.md already has CRC content
    if claude_md.is_file():
        with open(claude_md, 'r') as f:
            content = f.read()
            if 'three-tier system' in content:
                print(f"{Colors.BLUE}→{Colors.NC} CLAUDE.md already has CRC sections")
                return

    # CLAUDE.md missing CRC content (or doesn't exist)
    if not claude_md.is_file():
        print(f"{Colors.YELLOW}⚠{Colors.NC} CLAUDE.md not found in project root")
        print(f"   {Colors.BLUE}→{Colors.NC} Creating CLAUDE.md with CRC sections...")
    else:
        print(f"{Colors.YELLOW}⚠{Colors.NC} CLAUDE.md exists but missing CRC sections")
        print(f"   {Colors.BLUE}→{Colors.NC} Appending CRC workflow sections...")

    # CRC sections content
    crc_content = '''# Project Instructions

## CRC Modeling Workflow

**DO NOT generate code directly from `specs/*.md` files!**

**Use a three-tier system:**
```
Level 1: Human specs (specs/*.md)
   ↓
Level 2: Design models (design/*.md) ← CREATE THESE FIRST
   ↓
Level 3: Implementation (source code)
```

**Workflow:**
1. Read human specs (`specs/*.md`) for design intent
2. Use `designer` agent to create Level 2 specs (CRC cards, sequences, UI specs, architecture mapping, **and test designs**)
   - Designer agent MUST invoke test-designer sub-agent (automatic, mandatory step)
   - Verify test design files (`design/test-*.md`) are created before proceeding
3. Generate code following complete specification with traceability comments

**When Designer Agent is Required vs Direct CRC Creation:**

| Scenario | Use Designer Agent? | Required Follow-up |
|----------|---------------------|-------------------|
| New feature design | YES | Full workflow (sequences, test designs, gap analysis) |
| Significant architectural change | YES | Full workflow |
| Documenting existing code | Optional | Run gap-analyzer to verify completeness |
| Fixing/cleaning up CRC cards | No | Verify sequence references exist |
| Creating CRC for existing interface | Optional | Run gap-analyzer to verify completeness |

**CRITICAL: Regardless of how CRC cards are created:**
1. All sequence references must point to existing files (fix or create)
2. Non-trivial "Does" behaviors need sequence diagrams
3. Run `gap-analyzer` agent after creating/modifying CRC cards
4. Update `design/traceability.md` and `design/architecture.md`

**Design Entry Point:**
- `design/architecture.md` serves as the "main program" for the design
- Shows how design elements are organized into logical systems
- Start here to understand the overall architecture
- **Use for problem diagnosis and impact analysis** - quickly localize issues and assess change scope

**When to Read architecture.md:**
- **When working with design files, implementing features, or diagnosing issues, always read `design/architecture.md` first to understand the system structure and component relationships.**

**Traceability Comment Format:**
- Use simple filenames WITHOUT directory paths
- ✅ Correct: `CRC: crc-Person.md`, `Spec: main.md`, `Sequence: seq-create-user.md`
- ❌ Wrong: `CRC: design/crc-Person.md`, `Spec: specs/main.md`

**Finding Implementations:**
- To find where a design element is implemented, grep for its filename (e.g., `grep "seq-get-file.md"`)

**Test Implementation:**
- **Test designs are Level 2 artifacts**: Designer agent automatically generates test design specs (`design/test-*.md`) via the test-designer sub-agent
- **ALWAYS read test designs BEFORE writing test code**: Test designs specify what to test, test code implements those specifications
- **Test code MUST implement all scenarios from test designs**: Every test scenario in `design/test-*.md` must have corresponding test code
- **Traceability**: Test files reference test designs in comments: `// Test Design: test-ComponentName.md`
- Test files belong in top-level `tests/` directory (NOT nested under `src/`)
- When configuring build tools (Vite, Webpack, etc.), ensure test runner configurations are separate from application build configurations
- If build config sets a custom `root` directory, create a separate test configuration file to avoid test discovery issues
- Run `npm test` to verify test discovery works correctly before considering tests complete

**Test Design Workflow:**
1. Designer agent creates CRC cards and sequences (Level 2)
2. Designer agent invokes test-designer agent (automatic, mandatory)
3. Test-designer generates test design specs (`design/test-*.md`)
4. Read test designs to understand what needs testing
5. Implement tests following test design specifications
6. Reference test designs in test code comments

See `.claude/doc/crc.md` for complete documentation.

### 🔄 Bidirectional Traceability Principle

**When changes occur at any level, propagate updates through the documentation hierarchy:**

**Source Code Changes → Design Specs:**
- Modified implementation → Update CRC cards/sequences/UI specs if structure/behavior changed
- New classes/methods → Create corresponding CRC cards
- Changed interactions → Update sequence diagrams
- Template/view changes → Update UI specs

**Use the `design-maintainer` agent to automate this:**
```
When you've made code changes, invoke the design-maintainer agent to:
- Update CRC cards with new methods/fields
- Update sequence diagrams for changed workflows
- Add traceability comments to new code
- Check off traceability.md checkboxes
```

**Design Spec Changes → Architectural Specs:**
- Modified CRC cards/sequences → Update high-level specs if requirements/architecture affected
- New components → Document in feature specs and update `design/architecture.md`
- Changed workflows → Update architectural documentation
- System reorganization → Update `design/architecture.md` to reflect new system boundaries

**Key Rules:**
1. **Always update up**: When code/design changes, ripple changes upward through documentation
2. **Maintain abstraction**: Each level documents at its appropriate abstraction
3. **Keep consistency**: All three tiers must tell the same story at their respective levels
4. **Update traceability comments**: When docs change, update CRC/spec references in code comments

**Agent Workflow:**
- **Requirements → Design**: Use `designer` agent (Level 1 → Level 2)
- **Code → Design**: Use `design-maintainer` agent (Level 3 → Level 2)
- **Design → Documentation**: Use `documenter` agent (Level 2 → Docs)

### 🔧 Design Update Requests

**When the user asks to update, modify, or add to the design (Level 2 artifacts), ALWAYS use the appropriate agent:**

| User Request | Agent to Use |
|--------------|--------------|
| "Update the design for X" | `designer` |
| "Add X to the design" | `designer` |
| "Reflect spec changes in design" | `designer` |
| "Update CRC cards / sequences" | `designer` |
| "Update design based on these changes" | `designer` |
| "Update design after code changes" | `design-maintainer` |
| "Run gap analysis" | `gap-analyzer` |
| "Generate/update documentation" | `documenter` |

**Do NOT manually edit design files** unless it's a trivial fix (typo, formatting). Always delegate to the appropriate agent to ensure:
- Consistency across CRC cards, sequences, and architecture
- Proper traceability updates
- Test design updates when needed

### 📚 Documentation Generation

**After completing design or implementation work, offer to generate or update project documentation.**

Use the `documenter` agent to create:
- `docs/requirements.md` - Requirements documentation from specs
- `docs/design.md` - Design overview from CRC cards and sequences
- `docs/developer-guide.md` - Developer documentation with architecture and setup
- `docs/user-manual.md` - User manual with features and how-to guides
- `design/traceability-docs.md` - Documentation traceability map

**When to offer documentation generation:**
- After creating/updating Level 2 design specs
- After implementing Level 3 code
- When specs or design changes significantly
- When user explicitly requests it

**Example offer:**
"I've completed the [design/implementation]. Would you like me to generate/update the project documentation (requirements, design overview, developer guide, and user manual)?"
'''

    # Append CRC sections (creates file if doesn't exist)
    with open(claude_md, 'a') as f:
        f.write(crc_content)

    print(f"{Colors.GREEN}✓{Colors.NC} Added CRC sections to CLAUDE.md")

def print_summary(missing_components):
    """Print summary of initialization."""
    print()
    print("━" * 80)
    print()

    if not missing_components:
        print(f"{Colors.GREEN}🎉 CRC Modeling initialized successfully!{Colors.NC}")
    else:
        print(f"{Colors.YELLOW}⚠ CRC Modeling partially initialized{Colors.NC}")
        print()
        print("Missing components:")
        for component in missing_components:
            print(f"  - {component}")
        print()
        print("See .claude/doc/crc.md for setup instructions")

    print()
    print(f"{Colors.BLUE}📚 Documentation:{Colors.NC} .claude/doc/crc.md")
    print()
    print(f"{Colors.BLUE}🚀 Next steps:{Colors.NC}")
    print("   1. Write Level 1 specs in specs/*.md")
    print("   2. Generate Level 2 designs: Task(subagent_type=\"designer\", ...)")
    print("   3. Implement Level 3 code with traceability comments")
    print()
    print("━" * 80)

def main():
    """Main initialization function."""
    print("🎯 Initializing CRC Modeling System...")
    print()

    project_root = get_paths()

    # Create directories
    create_directories(project_root)

    # Check for required components
    missing_components = check_required_components(project_root)

    # Check and update CLAUDE.md
    check_claude_md(project_root)

    # Print summary
    print_summary(missing_components)

if __name__ == '__main__':
    main()
