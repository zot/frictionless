# CRC Modeling Documentation

Complete guide to Class-Responsibility-Collaboration (CRC) modeling for software projects.

## Table of Contents

1. [What is CRC Modeling?](#what-is-crc-modeling)
2. [Why CRC Modeling?](#why-crc-modeling)
3. [Quick Start](#quick-start)
4. [Three-Tier System](#three-tier-system)
5. [Directory Structure](#directory-structure)
6. [Required Components](#required-components)
7. [Workflow](#workflow)
8. [CRC Cards](#crc-cards)
9. [Sequence Diagrams](#sequence-diagrams)
10. [UI Specifications](#ui-specifications)
11. [Architecture Mapping](#architecture-mapping)
12. [Traceability](#traceability)
13. [Bidirectional Updates](#bidirectional-updates)
14. [Benefits](#benefits)
15. [Reverse Engineering Existing Projects](#reverse-engineering-existing-projects)

---

## What is CRC Modeling?

**CRC (Class-Responsibility-Collaboration)** is a design methodology that creates an intermediate layer between human requirements and code implementation.

**What CRC produces:**
- **Classes** - Objects/components that make up your system
- **Responsibilities** - What each class knows (data) and does (behavior)
- **Collaborations** - How classes work together to fulfill requirements
- **Sequence diagrams** - Object interactions over time for specific scenarios
- **UI specifications** - Layout structure (ASCII art), data bindings, event handlers
- **Traceability maps** - Links connecting requirements → design → code

**Core principles:**
- **Explicit design phase** prevents architectural problems and ensures complete specifications
- **Traceability** is key for iterative development —- bidirectional links between specs/design/code enable safe evolution and refactoring

**Three-tier system:**
```
Level 1: Human specs (specs/*.md)        ← Requirements, intent, "WHAT" and "WHY"
   ↓
Level 2: Design models (design/*.md)     ← Structure, behavior, "HOW" at design level
   ↓
Level 3: Implementation (source code)    ← Concrete code, "HOW" at implementation level
```

---

## Why CRC Modeling?

### Solve the Human-LLM Communication Problem

**The killer advantage:** Level 2 reveals how the LLM interprets your specs—in a format you can review before any code is written.

**The fundamental problem with vibe-coding and simple, two-level spec-driven development:**

When you write requirements and an LLM generates code directly, you have a **communication gap**:
- Did the LLM understand your intent correctly?
- What assumptions did it make?
- What did it infer that you didn't explicitly state?
- Did it misinterpret ambiguous requirements?

**You only discover these issues after reviewing generated code**—which requires reading implementation details, tracing through logic, and mentally reconstructing the design the LLM had in mind.

**Level 2 solves this communication problem:**

The middle layer (CRC cards, sequence diagrams, UI specs) forces the LLM to **explicitly document its interpretation** of your requirements in **human-readable design artifacts**:

- **CRC cards** show what classes/objects it plans to create and their responsibilities
- **Sequence diagrams** show how it thinks objects should interact for each scenario
- **UI specs** show the layout structure it inferred from your descriptions

**This lets you verify the LLM's understanding BEFORE implementation:**
- Review designs in minutes instead of code in hours
- Catch misinterpretations when they're cheap to fix (markdown, not code)
- Provide feedback on architecture while it's still abstract
- Ensure you and the LLM agree on the approach before coding begins
- **Establish patterns for future work** - Level 2 specs preserve design intent and provide architectural patterns the LLM will follow in future sessions and changes

**In short:** Level 2 is a **human-readable translation layer** between your intent and the code. It makes the LLM's interpretation explicit and reviewable, solving the core communication problem that plagues both vibe-coding and naive spec-driven approaches.

### Discover Design Problems Before Coding

**The central advantage:** You find architectural flaws in markdown, not in code.

**The three-tier process forces you to think through:**
- What classes/objects you need (CRC cards)
- How they interact for each scenario (sequence diagrams)
- What edge cases and error conditions exist (sequences catch these)
- Where responsibilities belong (SOLID emerges naturally)

### Without Level 2 (jumping specs → code):

❌ You discover missing requirements halfway through implementation
❌ Architectural problems appear after significant code is written
❌ Edge cases surprise you during testing
❌ Refactoring is expensive and risky
❌ "God classes" emerge organically
❌ Integration problems discovered late

### With Level 2 (specs → design → code):

✅ **Sequences expose missing requirements** - "Wait, what happens if the peer is offline?"
✅ **CRC cards expose god classes** - "This class has 15 responsibilities, we need to split it"
✅ **Design reviews happen before coding** - Problems caught in markdown, not code
✅ **Complete specification exists** - No guessing during implementation
✅ **Edge cases documented upfront** - Sequences force you to think through error paths
✅ **Collaboration issues visible** - Circular dependencies spotted in design phase

### The Designer Agent Makes It Practical

The designer agent automates creating Level 2 specs, making thorough design practical rather than theoretical. You get the benefits of careful architecture without the tedious manual documentation work.

**In short:** Level 2 is cheap error detection. Finding design flaws in markdown is **orders of magnitude cheaper** than finding them in code.

### Traceability Enables Long-Term Maintainability

**The maintainability advantage:** Explicit links between specs → design → code → tests → documentation create a navigable map of your entire system.

**The three-tier process creates bidirectional traceability:**

Each level references the others through explicit links:
- **Specs reference design artifacts** - "See design/crc-FriendsManager.md for architecture"
- **Design artifacts reference specs** - "Source Spec: specs/friends.md"
- **Code references design** - Traceability comments link implementation to CRC cards and sequences
- **Tests reference design** - Test files cite the CRC cards and sequences they validate

**This traceability facilitates maintainability across the development lifecycle:**

✅ **Impact analysis** - "This requirement changed → which designs are affected → which code → which tests"
- Find all code implementing a specific requirement in seconds
- Trace from bug report → test → code → design → spec to understand root cause
- Identify what needs updating when requirements evolve

✅ **Bidirectional updates** - Changes propagate through documentation hierarchy
- Code changes trigger design spec updates when structure/behavior changes
- Design changes trigger high-level spec updates when architecture evolves
- All three tiers stay synchronized and tell the same story

✅ **Cross-session stability** - Explicit design documentation provides consistency over time
- Level 2 specs preserve architectural intent when resuming work after breaks
- Design patterns guide future decisions, ensuring consistency with established approaches
- LLMs can reference design artifacts to maintain architectural coherence across sessions
- Works hand-in-hand with traceability to prevent architectural drift

✅ **Onboarding and knowledge transfer** - New developers navigate the system at multiple abstraction levels
- Start with high-level specs to understand "what" and "why"
- Review CRC cards and sequences to understand "how" at design level
- Dive into code with context from design artifacts
- Traceability comments guide exploration

✅ **Audit trail** - Document architectural decisions and evolution
- See why each class exists and what requirements it fulfills
- Understand how features evolved over time
- Justify design choices to stakeholders and future maintainers

✅ **Refactoring confidence** - Complete specification guides safe restructuring
- Sequences document all expected interactions (test scenarios)
- CRC cards define clear responsibilities (refactoring boundaries)
- Impact analysis prevents breaking changes

**In short:** Traceability transforms your codebase from a black box into a **documented, navigable system** where every line of code traces back to requirements and design decisions. This makes maintenance, debugging, and evolution dramatically easier.

**Note:** You can apply CRC modeling to existing projects through reverse engineering—LLMs can analyze code to extract Level 2 designs, then generate Level 1 specs. See [Reverse Engineering Existing Projects](#reverse-engineering-existing-projects) for the complete workflow.

---

## Quick Start

### Initialize CRC in Your Project

The install command will:
- Create `specs/` and `design/` directories
- Check for required components (designer agent, PlantUML)
- Update CLAUDE.md with CRC workflow sections
- Show next steps

#### Using slash command (recommended)
```
/init-crc-project
```

#### Python command
```bash
# Or run the script directly
python3 ./.claude/scripts/init-crc-project.py
```

### Basic Workflow

1. **Write specs** in `specs/*.md` (human-readable requirements)
2. **Generate designs** using designer agent:
   ```
   Task(subagent_type="designer", prompt="Generate Level 2 specs for specs/feature.md")
   ```
3. **Generate test designs** (automatic via designer agent):
   ```
   # Test designs are automatically generated by the designer agent
   # No separate step needed
   ```
4. **Implement code** following CRC cards and sequences, adding traceability comments
5. **Implement tests** following test designs (if generated)

---

## Three-Tier System

### Level 1: Human Specs (`specs/*.md`)

**Purpose:** Document WHAT needs to be built and WHY

**Content:**
- Requirements and user stories
- Architecture and design intent
- UX flows and interaction patterns
- Business logic and rules
- Principles and constraints

**Format:** Markdown, human-readable, focuses on intent

**Example:** `specs/friends.md` describes friend management requirements

### Level 2: Design Models (`design/*.md`)

**Purpose:** Document HOW it will be structured (design level)

**Content:**
- CRC cards: Classes, responsibilities, collaborators
- Sequence diagrams: Object interactions over time
- UI specs: Layout structure, data bindings, events
- Test designs: Test specifications derived from CRC cards and sequences
- Traceability map: Links between all levels
- Gap analysis: What's missing or ambiguous

**Format:** Structured markdown with ASCII diagrams

**Example:** `design/crc-FriendsManager.md` defines the FriendsManager class design

### Level 3: Implementation (Source Code)

**Purpose:** Concrete implementation (code level)

**Content:**
- Classes, methods, functions
- Templates, views, components
- Tests and documentation
- Traceability comments linking to Level 2

**Format:** Your programming language

**Example:** `src/FriendsManager.ts` implements the design from the CRC card

---

## Directory Structure

```
project-root/
├── .claude/
│   ├── agents/
│   │   ├── designer.md          # Level 2 spec generator (REQUIRED)
│   │   ├── sequence-diagrammer.md # PlantUML ASCII converter (REQUIRED)
│   │   ├── test-designer.md     # Test design generator (REQUIRED)
│   │   └── gap-analyzer.md      # Gap analysis (REQUIRED)
│   ├── scripts/
│   │   ├── init-crc-project.py  # Initialization command
│   │   └── plantuml.py          # PlantUML ASCII art generator
│   ├── skills/
│   │   ├── init-crc-project.md  # Initialization skill
│   │   └── plantuml.md          # PlantUML skill documentation
│   ├── doc/
│   │   └── crc.md               # This file
│   └── bin/
│       └── plantuml.jar         # PlantUML executable
├── specs/                        # Level 1: Human-written specs
│   ├── feature1.md
│   └── feature2.md
├── design/                       # Level 2: Generated CRC/sequences/UI specs
│   ├── architecture.md          # Architecture mapping - ENTRY POINT (systems & cross-cutting)
│   ├── crc-ClassName.md         # CRC cards (one per class)
│   ├── seq-scenario.md          # Sequence diagrams (one per scenario)
│   ├── ui-view-name.md          # UI layout specs (one per view)
│   ├── test-ComponentName.md    # Test designs (one per component/feature)
│   ├── manifest-ui.md           # Global UI concerns
│   ├── traceability.md          # Formal traceability map
│   └── gaps.md                  # Gap analysis
└── [source-code]/                # Level 3: Implementation
    └── ...                       # (src/, lib/, app/, pkg/, etc.)
```

**Note:** Source code directory name varies by project. Adapt to your conventions.

---

## Required Components

### 1. Designer Agent

**File:** `.claude/agents/designer.md`

**Purpose:** Generates Level 2 design specs from Level 1 human specs

**What it creates:**
- CRC cards for all classes
- Sequence diagrams for all scenarios
- UI layout specifications
- Traceability map
- Gap analysis

**Usage:**
```
Task(
  subagent_type="designer",
  prompt="Generate Level 2 specs for specs/feature.md. Output to design/"
)
```

### 2. Sequence Diagrammer Agent (Optional)

**File:** `.claude/agents/sequence-diagrammer.md`

**Purpose:** Converts sequence diagrams to PlantUML ASCII format

**What it does:**
- Reads existing sequence diagram markdown files
- Extracts or creates PlantUML source from diagram descriptions
- Generates ASCII art diagrams using PlantUML
- Updates files with properly formatted diagrams
- Preserves all metadata and analysis sections

**Usage:**
```
Task(
  subagent_type="sequence-diagrammer",
  description="Convert sequence diagrams to PlantUML ASCII",
  prompt="Convert design/seq-*.md files to PlantUML ASCII format.

  Process:
  1. Read the existing files
  2. For each diagram section, create PlantUML source
  3. Use Skill tool to invoke plantuml skill
  4. Update files with generated ASCII art
  5. Preserve all metadata and analysis sections"
)
```

**Note:** This agent is invoked separately after designer creates sequence diagrams. It converts plain text diagrams to properly formatted PlantUML ASCII art.

### 3. Test Designer Agent (Required)

**File:** `.claude/agents/test-designer.md`

**Purpose:** Generates Level 2 test design specs from CRC cards and sequence diagrams

**What it creates:**
- Test design documents (`design/test-*.md`)
- Test cases with name, purpose, input, and expected results
- Coverage analysis of CRC responsibilities and sequences
- Test specifications in human-readable English (not code)

**Usage:**
```
# Automatically invoked by designer agent as step 11
# No separate invocation needed
```

**Note:** Test designs are Level 2 artifacts (test specifications), not Level 3 (test code). They guide the implementation of actual tests. The designer agent automatically invokes this agent after creating CRC cards and sequences.

### 3. PlantUML Setup

**Required files:**
- `.claude/scripts/plantuml.py` - Wrapper script
- `.claude/skills/plantuml.md` - Skill documentation
- `.claude/bin/plantuml.jar` - PlantUML executable (download from https://plantuml.com/download)
- Java runtime (required by PlantUML)
- Python 3.6+ (for plantuml.py script)

**Usage:** Script generates ASCII sequence diagrams from PlantUML syntax

---

## Workflow

### Step 1: Write Level 1 Specs

Create human-readable specifications in `specs/*.md`:

```markdown
# Friends Feature

## Requirements

Users can add friends by peer ID, view friend status, and manage their friends list.

## User Stories

1. As a user, I want to add a friend by entering their peer ID
2. As a user, I want to see which friends are online
3. As a user, I want to remove friends I no longer connect with

## Data

- Friend: name, peerId, notes, pending status
- Friends list stored in LocalStorage
```

### Step 2: Generate Level 2 Designs

Use the designer agent:

```
Task(
  subagent_type="designer",
  description="Generate design specs for friends",
  prompt="Generate complete Level 2 design specifications for specs/friends.md.

  Output to design/ directory.

  Create:
  - CRC cards for all classes
  - Sequence diagrams for all scenarios
  - UI layout specifications
  - Traceability map
  - Gap analysis"
)
```

**Output:** CRC cards, sequences, UI specs, traceability map, gaps document

### Step 3: Review Generated Designs

Check generated files:
- `design/crc-*.md` - Do classes have clear responsibilities?
- `design/seq-*.md` - Do sequences cover all scenarios?
- `design/ui-*.md` - Are layouts complete?
- `design/gaps.md` - Are there critical gaps to address?

### Step 4: Implement Level 3 Code

Write code following the CRC cards and sequences:

**Add traceability comments** linking to design specs:

```typescript
/**
 * CRC: design/crc-FriendsManager.md
 * Spec: specs/friends.md
 * Sequences: design/seq-add-friend.md, design/seq-remove-friend.md
 */
class FriendsManager {
  /**
   * CRC: design/crc-FriendsManager.md - "Does: Add friend to list"
   * Sequence: design/seq-add-friend.md
   */
  addFriend(peerId: string, name: string): void {
    // implementation
  }
}
```

### Step 5: Update Bidirectionally

When code changes, update design specs. When design changes, update high-level specs.

---

## CRC Cards

### Purpose

CRC cards define **classes**, their **responsibilities**, and **collaborations**.

### Format

**File naming:** `design/crc-ClassName.md`

**Template:**
```markdown
# ClassName

**Source Spec:** source-file.md

## Responsibilities

### Knows
- attribute1: description
- attribute2: description

### Does
- behavior1: description
- behavior2: description

## Collaborators

- CollaboratorClass1: why/when collaboration occurs
- CollaboratorClass2: why/when collaboration occurs

## Sequences

- seq-scenario1.md: brief description
- seq-scenario2.md: brief description
```

### Example

```markdown
# FriendsManager

**Source Spec:** friends.md

## Responsibilities

### Knows
- friends: Array of Friend objects
- unsentRequests: Set of peer IDs with unsent friend requests

### Does
- addFriend: Create new friend with 'unsent' status
- removeFriend: Delete friend from list
- saveFriends: Persist to LocalStorage
- notifyUpdateListeners: Trigger UI updates

## Collaborators

- LocalStorage: For persisting friend data
- P2PService: For sending friend requests
- FriendsView: Notifies of data changes

## Sequences

- seq-add-friend.md: User adds friend by peer ID
- seq-remove-friend.md: User removes friend
```

---

## Sequence Diagrams

### Purpose

Sequence diagrams show **object interactions over time** for specific scenarios.

### Format

**File naming:** `design/seq-scenario-name.md`

**Template:**
```markdown
# Sequence: Scenario Name

**Source Spec:** source-file.md
**Use Case:** Brief description

## Participants

- Participant1: role/description
- Participant2: role/description

## Sequence

[PlantUML-generated ASCII art here]

## Notes

- Special considerations
- Error conditions
- Alternative flows
```

### Creating Sequences

Use PlantUML via Python (pre-approved):

```bash
python3 ./.claude/scripts/plantuml.py sequence "User -> FriendsManager: addFriend()
FriendsManager -> LocalStorage: save()
FriendsManager -> P2PService: sendRequest()
P2PService --> FriendsManager: ack"
```

**Important:** Pass PlantUML source as quoted argument (NOT heredoc) for pre-approval.

---

## UI Specifications

### Purpose

UI specs define **layout structure** for views and components.

### Format

**File naming:** `design/ui-view-name.md`

**Template:**
```markdown
# ViewName

**Source**: ui.feature.md
**Route**: /route (see manifest-ui.md)

**Purpose**: Brief description

**Data** (see crc-*.md):
- `dataItem1: Type` - Description

**Layout**:
[ASCII art diagram here]

**Events** (see crc-*.md):
- `eventName()` - Description

**CSS Classes**:
- `class-name` - Usage
```

### Principles

1. **Separation**: Layout (design) ≠ Styling (CSS) ≠ Behavior (CRC)
2. **Terseness**: Scannable lists, ASCII art, minimal prose
3. **Data clarity**: Always reference CRC cards for types
4. **References**: Point to CRC cards for behavior, manifest-ui.md for global concerns

---

## Architecture Mapping

### Purpose

**`design/architecture.md` serves as the "main program" for the design** - the entry point that organizes all design elements into logical systems.

This file:
- Groups related design elements (CRC cards, sequences, UI specs) into cohesive systems
- Identifies cross-cutting concerns that span multiple systems
- Provides a navigational overview of the entire design
- Makes the architectural structure immediately clear

### Why This Matters

When reviewing or working with a design, `architecture.md` answers:
- **"What systems exist?"** - See logical groupings at a glance
- **"Where do I start?"** - Entry point to understand the design structure
- **"Which components work together?"** - Related elements are grouped
- **"What's shared?"** - Cross-cutting concerns are explicit

### Structure

**File:** `design/architecture.md`

**Format:**
```markdown
# Architecture

**Entry point to the design - shows how design elements are organized into logical systems**

**Sources**: All CRC cards, sequences, UI specs, and manifest created from Level 1 specs

---

## Systems

### Character Management System

**Purpose**: Handle character creation, editing, and persistence

**Design Elements**:
- crc-Character.md
- crc-CharacterEditor.md
- seq-create-character.md
- ui-character-editor.md

### Friend System

**Purpose**: Manage peer relationships and friend lists

**Design Elements**:
- crc-Friend.md
- crc-FriendsManager.md
- seq-add-friend.md
- ui-friends-view.md

---

## Cross-Cutting Concerns

**Design elements that span multiple systems**

**Design Elements**:
- crc-Storage.md
- crc-Router.md
- manifest-ui.md
- seq-app-startup.md

---

*This file serves as the architectural "main program" - start here to understand the design structure*
```

### Key Principles

1. **Brevity** - Just systems, purposes, and file lists
2. **Complete coverage** - Every design file listed exactly once
3. **No duplicates** - Files in cross-cutting are NOT in any system
4. **Clear grouping** - Related elements grouped together
5. **Entry point** - Start here to understand the design

### Created When

The `designer` agent creates `architecture.md` after generating all CRC cards, sequences, and UI specs. It's part of the standard design workflow (step 10).

### Diagnostic Benefits

**`architecture.md` is invaluable for diagnosing problems:**

**1. Rapid Problem Localization**
- Immediately identify which system owns the functionality
- See all related components (CRC cards, sequences, UI specs) at a glance
- Identify cross-cutting concerns that might be involved

**2. Understanding Impact Scope**
- Is this isolated to one system? (Safe to change)
- Does it span multiple systems? (Need to coordinate changes)
- Is it in cross-cutting concerns? (Changes affect everything)

**3. Gap Detection**
- Missing components become obvious ("UI but no CRC card?")
- Unclear responsibilities stand out ("Why is this in two systems?")
- Unnecessary components are visible ("Doesn't fit anywhere... is it needed?")

**4. Communication with LLMs**
- Point to specific system: "Fix the Friend System's add-friend flow"
- LLM reads architecture.md → knows exactly which files to examine
- Faster, more accurate fixes with no guessing

**5. Debugging Interaction Problems**
- See system boundaries clearly ("Character System shouldn't call Friend System directly")
- Check cross-cutting patterns ("Should this go through Router instead?")
- Identify coupling issues ("Too many connections between these systems")

**In short:** `architecture.md` is a **map of the codebase at the design level** - diagnose architectural problems before even looking at code.

---

## Traceability

### Purpose

Maintain **bidirectional links** between all three levels.

### Traceability Map

**File:** `design/traceability.md`

**Structure:**
```markdown
# Traceability Map

## Level 1 ↔ Level 2 (Specs to Models)

### feature.md

**CRC Cards:**
- crc-Class1.md
- crc-Class2.md

**Sequence Diagrams:**
- seq-scenario1.md

**UI Specs:**
- ui-view-name.md

## Level 2 ↔ Level 3 (Models to Implementation)

### crc-ClassName.md

**Source Spec:** feature.md

**Implementation:**
- **path/to/ClassName.ext**
  - [ ] File header (CRC + Spec + Sequences)
  - [ ] ClassName class comment → crc-ClassName.md
  - [ ] methodName() comment → seq-scenario.md

**Tests:**
- **path/to/ClassName.test.ext**
  - [ ] File header referencing CRC card
```

### Traceability Comments in Code

Add comments linking implementation to design:

**TypeScript/JavaScript/Java/C#:**
```typescript
/**
 * CRC: design/crc-ClassName.md
 * Spec: specs/feature.md
 * Sequences: design/seq-scenario.md
 */
class ClassName {
  /**
   * CRC: design/crc-ClassName.md - "Does: methodName behavior"
   * Sequence: design/seq-scenario.md
   */
  methodName() {
    // implementation
  }
}
```

**Python:**
```python
"""
CRC: design/crc-ClassName.md
Spec: specs/feature.md
Sequences: design/seq-scenario.md
"""
class ClassName:
    def method_name(self):
        """
        CRC: design/crc-ClassName.md - "Does: method_name behavior"
        Sequence: design/seq-scenario.md
        """
        # implementation
```

**Adapt comment syntax to your language.**

---

## Bidirectional Updates

### Principle

**When any level changes, propagate updates through the documentation hierarchy.**

### Source Code Changes → Design Specs

- Modified implementation → Update CRC cards/sequences/UI specs if structure/behavior changed
- New classes/methods → Create corresponding CRC cards
- Changed interactions → Update sequence diagrams
- Template/view changes → Update UI specs

### Design Spec Changes → Architectural Specs

- Modified CRC cards/sequences → Update high-level specs if requirements/architecture affected
- New components → Document in feature specs
- Changed workflows → Update architectural documentation
- UI pattern changes → Update UI principles

### Abstraction Rules

1. **Always update up**: When code/design changes, ripple changes upward
2. **Maintain abstraction**: Each level documents at its appropriate abstraction level
3. **Keep consistency**: All three tiers must tell the same story
4. **Update traceability comments**: When docs change, update references in code

### Example Flow

```
User identifies missing persistence →
Update specs/game-worlds.md (add persistence requirement) →
Update design/crc-AdventureMode.md (add persistence responsibility) →
Update source code (implement + add traceability comments)

Later: Bug fix in implementation (terminateActiveWorld logic) →
Review if CRC card needs update (it does - split cleanup behavior) →
Review if specs/game-worlds.md needs update (it does - clarify persistence rules)
```

This ensures documentation remains **accurate and useful**, not just aspirational.

---

## Benefits

### 1. Better Architecture

- **Explicit design phase** prevents shotgun surgery and god classes
- **SOLID principles** naturally emerge from CRC (single responsibility, clear collaborations)
- **Early problem detection** catches design issues before coding

### 2. Complete Specifications

- **Sequences catch edge cases** that specs miss
- **All scenarios documented** before implementation
- **Clear interaction patterns** prevent integration problems

### 3. Traceability

- **Every line traces to design and requirements**
- **Impact analysis** easy (find all code for a requirement)
- **Audit trail** for decisions and changes

### 4. Maintainability

- **Changes propagate** through all documentation levels
- **Consistent story** at all abstraction levels
- **Refactoring guided** by design specs

### 5. Onboarding

- **New developers** understand system from design docs
- **Three levels** provide entry points for different learning styles
- **Complete documentation** reduces tribal knowledge

### 6. Quality

- **Testable** - sequences define test scenarios
- **Reviewable** - designs reviewed before coding
- **Verifiable** - traceability ensures completeness

---

## Reverse Engineering Existing Projects

### Applying CRC to Legacy Code

**You can apply CRC modeling to existing projects** by working backward through the three tiers.

**The reverse engineering workflow:**
```
Level 3: Existing code (analyze what exists)
   ↓
Level 2: Extract design (generate CRC cards, sequences, UI specs)
   ↓
Level 1: Document intent (produce human-readable specs)
```

### Step 1: Code Analysis → Design Extraction

**Use an LLM to analyze existing code and extract Level 2 design specs:**

**For each component:**
- Analyze class structure → Generate CRC cards
  - Identify responsibilities (what it knows/does)
  - Document collaborations (what it works with)
  - List methods and attributes

- Trace execution flows → Generate sequence diagrams
  - Follow key user scenarios through the code
  - Document object interactions
  - Capture error handling paths

- Examine templates/views → Generate UI specs
  - Extract layout structure
  - Document data bindings
  - Identify event handlers

**Process:**
```
Task(
  subagent_type="Explore",
  prompt="Analyze src/FriendsManager.ts and generate a CRC card in design/crc-FriendsManager.md.

  Document:
  - Responsibilities (knows/does)
  - Collaborators
  - Key methods and their purposes"
)
```

### Step 2: Design Analysis → Spec Generation

**Once you have Level 2 design specs, use them to generate Level 1 human-readable specs:**

**Review the extracted designs to understand:**
- What requirements do these classes fulfill?
- What user stories do these sequences support?
- What architectural patterns are in use?
- What design decisions were made?

**Generate high-level specs documenting:**
- Requirements and user stories
- Architecture and design intent
- UX flows and interaction patterns
- Business logic and rules
- Principles and constraints

**Process:**
```
Analyze design/crc-*.md and design/seq-*.md files.

Generate specs/friends.md documenting:
- What requirements the friends system fulfills
- User stories for friend management
- Design principles and constraints
```

### Benefits of Reverse Engineering

✅ **Document existing systems** - Create maintainable documentation for legacy code
✅ **Understand inherited projects** - Build mental model through structured analysis
✅ **Prepare for migration** - Document current state before major refactoring
✅ **Onboard to unfamiliar codebases** - Generate learning materials from code
✅ **Establish baseline** - Start traceability for previously undocumented projects

### Iterative Refinement

**Reverse engineering is iterative:**

1. **Extract initial designs** - Generate rough CRC cards and sequences
2. **Review and refine** - Correct misunderstandings, fill gaps
3. **Generate specs** - Create high-level documentation
4. **Validate** - Check specs against actual code behavior
5. **Update designs** - Refine Level 2 to match reality
6. **Establish traceability** - Add comments linking code to design

**Result:** A legacy codebase transformed into a documented, traceable system ready for maintenance and evolution.

---

## Additional Resources

### CLAUDE.md Sections

To add CRC workflow to your project's `CLAUDE.md`, see `.claude/shared/crc-install.md` for copy-paste snippets.

### Agent Documentation

- `.claude/agents/designer.md` - Designer agent (generates CRC cards, sequences, UI specs)
- `.claude/agents/sequence-diagrammer.md` - Sequence diagrammer agent (converts diagrams to PlantUML ASCII)
- `.claude/agents/test-designer.md` - Test designer agent (generates test designs)
- `.claude/agents/gap-analyzer.md` - Gap analysis agent
- `.claude/agents/documenter.md` - Documenter agent (generates project documentation)

### Skills

- `.claude/skills/init-crc-project.md` - Initialization command
- `.claude/skills/plantuml.md` - PlantUML sequence diagrams

---

**Last updated:** 2025-11-14
