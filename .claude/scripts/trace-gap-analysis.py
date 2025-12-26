#!/usr/bin/env python3
"""
Trace Gap Analysis - Find specs with no implementation references

Scans the codebase to find:
- CRC cards (design/crc-*.md) with no source code referencing them
- Sequence diagrams (design/seq-*.md) with no source code referencing them
- UI specs (design/ui-*.md) with no templates referencing them

Usage:
    ./trace-gap-analysis.py [--output FILE]

Examples:
    ./trace-gap-analysis.py                    # Print to stdout
    ./trace-gap-analysis.py --output plan.md   # Save to file
"""

import re
import sys
from pathlib import Path
from typing import Dict, List, Set, Optional
from dataclasses import dataclass, field

@dataclass
class SpecStatus:
    """Status of a spec file"""
    spec_file: str
    spec_type: str  # "CRC", "SEQ", or "UI"
    referenced_by: List[str] = field(default_factory=list)

    @property
    def is_implemented(self) -> bool:
        return len(self.referenced_by) > 0

def find_all_crc_cards() -> List[Path]:
    """Find all CRC cards in design/"""
    crc_dir = Path("design")
    if not crc_dir.exists():
        return []

    return sorted(crc_dir.glob("crc-*.md"))

def find_all_seq_diagrams() -> List[Path]:
    """Find all sequence diagrams in design/"""
    seq_dir = Path("design")
    if not seq_dir.exists():
        return []

    return sorted(seq_dir.glob("seq-*.md"))

def find_all_ui_specs() -> List[Path]:
    """Find all UI specs in design/"""
    ui_dir = Path("design")
    if not ui_dir.exists():
        return []

    return sorted(ui_dir.glob("ui-*.md"))

def find_source_files() -> List[Path]:
    """Find all TypeScript source files"""
    src_dir = Path("src")
    if not src_dir.exists():
        return []

    return sorted(src_dir.rglob("*.ts"))

def find_template_files() -> List[Path]:
    """Find all HTML template files"""
    templates_dir = Path("public/templates")
    if not templates_dir.exists():
        return []

    return sorted(templates_dir.rglob("*.html"))

def check_crc_references(crc_cards: List[Path], source_files: List[Path]) -> Dict[str, SpecStatus]:
    """Check which CRC cards are referenced in source code"""

    results = {}

    for crc_path in crc_cards:
        crc_name = crc_path.name
        status = SpecStatus(spec_file=str(crc_path), spec_type="CRC")

        # Search for references to this CRC card in source files
        for src_path in source_files:
            try:
                content = src_path.read_text()
                # Look for "CRC: design/crc-*.md" or "CRC: crc-*.md" pattern
                if f"design/{crc_name}" in content or f"CRC: {crc_name}" in content:
                    status.referenced_by.append(str(src_path))
            except Exception as e:
                print(f"⚠️  Error reading {src_path}: {e}", file=sys.stderr)

        results[crc_name] = status

    return results

def check_seq_references(seq_diagrams: List[Path], source_files: List[Path]) -> Dict[str, SpecStatus]:
    """Check which sequence diagrams are referenced in source code"""

    results = {}

    for seq_path in seq_diagrams:
        seq_name = seq_path.name
        status = SpecStatus(spec_file=str(seq_path), spec_type="SEQ")

        # Search for references to this sequence diagram in source files
        for src_path in source_files:
            try:
                content = src_path.read_text()
                # Look for "SEQ: design/seq-*.md" or "SEQ: seq-*.md" pattern
                if f"design/{seq_name}" in content or f"SEQ: {seq_name}" in content:
                    status.referenced_by.append(str(src_path))
            except Exception as e:
                print(f"⚠️  Error reading {src_path}: {e}", file=sys.stderr)

        results[seq_name] = status

    return results

def check_ui_references(ui_specs: List[Path], template_files: List[Path]) -> Dict[str, SpecStatus]:
    """Check which UI specs are referenced in template files"""

    results = {}

    for ui_path in ui_specs:
        ui_name = ui_path.name
        status = SpecStatus(spec_file=str(ui_path), spec_type="UI")

        # Search for references to this UI spec in template files
        for tmpl_path in template_files:
            try:
                content = tmpl_path.read_text()
                # Look for "@layout: design/ui-*.md" or similar pattern
                if f"design/{ui_name}" in content or f"@layout: {ui_name}" in content or f"@layout {ui_name}" in content:
                    status.referenced_by.append(str(tmpl_path))
            except Exception as e:
                print(f"⚠️  Error reading {tmpl_path}: {e}", file=sys.stderr)

        results[ui_name] = status

    return results

