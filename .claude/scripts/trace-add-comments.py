#!/usr/bin/env python3
"""
Add traceability comments to implementation methods.

Computes new file version with comments inserted, writes to .claude/scratch/
Claude then uses Edit tool to show diff for approval.

Usage: trace-add-comments.py <CRC-card-name>
       trace-add-comments.py --cleanup <CRC-card-name>

Example: trace-add-comments.py ProfileService
         trace-add-comments.py --cleanup ProfileService
"""

import sys
import re
from pathlib import Path

def read_crc_card(crc_name):
    """Read CRC card to get implementation file and spec references."""
    crc_file = Path(f"design/crc-{crc_name}.md")
    if not crc_file.exists():
        print(f"❌ CRC card not found: {crc_file}")
        sys.exit(1)

    content = crc_file.read_text()

    # Extract implementation file
    impl_match = re.search(r'\*\*Existing Code:\*\*\s+(.+)', content)
    if not impl_match:
        print(f"❌ No 'Existing Code:' found in CRC card")
        sys.exit(1)
    impl_file = impl_match.group(1).strip()

    # Extract source spec
    spec_match = re.search(r'\*\*Source Spec:\*\*\s+(.+)', content)
    specs = spec_match.group(1).strip() if spec_match else ""

    # Extract sequences
    sequences = []
    in_sequences = False
    for line in content.split('\n'):
        if line.startswith('## Sequences'):
            in_sequences = True
            continue
        if in_sequences:
            if line.startswith('##') or line.startswith('#'):
                break
            if line.strip().startswith('- '):
                seq = line.strip()[2:]
                if not seq.startswith('design/'):
                    seq = f"design/{seq}"
                sequences.append(seq)

    return {
        "name": crc_name,
        "file": impl_file,
        "specs": specs,
        "sequences": sequences
    }

def read_traceability_map(crc_name):
    """Read traceability.md to get method-specific sequence references."""
    trace_file = Path("design/traceability.md")
    if not trace_file.exists():
        return {}

    content = trace_file.read_text()

    # Find the section for this CRC card
    pattern = f"### crc-{crc_name}.md"
    section_start = content.find(pattern)
    if section_start == -1:
        return {}

    # Find next ### to get section bounds
    next_section = content.find("\n### ", section_start + len(pattern))
    if next_section == -1:
        section = content[section_start:]
    else:
        section = content[section_start:next_section]

    # Parse method -> sequence mappings
    method_sequences = {}
    for line in section.split('\n'):
        # Pattern: "  - [ ] ClassName.methodName() implementation → seq-name.md (lines X-Y)"
        match = re.search(r'- \[.\]\s+\w+\.(\w+)\(\)\s+implementation.*?→\s*(.+)', line)
        if match:
            method_name = match.group(1)
            sequences_str = match.group(2)

            # Parse multiple sequence references
            seq_refs = []
            for seq_match in re.finditer(r'(seq-[\w-]+\.md)\s*\(lines\s+([\d-]+)\)', sequences_str):
                seq_refs.append({
                    "file": f"design/{seq_match.group(1)}",
                    "lines": seq_match.group(2)
                })

            if seq_refs:
                method_sequences[method_name] = seq_refs

    return method_sequences

def find_methods(file_content, class_name):
    """Find all method declarations in the class implementation."""
    methods = []
    lines = file_content.split('\n')

    # Find class declaration
    class_line = -1
    for i, line in enumerate(lines):
        if re.search(rf'class\s+{class_name}\s+', line):
            class_line = i
            break

    if class_line == -1:
        return methods

    # Find methods (look for function declarations)
    # Pattern: methodName(...): ReturnType {
    for i in range(class_line + 1, len(lines)):
        # Skip if we've left the class (closed brace at indent 0)
        if i > class_line and lines[i].strip() == '}' and not lines[i].startswith('    '):
            break

        # Look for method pattern: indent + name + ( + params + ): + type
        match = re.match(r'^    (\w+)\([^)]*\):\s*\w+.*\{?\s*$', lines[i])
        if match:
            method_name = match.group(1)
            methods.append({
                "name": method_name,
                "line": i + 1  # 1-based line numbers
            })

    return methods

