# Traceability Comment Insertion Skill

Add traceability comments to implementation methods based on CRC cards.

## Usage

**Pre-approved scripts:**
```bash
# Add traceability comments
./.claude/scripts/trace-add-comments.py <CRC-card-name>
./.claude/scripts/trace-add-comments.py --cleanup <CRC-card-name>

# Verify traceability comments are present
./.claude/scripts/trace-verify.py [phase-number]

# Find specs with no implementation
./.claude/scripts/trace-gap-analysis.py [--output FILE]
```

Examples:
- `./.claude/scripts/trace-add-comments.py ProfileService` - Add comments
- `./.claude/scripts/trace-add-comments.py --cleanup ProfileService` - Cleanup
- `./.claude/scripts/trace-verify.py` - Verify all files
- `./.claude/scripts/trace-gap-analysis.py` - Print gap analysis to stdout
- `./.claude/scripts/trace-gap-analysis.py --output plan.md` - Save to file

## What this skill does

Uses the **three-pass process** from `design/crc.md`:

### Pass 1: Read CRC card and traceability map
- Reads `design/crc-<card-name>.md` for implementation file and sequences
- Reads `design/traceability.md` for method-specific sequence references
- Reports what was found

### Pass 2: Find methods needing comments
- Scans implementation file for method declarations
- Checks which methods already have `CRC:` traceability comments
- Reports which methods need comments

### Pass 3: Compute new file version
- Inserts traceability comments for methods (in reverse line order)
- Writes new version to `.claude/scratch/<filename>.trace-new.ts`
- **Does NOT modify original file**
- Preserves file extension for syntax highlighting

## Workflow

1. **Run script:**
   ```bash
   ./.claude/scripts/trace-add-comments.py ProfileService
   ```

2. **Script outputs:**
   ```
   ✅ New version computed
   📄 Original: src/services/ProfileService.ts
   📝 New:      .claude/scratch/ProfileService.trace-new.ts
   📊 Changes:  5 comments added
   ```

3. **Claude reads both files and uses Edit tool**

4. **You see diff in IDE and approve/reject**

5. **If approved, Claude updates traceability.md checkboxes**

6. **Cleanup scratch file:**
   ```bash
   ./.claude/scripts/trace-add-comments.py --cleanup ProfileService
   ```

## Comment format

### Method implementation (with sequence)
```typescript
/**
 * methodName implementation
 *
 * CRC: design/crc-ClassName.md
 * Sequences:
 * - design/seq-operation.md (lines X-Y)
 * - design/seq-other.md (lines A-B)
 */
methodName(): void {
```

### Method implementation (without sequence)
```typescript
/**
 * methodName implementation
 *
 * CRC: design/crc-ClassName.md
 */
methodName(): void {
```

## Important notes

1. **Non-destructive** - Original file never modified directly
2. **Approval required** - User sees diff and approves via IDE
3. **Skips existing comments** - Won't add if method already has `CRC:` comment
4. **Reverse insertion order** - Inserts from bottom to top to preserve line numbers
5. **Requires CRC card** - Must have `design/crc-<name>.md` with `**Existing Code:**` field
6. **Cleanup option** - Use `--cleanup` flag to safely remove scratch file after applying changes

## Example output

```
📖 Reading CRC card...
   File: src/services/ProfileService.ts
   Specs: specs/storage.md
   Sequences: 4

📖 Reading traceability map...
   Method mappings: 2
   - getItem: 2 sequence(s)
   - setItem: 2 sequence(s)

📄 Reading src/services/ProfileService.ts...

📋 Found 8 methods:
   - getCurrentProfile at line 110
   - setCurrentProfile at line 117
   - getAllProfiles at line 137
   - createProfile at line 144
   - deleteProfile at line 164
   - getItem at line 220
   - setItem at line 231
   - removeItem at line 238

📝 Adding comments to 5 methods:
   ✓ getCurrentProfile (line 110)
   ✓ setCurrentProfile (line 117)
   ✓ getAllProfiles (line 137)
   ✓ createProfile (line 144)
   ✓ deleteProfile (line 164)

✅ New version computed
📄 Original: src/services/ProfileService.ts
📝 New:      .claude/scratch/ProfileService.trace-new.ts
📊 Changes:  5 comments added

💡 Next: Claude will use Edit tool to show diff for your approval
```