def parse_crc_implementation_path(crc_path: Path) -> Optional[str]:
    """Extract implementation file path from CRC card"""
    try:
        content = crc_path.read_text()

        # Look for "**Implementation:**" section
        impl_match = re.search(r'\*\*Implementation:\*\*\n- \*\*([^*]+)\*\*', content)
        if impl_match:
            return impl_match.group(1).strip()
    except Exception:
        pass

    return None

def parse_ui_template_path(ui_path: Path) -> Optional[str]:
    """Extract template file path from UI spec"""
    try:
        content = ui_path.read_text()

        # Look for "**Primary**: `public/templates/...`" or "**Template**: ..."
        tmpl_match = re.search(r'\*\*(?:Primary|Template|Templates):\*\*[^\n]*?`([^`]+)`', content)
        if tmpl_match:
            return tmpl_match.group(1).strip()
    except Exception:
        pass

    return None

def format_markdown_report(crc_results: Dict[str, SpecStatus], seq_results: Dict[str, SpecStatus], ui_results: Dict[str, SpecStatus]) -> str:
    """Format gap analysis as markdown"""

    output = []

    # Header
    output.append("# Traceability Gap Analysis Report")
    output.append("")
    output.append("**Generated by:** .claude/scripts/trace-gap-analysis.py")
    output.append("")
    output.append("This report identifies specs that have no implementation references.")
    output.append("")
    output.append("---")
    output.append("")

    # CRC Cards Analysis
    output.append("## CRC Cards")
    output.append("")

    unimplemented_crc = [name for name, status in crc_results.items() if not status.is_implemented]
    implemented_crc = [name for name, status in crc_results.items() if status.is_implemented]

    output.append(f"**Total CRC cards:** {len(crc_results)}")
    output.append(f"**Implemented:** {len(implemented_crc)}")
    output.append(f"**Not implemented:** {len(unimplemented_crc)}")
    output.append("")

    if unimplemented_crc:
        output.append("### ❌ CRC Cards with NO Implementation")
        output.append("")
        output.append("These CRC cards have no source files referencing them:")
        output.append("")

        for crc_name in sorted(unimplemented_crc):
            status = crc_results[crc_name]
            crc_path = Path(status.spec_file)
            impl_path = parse_crc_implementation_path(crc_path)

            output.append(f"- **{crc_name}**")
            if impl_path:
                file_exists = Path(impl_path).exists()
                exists_str = "✅ exists" if file_exists else "❌ missing"
                output.append(f"  - Expected implementation: `{impl_path}` ({exists_str})")
            output.append("")

    if implemented_crc:
        output.append("### ✅ CRC Cards with Implementation")
        output.append("")

        for crc_name in sorted(implemented_crc):
            status = crc_results[crc_name]
            output.append(f"- **{crc_name}**")
            for ref in status.referenced_by:
                output.append(f"  - `{ref}`")
            output.append("")

    output.append("---")
    output.append("")

    # Sequence Diagrams Analysis
    output.append("## Sequence Diagrams")
    output.append("")

    unimplemented_seq = [name for name, status in seq_results.items() if not status.is_implemented]
    implemented_seq = [name for name, status in seq_results.items() if status.is_implemented]

    output.append(f"**Total sequence diagrams:** {len(seq_results)}")
    output.append(f"**Implemented:** {len(implemented_seq)}")
    output.append(f"**Not implemented:** {len(unimplemented_seq)}")
    output.append("")

    if unimplemented_seq:
        output.append("### ❌ Sequence Diagrams with NO Implementation")
        output.append("")
        output.append("These sequence diagrams have no source files referencing them:")
        output.append("")

        for seq_name in sorted(unimplemented_seq):
            output.append(f"- **{seq_name}**")
            output.append("")

    if implemented_seq:
        output.append("### ✅ Sequence Diagrams with Implementation")
        output.append("")

        for seq_name in sorted(implemented_seq):
            status = seq_results[seq_name]
            output.append(f"- **{seq_name}**")
            for ref in status.referenced_by:
                output.append(f"  - `{ref}`")
            output.append("")

    output.append("---")
    output.append("")

    # UI Specs Analysis
    output.append("## UI Specs")
    output.append("")

    unimplemented_ui = [name for name, status in ui_results.items() if not status.is_implemented]
    implemented_ui = [name for name, status in ui_results.items() if status.is_implemented]

    output.append(f"**Total UI specs:** {len(ui_results)}")
    output.append(f"**Implemented:** {len(implemented_ui)}")
    output.append(f"**Not implemented:** {len(unimplemented_ui)}")
    output.append("")

    if unimplemented_ui:
        output.append("### ❌ UI Specs with NO Templates")
        output.append("")
        output.append("These UI specs have no template files referencing them:")
        output.append("")

        for ui_name in sorted(unimplemented_ui):
            status = ui_results[ui_name]
            ui_path = Path(status.spec_file)
            tmpl_path = parse_ui_template_path(ui_path)

            output.append(f"- **{ui_name}**")
            if tmpl_path:
                file_exists = Path(tmpl_path).exists()
                exists_str = "✅ exists" if file_exists else "❌ missing"
                output.append(f"  - Expected template: `{tmpl_path}` ({exists_str})")
            output.append("")

    if implemented_ui:
        output.append("### ✅ UI Specs with Templates")
        output.append("")

        for ui_name in sorted(implemented_ui):
            status = ui_results[ui_name]
            output.append(f"- **{ui_name}**")
            for ref in status.referenced_by:
                output.append(f"  - `{ref}`")
            output.append("")

    output.append("---")
    output.append("")

    # Migration Priority
    if unimplemented_crc or unimplemented_seq or unimplemented_ui:
        output.append("## 🎯 Migration Priority")
        output.append("")
        output.append("**Files to create/update:**")
        output.append("")

        if unimplemented_crc:
            output.append("### Source Files (CRC implementations)")
            output.append("")
            for crc_name in sorted(unimplemented_crc):
                crc_path = Path(crc_results[crc_name].spec_file)
                impl_path = parse_crc_implementation_path(crc_path)
                if impl_path:
                    file_exists = Path(impl_path).exists()
                    if file_exists:
                        output.append(f"- [ ] Add traceability comments to `{impl_path}` → {crc_name}")
                    else:
                        output.append(f"- [ ] Create `{impl_path}` implementing {crc_name}")
            output.append("")

        if unimplemented_seq:
            output.append("### Source Files (Sequence diagram implementations)")
            output.append("")
            for seq_name in sorted(unimplemented_seq):
                output.append(f"- [ ] Add traceability comments referencing {seq_name}")
            output.append("")

        if unimplemented_ui:
            output.append("### Template Files (UI spec implementations)")
            output.append("")
            for ui_name in sorted(unimplemented_ui):
                ui_path = Path(ui_results[ui_name].spec_file)
                tmpl_path = parse_ui_template_path(ui_path)
                if tmpl_path:
                    file_exists = Path(tmpl_path).exists()
                    if file_exists:
                        output.append(f"- [ ] Add @layout comment to `{tmpl_path}` → {ui_name}")
                    else:
                        output.append(f"- [ ] Create `{tmpl_path}` implementing {ui_name}")
            output.append("")

    return '\n'.join(output)