def find_existing_comment_range(lines, method_line):
    """Find the range of existing comment block before method.

    Returns: (start_line, end_line) or None if no comment exists.
    Lines are 0-based indices.
    """
    # Look backwards from method line for comment block
    end_line = method_line - 1

    # Skip empty lines
    while end_line >= 0 and not lines[end_line].strip():
        end_line -= 1

    if end_line < 0:
        return None

    # Check if this is end of comment block (*/)
    if not lines[end_line].strip().endswith('*/'):
        return None

    # Find start of comment block (/**)
    start_line = end_line
    while start_line > 0:
        if lines[start_line].strip().startswith('/**'):
            return (start_line, end_line)
        start_line -= 1

    return None

def extract_description_from_comment(file_lines, comment_range):
    """Extract description lines from existing comment (everything before CRC/Sequence)."""
    if not comment_range:
        return []

    start, end = comment_range
    descriptions = []

    for i in range(start + 1, end):  # Skip /** and */
        line = file_lines[i].strip()

        # Skip empty comment lines
        if line in ('*', ''):
            continue

        # Stop at CRC/Sequence/existing traceability markers
        if any(marker in line for marker in ['CRC:', 'Sequence:', 'Spec:', '@']):
            break

        # Remove leading * and whitespace
        if line.startswith('*'):
            line = line[1:].strip()
            if line:
                descriptions.append(line)

    return descriptions

def generate_comment(method_name, crc_data, method_sequences, existing_descriptions=None):
    """Generate traceability comment, merging with existing description."""
    lines = ["    /**"]

    # Add existing description lines if available
    if existing_descriptions:
        for desc in existing_descriptions:
            lines.append(f"     * {desc}")
        lines.append("     *")
    else:
        # Fallback to generic description
        lines.append(f"     * {method_name} implementation")
        lines.append("     *")

    # Add CRC reference
    lines.append(f"     * CRC: design/crc-{crc_data['name']}.md")

    # Add sequence references if available
    if method_name in method_sequences:
        seqs = method_sequences[method_name]
        if len(seqs) == 1:
            lines.append(f"     * Sequence: {seqs[0]['file']} (lines {seqs[0]['lines']})")
        else:
            lines.append("     * Sequences:")
            for seq in seqs:
                lines.append(f"     * - {seq['file']} (lines {seq['lines']})")

    lines.append("     */")
    return '\n'.join(lines)

def insert_comments(file_content, crc_data, method_sequences):
    """Insert or replace traceability comments for methods."""
    lines = file_content.split('\n')

    # Find all methods
    methods = find_methods(file_content, crc_data['name'])

    if not methods:
        print(f"⚠️  No methods found in {crc_data['name']} class")
        return file_content, []

    print(f"📋 Found {len(methods)} methods:")
    for m in methods:
        print(f"   - {m['name']} at line {m['line']}")

    # Check which methods need new/updated comments (reverse order for modification)
    methods_to_update = []
    for method in reversed(methods):
        method_line_idx = method['line'] - 1  # Convert to 0-based
        comment_range = find_existing_comment_range(lines, method_line_idx)

        # Check if existing comment has CRC reference
        has_crc = False
        if comment_range:
            start, end = comment_range
            for i in range(start, end + 1):
                if 'CRC:' in lines[i]:
                    has_crc = True
                    break

        if not has_crc:
            methods_to_update.append({
                'method': method,
                'comment_range': comment_range
            })

    if not methods_to_update:
        print(f"✅ All methods already have traceability comments")
        return file_content, []

    print(f"\n📝 Updating comments for {len(methods_to_update)} methods:")

    # Update comments (working backwards to preserve line numbers)
    changes = []
    for item in methods_to_update:
        method = item['method']
        comment_range = item['comment_range']

        # Extract existing descriptions if comment exists
        existing_descriptions = None
        if comment_range:
            existing_descriptions = extract_description_from_comment(lines, comment_range)

        new_comment = generate_comment(method['name'], crc_data, method_sequences, existing_descriptions)

        if comment_range:
            # Replace existing comment
            start, end = comment_range
            # Delete old comment lines
            for _ in range(end - start + 1):
                del lines[start]
            # Insert new comment
            lines.insert(start, new_comment)
            changes.append(f"{method['name']} (replaced)")
            print(f"   ✓ {method['name']} (replaced existing comment)")
        else:
            # Insert new comment before method
            insert_at = method['line'] - 1  # Convert to 0-based
            lines.insert(insert_at, new_comment)
            changes.append(f"{method['name']} (new)")
            print(f"   ✓ {method['name']} (added new comment)")

    return '\n'.join(lines), changes