---

## Verification & Sync Mode

### What it does

**Bidirectional sync** between traceability.md and code:
- Treats **traceability.md as single source of truth** for what should be commented
- Checks each file/method against actual code
- **Automatically updates checkboxes** in traceability.md:
  - ✅ Checked if CRC comment exists in code
  - ⬜ Unchecked if CRC comment missing from code
- Dynamically discovers files from traceability.md (no hard-coded lists)
- Verifies file headers, class comments, and method comments

### Usage

```bash
python3 ./.claude/scripts/trace-verify.py
```

### Output

```
=== Phase 1 Traceability Sync & Verification ===

Found CRC cards for Phase 1:
  - crc-Character.md
  - crc-CharacterStorageService.md
  [...]

Checking crc-Character.md...
Checking crc-CharacterStorageService.md...
  ✅ CHECKING: src/services/CharacterStorageService.ts - File header (CRC + Spec + Sequences)
  ❌ UNCHECKING: src/services/CharacterStorageService.ts - saveCharacter() method comment (comment missing)
[...]

=== Summary ===
Total items checked: 86
Already correct:
  ✅ Checked (has comment): 28
  ⬜ Unchecked (no comment): 58

Changes made:
  ✅ Newly checked: 3
  ❌ Newly unchecked: 2

✅ PASS: All checkboxes in sync with code
```

### Exit codes

- `0` - All checkboxes in sync with code (or successfully synced)
- `1` - Missing files or checkboxes unchecked (comments need to be added)

---

## Gap Analysis Mode

### What it does

**Finds unimplemented specs** by scanning the entire codebase:
- Identifies CRC cards with no source code referencing them
- Identifies UI specs with no template files referencing them
- Checks if expected implementation files actually exist
- Generates migration plan with prioritized checklist

### Usage

```bash
# Print to stdout
./.claude/scripts/trace-gap-analysis.py

# Save to file
./.claude/scripts/trace-gap-analysis.py --output textcraft-migration-plan.md
```

### Output

```
📊 Trace Gap Analysis

🔍 Finding specs...
   Found 42 CRC cards
   Found 15 UI specs

🔍 Finding implementation files...
   Found 87 source files
   Found 71 template files

🔗 Checking CRC card references in source files...
   34 implemented, 8 not implemented

🔗 Checking UI spec references in template files...
   10 implemented, 5 not implemented

✅ Gap analysis written to: textcraft-migration-plan.md
```

### Report Contents

1. **CRC Cards Section:**
   - Total count and implementation status
   - List of unimplemented CRC cards with expected file paths
   - List of implemented CRC cards with references

2. **UI Specs Section:**
   - Total count and implementation status
   - List of unimplemented UI specs with expected template paths
   - List of implemented UI specs with references

3. **Migration Priority:**
   - Checklist of files to create
   - Checklist of files needing traceability comments
   - Distinguishes between missing files and existing files without comments

### Example Report Snippet

```markdown
## CRC Cards

**Total CRC cards:** 42
**Implemented:** 34
**Not implemented:** 8

### ❌ CRC Cards with NO Implementation

- **crc-AdventureMode.md**
  - Expected implementation: `src/ui/AdventureMode.ts` (❌ missing)

- **crc-WorldListView.md**
  - Expected implementation: `src/ui/WorldListView.ts` (❌ missing)

## 🎯 Migration Priority

### Source Files (CRC implementations)

- [ ] Create `src/ui/AdventureMode.ts` implementing crc-AdventureMode.md
- [ ] Create `src/ui/WorldListView.ts` implementing crc-WorldListView.md
```

---

## Summary

The trace skill provides three complementary modes:

1. **Add Comments** - Inserts traceability comments into existing code
2. **Verify** - Syncs checkboxes between traceability.md and actual code
3. **Gap Analysis** - Identifies specs with no implementation (NEW)

Use gap analysis to create migration plans for major refactorings!