def main():
    """Main entry point"""

    # Parse arguments
    output_file = None
    args = sys.argv[1:]

    i = 0
    while i < len(args):
        if args[i] == '--output':
            if i + 1 < len(args):
                output_file = args[i + 1]
                i += 2
            else:
                print("❌ Error: --output requires a filename")
                sys.exit(1)
        else:
            print(f"❌ Error: Unknown argument: {args[i]}")
            sys.exit(1)

    print("📊 Trace Gap Analysis")
    print("")

    # Find all specs
    print("🔍 Finding specs...")
    crc_cards = find_all_crc_cards()
    seq_diagrams = find_all_seq_diagrams()
    ui_specs = find_all_ui_specs()
    print(f"   Found {len(crc_cards)} CRC cards")
    print(f"   Found {len(seq_diagrams)} sequence diagrams")
    print(f"   Found {len(ui_specs)} UI specs")
    print("")

    # Find all implementation files
    print("🔍 Finding implementation files...")
    source_files = find_source_files()
    template_files = find_template_files()
    print(f"   Found {len(source_files)} source files")
    print(f"   Found {len(template_files)} template files")
    print("")

    # Check references
    print("🔗 Checking CRC card references in source files...")
    crc_results = check_crc_references(crc_cards, source_files)
    unimplemented_crc = sum(1 for s in crc_results.values() if not s.is_implemented)
    print(f"   {len(crc_results) - unimplemented_crc} implemented, {unimplemented_crc} not implemented")
    print("")

    print("🔗 Checking sequence diagram references in source files...")
    seq_results = check_seq_references(seq_diagrams, source_files)
    unimplemented_seq = sum(1 for s in seq_results.values() if not s.is_implemented)
    print(f"   {len(seq_results) - unimplemented_seq} implemented, {unimplemented_seq} not implemented")
    print("")

    print("🔗 Checking UI spec references in template files...")
    ui_results = check_ui_references(ui_specs, template_files)
    unimplemented_ui = sum(1 for s in ui_results.values() if not s.is_implemented)
    print(f"   {len(ui_results) - unimplemented_ui} implemented, {unimplemented_ui} not implemented")
    print("")

    # Format report
    markdown = format_markdown_report(crc_results, seq_results, ui_results)

    # Write to file or stdout
    if output_file:
        Path(output_file).write_text(markdown)
        print(f"✅ Gap analysis written to: {output_file}")
    else:
        print("")
        print("=" * 80)
        print("")
        print(markdown)

if __name__ == '__main__':
    main()