def cleanup_scratch_file(crc_name):
    """Remove the scratch file for the given CRC card."""
    # Read CRC card to get implementation file
    crc_data = read_crc_card(crc_name)
    impl_file = Path(crc_data['file'])

    # Build scratch file path
    scratch_dir = Path(".claude/scratch")
    stem = impl_file.stem
    ext = impl_file.suffix
    scratch_file = scratch_dir / f"{stem}.trace-new{ext}"

    if scratch_file.exists():
        scratch_file.unlink()
        print(f"✅ Cleaned up: {scratch_file}")
    else:
        print(f"ℹ️  No scratch file found: {scratch_file}")

def main():
    # Check for cleanup mode
    if len(sys.argv) >= 2 and sys.argv[1] == "--cleanup":
        if len(sys.argv) != 3:
            print("Usage: trace-add-comments.py --cleanup <CRC-card-name>")
            print("Example: trace-add-comments.py --cleanup ProfileService")
            sys.exit(1)
        cleanup_scratch_file(sys.argv[2])
        sys.exit(0)

    # Normal mode
    if len(sys.argv) != 2:
        print("Usage: trace-add-comments.py <CRC-card-name>")
        print("       trace-add-comments.py --cleanup <CRC-card-name>")
        print("Example: trace-add-comments.py ProfileService")
        print("         trace-add-comments.py --cleanup ProfileService")
        sys.exit(1)

    crc_name = sys.argv[1]
    print(f"\n🔍 Computing traceability comments for {crc_name}\n")

    # Read CRC card
    print("📖 Reading CRC card...")
    crc_data = read_crc_card(crc_name)
    print(f"   File: {crc_data['file']}")
    print(f"   Specs: {crc_data['specs']}")
    print(f"   Sequences: {len(crc_data['sequences'])}")

    # Read traceability map
    print("\n📖 Reading traceability map...")
    method_sequences = read_traceability_map(crc_name)
    print(f"   Method mappings: {len(method_sequences)}")
    for method, seqs in method_sequences.items():
        print(f"   - {method}: {len(seqs)} sequence(s)")

    # Read implementation file
    impl_file = Path(crc_data['file'])
    if not impl_file.exists():
        print(f"\n❌ Implementation file not found: {impl_file}")
        sys.exit(1)

    print(f"\n📄 Reading {impl_file}...")
    original_content = impl_file.read_text()

    # Insert comments
    print()
    new_content, changes = insert_comments(original_content, crc_data, method_sequences)

    if not changes:
        print(f"\n✅ No changes needed - all methods already have comments")
        sys.exit(0)

    # Write new version (keep extension for syntax highlighting)
    scratch_dir = Path(".claude/scratch")
    scratch_dir.mkdir(parents=True, exist_ok=True)

    # Keep extension: ProfileService.ts -> ProfileService.trace-new.ts
    stem = impl_file.stem  # "ProfileService"
    ext = impl_file.suffix  # ".ts"
    output_file = scratch_dir / f"{stem}.trace-new{ext}"
    output_file.write_text(new_content)

    print(f"\n✅ New version computed")
    print(f"📄 Original: {impl_file}")
    print(f"📝 New:      {output_file}")
    print(f"📊 Changes:  {len(changes)} comments added")
    print(f"\n💡 Next: Claude will use Edit tool to show diff for your approval")

if __name__ == "__main__":
    main()
