# CRC Modeling - Executive Summary

## The Core Problem

Claude "generously" infers features that aren't in your specs. By the time you notice (when a spec change removes a feature you rely on), you've invested hours reviewing thousands of lines of code.

**The fundamental issue:** You need to verify Claude's interpretation of your requirements *before* it generates code, not after.

**What you get:** The power to catch misunderstandings, architectural flaws, and unwanted inferences when they're trivial to fix rather than expensive to refactor.

## The Solution

The three-tier system (specs → design → code) with bidirectional traceability transforms your codebase from a black box into a navigable, maintainable asset.

**Design artifacts produced:**
- CRC cards (classes, responsibilities, collaborations)
- Sequence diagrams (object interactions over time)
- UI specifications (layout structure, data bindings, event handlers)
- Test designs (test specifications with input/output, optional)

## Key Benefits

✅ **Catch misunderstandings early** - Review design before committing to code
✅ **Stop losing features** - Detect "generous" inferences before they become code
✅ **Maintain coherence** - Design layer provides consistency across sessions
✅ **Knowledge retention** - Documentation survives staff transitions
✅ **Trace bugs cleanly** - Follow from symptom → code → design → requirement
✅ **Safe refactoring** - Impact analysis prevents breaking changes
✅ **Audit trails** - Link requirements to implementation for compliance
✅ **Ship with confidence** - Validate architecture before writing thousands of lines

## Key Capabilities

**Impact analysis:**
- Find all code implementing a requirement in seconds
- Trace bug reports from symptom → test → code → design → spec
- Identify what needs updating when requirements evolve

**Safe refactoring:**
- Sequences document all expected interactions (test scenarios)
- CRC cards define clear responsibilities (refactoring boundaries)
- Impact analysis prevents breaking changes

**Legacy system support:**
- Apply CRC to existing code through reverse engineering
- Extract designs from inherited codebases
- Finally document that monolith you've been avoiding

**AI assistant stability:**
- Explicit design layer provides consistency across sessions
- Design patterns guide future decisions
- Maintains architectural coherence as projects evolve

## Automation

The designer agent automates the tedious parts, making thorough architecture practical rather than aspirational. You focus on intent and validation while the AI generates comprehensive design documentation.

## Bottom Line

Deliver better software faster with less risk, whether you're building your first MVP or maintaining a million-line enterprise system.

---

**Learn more:** See `.claude/doc/crc.md` for complete documentation.

**Get started:** Run `/init-crc-project` to set up CRC modeling in your project.
