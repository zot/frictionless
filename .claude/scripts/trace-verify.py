#!/usr/bin/env python3
"""
trace-verify.py - Verify and sync traceability comments with traceability.md
Treats traceability.md as source of truth for what should be commented
Syncs checkbox states: checked if comment exists, unchecked if missing

Usage: python3 trace-verify.py [phase-number]
Example: python3 trace-verify.py 2
"""
import sys
import os
import re
from pathlib import Path
from typing import List, Tuple, Set, Optional

# ANSI color codes
class Colors:
    GREEN = '\033[0;32m'
    RED = '\033[0;31m'
    YELLOW = '\033[1;33m'
    NC = '\033[0m'  # No Color

class TraceabilityVerifier:
    """Verifies and syncs traceability comments with traceability.md"""

    def __init__(self, phase: int = 1):
        self.phase = phase
        self.traceability_file = Path('design/traceability.md')
        self.total_items = 0
        self.checked_items = 0
        self.unchecked_items = 0
        self.newly_checked = 0
        self.newly_unchecked = 0
        self.missing_files = 0

        # Phase to spec patterns mapping
        self.phase_specs = {
            1: ['specs/characters.md', 'specs/storage.md', 'specs/logging.md'],
            2: ['specs/ui.characters.md', 'specs/ui.md', 'specs/routes.md'],
        }

    def get_phase_crc_cards(self) -> Set[str]:
        """Extract CRC cards for this phase from traceability.md"""
        if self.phase not in self.phase_specs:
            return set()

        spec_patterns = self.phase_specs[self.phase]
        crc_cards = set()

        with open(self.traceability_file, 'r', encoding='utf-8') as f:
            content = f.read()

        # Split into sections by ### headers
        sections = re.split(r'^###\s+', content, flags=re.MULTILINE)

        for section in sections:
            # Check if this section matches one of our phase specs
            first_line = section.split('\n')[0] if section else ''

            if not any(pattern in first_line for pattern in spec_patterns):
                continue

            # Extract CRC cards from **CRC Cards:** section
            in_crc_section = False
            for line in section.split('\n'):
                if line.strip() == '**CRC Cards:**':
                    in_crc_section = True
                    continue

                # Stop at next bold section
                if in_crc_section and line.strip().startswith('**') and line.strip().endswith('**'):
                    in_crc_section = False
                    continue

                # Stop at section boundary
                if in_crc_section and (line.startswith('---') or line.startswith('###')):
                    in_crc_section = False
                    continue

                # Extract CRC card name
                if in_crc_section and line.strip().startswith('- crc-'):
                    card_name = line.strip()[2:]  # Remove "- "
                    card_name = card_name.replace('.md', '')
                    crc_cards.add(card_name)

        return crc_cards

    def parse_crc_implementation(self, crc_card: str) -> List[Tuple[str, str, str]]:
        """
        Extract implementation files and checkboxes for a given CRC card.
        Returns list of tuples: (file_path, checkbox_state, description)
        """
        with open(self.traceability_file, 'r', encoding='utf-8') as f:
            content = f.read()

        # Find the section for this CRC card
        pattern = f'^### {re.escape(crc_card)}\\.md'
        sections = re.split(r'^###\s+', content, flags=re.MULTILINE)

        target_section = None
        for section in sections:
            if section.startswith(f'{crc_card}.md'):
                target_section = section
                break

        if not target_section:
            return []

        items = []
        current_file = None
        in_impl_or_tests = False

        for line in target_section.split('\n'):
            # Check for section markers
            if line.strip() == '**Implementation:**' or line.strip() == '**Tests:**':
                in_impl_or_tests = True
                continue

            # Stop at certain sections
            if line.strip().startswith('**Appears in Sequences:**'):
                break
            if line.strip() in ['**Interface:**', '**Implementations:**']:
                continue

            # Extract file path from **src/path/file.ts**
            if in_impl_or_tests and line.strip().startswith('- **src/'):
                match = re.search(r'\*\*([^*]+)\*\*', line)
                if match:
                    current_file = match.group(1)
                continue

            # Extract checkbox items
            if in_impl_or_tests and re.match(r'^\s+- \[.\]', line):
                match = re.match(r'^\s+- \[(.)\](.+)$', line)
                if match and current_file:
                    checkbox_state = match.group(1)
                    description = match.group(2).strip()
                    items.append((current_file, checkbox_state, description))

        return items

    def has_crc_header(self, file_path: str) -> bool:
        """Check if a file header has CRC: comment"""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                # Read first 20 lines
                for _ in range(20):
                    line = f.readline()
                    if not line:
                        break
                    if re.search(r'^\s*\*\s+CRC:\s+design/', line):
                        return True
            return False
        except (IOError, UnicodeDecodeError):
            return False

    def has_crc_comment(self, file_path: str, search_term: str) -> bool:
        """Check if a specific method/class has CRC comment"""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                lines = f.readlines()

            # Look for search_term and check 10 lines before it
            for i, line in enumerate(lines):
                if search_term in line:
                    # Check up to 10 lines before
                    start = max(0, i - 10)
                    context = ''.join(lines[start:i+1])
                    if re.search(r'CRC:\s+design/', context):
                        return True
            return False
        except (IOError, UnicodeDecodeError):
            return False

    def update_checkbox(self, file_path: str, old_state: str, description: str, new_state: str) -> None:
        """Update checkbox in traceability.md"""
        with open(self.traceability_file, 'r', encoding='utf-8') as f:
            content = f.read()

        # Escape special regex characters in description
        escaped_desc = re.escape(description)

        # Pattern: "  - [old_state] description"
        # Replace with: "  - [new_state] description"
        pattern = f'^  - \\[{re.escape(old_state)}\\] {escaped_desc}$'
        replacement = f'  - [{new_state}] {description}'

        new_content = re.sub(pattern, replacement, content, flags=re.MULTILINE)

        with open(self.traceability_file, 'w', encoding='utf-8') as f:
            f.write(new_content)

    def verify_and_sync(self) -> int:
        """Main verification and sync logic. Returns exit code."""
        print(f'=== Phase {self.phase} Traceability Sync & Verification ===')
        print()

        # Check if traceability file exists
        if not self.traceability_file.exists():
            print(f'{Colors.RED}❌ Error: {self.traceability_file} not found{Colors.NC}')
            return 1

        # Get CRC cards for this phase
        crc_cards = self.get_phase_crc_cards()

        if not crc_cards:
            print(f'{Colors.RED}❌ No CRC cards found for Phase {self.phase}{Colors.NC}')
            print('   (or phase not configured)')
            return 1

        print(f'Found CRC cards for Phase {self.phase}:')
        for card in sorted(crc_cards):
            print(f'  - {card}.md')
        print()

        # Process each CRC card
        for crc_card in sorted(crc_cards):
            print(f'Checking {crc_card}.md...')

            # Get implementation items for this card
            items = self.parse_crc_implementation(crc_card)

            if not items:
                print(f'  {Colors.YELLOW}⚠️  No implementation items found{Colors.NC}')
                continue

            # Process each checkbox item
            for file_path, checked_state, description in items:
                self.total_items += 1

                # Check if file exists
                if not Path(file_path).exists():
                    print(f'  {Colors.YELLOW}⚠️  File not found: {file_path}{Colors.NC}')
                    self.missing_files += 1
                    continue

                # Determine what kind of comment to check for
                has_comment = False

                if description.startswith('File header'):
                    # Check for file header CRC comment
                    has_comment = self.has_crc_header(file_path)
                else:
                    # Check for specific method/class comment
                    # Extract the key term from description
                    search_term = description
                    search_term = re.sub(r' comment.*', '', search_term)
                    search_term = re.sub(r' method$', '', search_term)
                    search_term = re.sub(r' class$', '', search_term)
                    search_term = re.sub(r' interface$', '', search_term)
                    search_term = re.sub(r'\(\)$', '', search_term)

                    has_comment = self.has_crc_comment(file_path, search_term)

                # Determine if checkbox state needs updating
                should_be_checked = 'x' if has_comment else ' '

                # Update checkbox if state is wrong
                if checked_state != should_be_checked:
                    if should_be_checked == 'x':
                        print(f'  {Colors.GREEN}✅ CHECKING: {file_path} - {description}{Colors.NC}')
                        self.newly_checked += 1
                    else:
                        print(f'  {Colors.RED}❌ UNCHECKING: {file_path} - {description} (comment missing){Colors.NC}')
                        self.newly_unchecked += 1

                    self.update_checkbox(file_path, checked_state, description, should_be_checked)
                else:
                    # State is correct
                    if checked_state == 'x':
                        self.checked_items += 1
                    else:
                        self.unchecked_items += 1

        # Print summary
        print()
        print('=== Summary ===')
        print(f'Total items checked: {self.total_items}')
        print('Already correct:')
        print(f'  ✅ Checked (has comment): {self.checked_items}')
        print(f'  ⬜ Unchecked (no comment): {self.unchecked_items}')
        print()
        print('Changes made:')
        print(f'  ✅ Newly checked: {self.newly_checked}')
        print(f'  ❌ Newly unchecked: {self.newly_unchecked}')

        if self.missing_files > 0:
            print(f'  {Colors.YELLOW}⚠️  Missing files: {self.missing_files}{Colors.NC}')

        print()

        # Determine exit code
        if self.newly_checked == 0 and self.newly_unchecked == 0 and self.missing_files == 0:
            print(f'{Colors.GREEN}✅ PASS: All checkboxes in sync with code{Colors.NC}')
            return 0
        elif self.newly_unchecked > 0 or self.missing_files > 0:
            print(f'{Colors.RED}❌ FAIL: Some items need traceability comments{Colors.NC}')
            return 1
        else:
            print(f'{Colors.GREEN}✅ PASS: Checkboxes updated to match code{Colors.NC}')
            return 0


def main():
    """Main entry point"""
    # Get phase from command line argument
    phase = 1
    if len(sys.argv) > 1:
        try:
            phase = int(sys.argv[1])
        except ValueError:
            print(f'Error: Invalid phase number: {sys.argv[1]}')
            print('Usage: python3 trace-verify.py [phase-number]')
            sys.exit(1)

    # Create verifier and run
    verifier = TraceabilityVerifier(phase)
    exit_code = verifier.verify_and_sync()
    sys.exit(exit_code)


if __name__ == '__main__':
    main()
