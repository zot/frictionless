# Traceability Guide

**Purpose:** Complete guide to maintaining bidirectional traceability between specs, CRC models, and code.

**Related Files:**
- `design/README.md` - Main CRC modeling documentation
- `design/traceability.md` - Formal traceability map with checkboxes
- `design/gaps.md` - Gap analysis documenting what CRC reveals beyond specs

---

## Bidirectional Traceability Requirements

All artifacts maintain **bidirectional links** to their sources:
- Level 1 (specs) ↔ Level 2 (CRC/sequences)
- Level 2 (CRC/sequences) ↔ Level 3 (code)
- The `traceability.md` file maintains the complete map

**⚠️ CRITICAL: Code must have traceability comments**

The following items that come from specs MUST have linking comments:
- **Implementation files** - Header comments linking to CRC cards and specs
- **Classes** - Comments linking to their CRC card
- **Interfaces** - Comments linking to CRC card or spec
- **Key methods** - Comments referencing spec sections or sequences

**Without these comments, traceability is lost and design intent cannot be recovered.**

---

## Comment Format Standards

### File Header Format

```typescript
/**
 * [Class/Interface Name] - [Brief description]
 *
 * CRC: crc-[ClassName].md
 * Spec: [spec-name].md
 * Sequences: seq-[operation].md
 */
```

**Examples:**

```typescript
/**
 * Character Storage Service - Handles character persistence
 *
 * CRC: crc-CharacterStorageService.md
 * Spec: characters.md, storage.md
 * Sequences: seq-save-character.md, seq-load-character.md
 */
export class CharacterStorageService implements ICharacterStorageService {
```

```typescript
/**
 * Profile Service - Profile-based storage isolation
 *
 * CRC: crc-ProfileService.md
 * Spec: main.md, storage.md
 */
export class ProfileService implements IProfileService {
```

### Method Comment Format

```typescript
/**
 * Save character with hash-based optimization
 * Spec: storage.md "Hash-based save optimization"
 * Sequence: seq-save-character.md (lines 46-57)
 */
async saveCharacter(character: ICharacter): Promise<void> {
```

### Method Comment with Multiple Sequences

```typescript
/**
 * Get item from profile-scoped storage
 *
 * CRC: crc-ProfileService.md
 * Sequences:
 * - seq-load-character.md (lines 29-38)
 * - seq-save-character.md (lines 68-78)
 */
getItem(key: string): string | null {
    return localStorage.getItem(this.getStorageKey(key));
}
```

---

## Formal Traceability Map Structure

The `traceability.md` file uses a **formal, checkbox-based format** to ensure complete traceability.

### Three-Tier Structure

1. **Level 1 ↔ Level 2**: Maps each spec to its CRC cards and sequence diagrams
2. **Level 2 ↔ Level 3**: Maps each CRC card to implementation files with **granular checkboxes**
3. **Sequence Participants**: Maps each sequence diagram participant to code with line numbers

### Checkbox Format

**Key feature: Checkboxes for every code element**

Each CRC card section lists:
- [ ] File header (CRC + Spec + Sequences)
- [ ] Class comment → crc-ClassName.md
- [ ] Interface comment → crc-ClassName.md
- [ ] Method comments → seq-name.md (lines X-Y)

### Why This Format?

1. **Prevents skipping** - Can't forget to add comments when there's an explicit checkbox
2. **Machine-parseable** - Could automate verification with scripts
3. **Clear completion criteria** - Phase is done when all checkboxes checked
4. **Line number precision** - Know exactly which methods need sequence references
5. **Audit trail** - Can verify traceability is complete by checking boxes

### Example from traceability.md

```markdown
### crc-CharacterStorageService.md

**Source Spec:** characters.md, storage.md

**Implementation:**
- **src/services/CharacterStorageService.ts**
  - [ ] File header (CRC + Spec + Sequences)
  - [ ] CharacterStorageService class comment → crc-CharacterStorageService.md
  - [ ] ICharacterStorageService interface comment → crc-CharacterStorageService.md
  - [ ] getAllCharacters() method comment → seq-load-character.md (lines 26-78)
  - [ ] saveCharacter() method comment → seq-save-character.md (lines 43-97)
  - [ ] getCharacter() method comment → seq-load-character.md (lines 80-95)
  - [ ] deleteCharacter() method comment → seq-delete-character.md (lines 15-42)

**Tests:**
- **src/services/CharacterStorageService.test.ts**
  - [ ] File header referencing CRC card
```

This ensures **nothing gets missed** when adding traceability comments to code (Step 9).

---

## Adding Traceability Comments

### Using the /trace Command

**Quick Reference:**
```bash
# Generate traceability comments for a CRC card
/trace ProfileService

# Or use the script directly:
./.claude/scripts/trace-add-comments.py ProfileService

# After approval, cleanup:
./.claude/scripts/trace-add-comments.py --cleanup ProfileService
```

**What it does:**
- Reads CRC card and traceability.md to find method-to-sequence mappings
- Scans implementation file for methods without CRC comments
- Computes new file with comments (preserves existing descriptions)
- Writes to `.claude/scratch/` for review
- Never modifies original file directly
- Requires user approval via IDE diff

**See:** `.claude/skills/trace.md` and `.claude/commands/trace.md` for complete documentation

### Three-Pass Process

**Pass 1: Read CRC card and traceability map**
- Reads `crc-<card-name>.md` for implementation file and sequences
- Reads `design/traceability.md` for method-specific sequence references
- Reports what was found

**Pass 2: Find methods needing comments**
- Scans implementation file for method declarations
- Checks which methods already have `CRC:` traceability comments
- Reports which methods need comments

**Pass 3: Compute new file version**
- Inserts traceability comments for methods (in reverse line order)
- Writes new version to `.claude/scratch/<filename>.trace-new.ts`
- **Does NOT modify original file**
- Preserves file extension for syntax highlighting

### Workflow

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

---

## Handling Interfaces with Multiple Implementations

**Problem:** Interfaces can have multiple implementations. The traceability checkboxes must account for BOTH interface declarations AND each implementation class.

**Solution:**

1. **Interface declaration** gets a checkbox in the file where it's declared
2. **Each implementation class** gets its own checkbox in its file

**Example:**

```markdown
### crc-CharacterStorageService.md

**Implementation:**
- **src/character/types.ts**
  - [ ] ICharacterStorageService interface comment → crc-CharacterStorageService.md

- **src/services/CharacterStorageService.ts**
  - [ ] File header (CRC + Spec + Sequences)
  - [ ] CharacterStorageService class comment → crc-CharacterStorageService.md
  - [ ] saveCharacter() method comment → seq-save-character.md
  - [ ] loadCharacter() method comment → seq-load-character.md
```

---

## Verification

### Manual Verification

Check that all code elements from CRC cards have traceability comments:

1. **File headers** - Open file, check first 10 lines for CRC comment
2. **Class comments** - Look immediately before `export class`
3. **Interface comments** - Look immediately before `export interface`
4. **Method comments** - Look immediately before key methods

### Automated Verification & Sync

Use the trace verification script for **bidirectional sync** between traceability.md and code:

```bash
# Default to Phase 1
python3 ./.claude/scripts/trace-verify.py

# Verify specific phase
python3 ./.claude/scripts/trace-verify.py 1
python3 ./.claude/scripts/trace-verify.py 2
```

**What it does:**
- Treats **traceability.md as single source of truth** for what should be commented
- Checks if each file/method has the required CRC comment
- **Automatically updates checkboxes** in traceability.md to match code reality
- Dynamically discovers files (no hard-coded lists)

**Output:**
```
=== Phase 1 Traceability Sync & Verification ===

Found CRC cards for Phase 1:
  - crc-Character.md
  - crc-CharacterStorageService.md
  ...

Checking crc-CharacterStorageService.md...
  ✅ CHECKING: src/services/CharacterStorageService.ts - File header
  ❌ UNCHECKING: src/services/CharacterStorageService.ts - saveCharacter() method comment

=== Summary ===
Total items checked: 86
Changes made:
  ✅ Newly checked: 3
  ❌ Newly unchecked: 2

✅ PASS: All checkboxes in sync with code
```

**Exit codes:**
- `0` - All checkboxes in sync with code
- `1` - Missing files or checkboxes unchecked (comments need to be added)

---

## Best Practices

1. **Add comments AFTER code exists** - Don't try to add comments while generating code
2. **Use /trace command** - Automates most of the work
3. **Run trace-verify.py** - Automatically syncs checkboxes in traceability.md with code
4. **Fix unchecked items** - If script unchecks boxes, add the missing CRC comments
5. **Re-run to verify** - Run trace-verify.py again until all items are in sync
6. **Keep comments up-to-date** - When refactoring, update traceability comments and re-sync
7. **Link to lines** - Include line numbers in sequence references
8. **Be specific** - Link to the exact CRC card or sequence diagram

---

## Common Patterns

### Service Class

```typescript
/**
 * Profile Service - Profile-based storage isolation
 *
 * CRC: crc-ProfileService.md
 * Spec: main.md, storage.md
 * Sequences: seq-initialize-profiles.md, seq-switch-profile.md
 */
export class ProfileService implements IProfileService {
  /**
   * Get current profile
   * CRC: crc-ProfileService.md
   */
  getCurrentProfile(): IProfile {
    // ...
  }

  /**
   * Switch to different profile
   * CRC: crc-ProfileService.md
   * Sequence: seq-switch-profile.md (lines 15-42)
   */
  setCurrentProfile(name: string): void {
    // ...
  }
}
```

### UI View Class

```typescript
/**
 * Character Editor View - Character editing UI
 *
 * CRC: crc-CharacterEditorView.md
 * Spec: ui.characters.md
 * Sequences: seq-edit-character.md, seq-save-character-ui.md
 */
export class CharacterEditorView {
  /**
   * Save character with validation
   * Sequence: seq-save-character-ui.md (lines 22-101)
   */
  private async saveCharacter(): Promise<void> {
    // ...
  }
}
```

### Utility Class

```typescript
/**
 * Character Calculations - Character stat calculations
 *
 * CRC: crc-CharacterCalculations.md
 * Spec: characters.md
 */
export class CharacterCalculations {
  /**
   * Calculate rank from total XP (backward compatibility)
   * CRC: crc-CharacterCalculations.md
   */
  static calculateRank(totalXP: number): number {
    // ...
  }
}
```

---

## Summary

**Traceability is bidirectional:**
- Specs → CRC cards → Code (forward)
- Code → CRC cards → Specs (backward)

**Key files:**
- `traceability.md` - Single source of truth for what should be commented
- CRC comment format - Links in every file
- `/trace` command - Automated comment insertion
- `trace-verify.py` - Automated bidirectional sync

**Process:**
1. Create CRC cards from specs
2. Generate code from CRC cards
3. Fill out traceability.md with checkboxes
4. Add traceability comments to code (use /trace)
5. Run trace-verify.py to sync checkboxes with code reality
6. Add missing comments for unchecked items
7. Re-run trace-verify.py until all items are in sync

**Result:** Complete bidirectional traceability with automated synchronization ensuring design intent is preserved forever.
