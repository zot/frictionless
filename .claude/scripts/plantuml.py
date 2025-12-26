#!/usr/bin/env python3
"""
plantuml.py - Generate ASCII sequence diagrams using PlantUML
"""
import sys
import os
import subprocess
import tempfile
import shutil
from pathlib import Path

def get_paths():
    """Get script directory and project root paths."""
    script_dir = Path(__file__).parent.resolve()
    project_root = script_dir.parent.parent
    plantuml_jar = project_root / '.claude' / 'bin' / 'plantuml.jar'
    return project_root, plantuml_jar

def usage():
    """Print usage information."""
    prog = sys.argv[0]
    print(f"""Usage: {prog} sequence [PLANTUML_SOURCE]

Generate ASCII sequence diagram from PlantUML syntax (via argument or stdin)

Options:
  --help            Show this help message

Examples:
  # Via command-line argument (recommended - no approval needed)
  {prog} sequence "User -> View: click save
  View -> Model: save()
  Model --> View: success"

  # Via stdin
  echo "User -> View: click save" | {prog} sequence

  # Multi-line with newlines
  {prog} sequence "User -> Editor: save
  alt valid
    Editor -> Storage: save()
  else invalid
    Editor -> User: show errors
  end"

PlantUML Syntax:
  ->      Synchronous message
  -->     Return message (dashed)
  note left of / note right of
  alt/else/end    Conditional
  loop/end        Loop
""")
    sys.exit(1)

def find_java():
    """Find Java executable."""
    # Check common locations
    java_paths = [
        Path.home() / 'bin' / 'java',
        shutil.which('java'),
    ]

    for java_path in java_paths:
        if java_path and Path(java_path).is_file():
            # Verify it's actually Java
            try:
                result = subprocess.run(
                    [str(java_path), '-version'],
                    capture_output=True,
                    text=True,
                    timeout=5
                )
                if result.returncode == 0 or 'version' in result.stderr.lower():
                    return str(java_path)
            except (subprocess.SubprocessError, FileNotFoundError):
                continue

    return None

def generate_sequence(plantuml_source):
    """Generate ASCII sequence diagram from PlantUML source."""
    project_root, plantuml_jar = get_paths()

    # Check if PlantUML jar exists
    if not plantuml_jar.is_file():
        print(f"Error: PlantUML jar not found at {plantuml_jar}", file=sys.stderr)
        print("Download from: https://plantuml.com/download", file=sys.stderr)
        sys.exit(1)

    # Find Java executable
    java_cmd = find_java()
    if not java_cmd:
        print("Error: Java is not installed or not in PATH", file=sys.stderr)
        print("PlantUML requires Java runtime", file=sys.stderr)
        sys.exit(1)

    # Create temporary files
    with tempfile.NamedTemporaryFile(mode='w', suffix='.puml', delete=False) as temp_input:
        temp_input_path = Path(temp_input.name)

        # Wrap input in PlantUML sequence diagram syntax
        temp_input.write("@startuml\n")
        temp_input.write(plantuml_source)
        temp_input.write("\n@enduml\n")
        temp_input.flush()

        try:
            # Generate ASCII diagram using PlantUML
            result = subprocess.run(
                [java_cmd, '-jar', str(plantuml_jar), '-tutxt', str(temp_input_path)],
                capture_output=True,
                text=True,
                timeout=30
            )

            # PlantUML outputs to .utxt file with same base name
            output_file = temp_input_path.with_suffix('.utxt')

            if output_file.is_file():
                with open(output_file, 'r') as f:
                    print(f.read(), end='')
                output_file.unlink()
            else:
                print("Error: PlantUML did not generate output file", file=sys.stderr)
                print(f"Expected: {output_file}", file=sys.stderr)
                print("Input was:", file=sys.stderr)
                with open(temp_input_path, 'r') as f:
                    print(f.read(), file=sys.stderr)
                sys.exit(1)

        finally:
            # Cleanup temporary files
            if temp_input_path.is_file():
                temp_input_path.unlink()

def main():
    """Main command dispatcher."""
    if len(sys.argv) < 2:
        usage()

    command = sys.argv[1]

    if command in ('--help', 'help'):
        usage()

    if command == 'sequence':
        # Check for --help
        if len(sys.argv) > 2 and sys.argv[2] == '--help':
            usage()

        # Get PlantUML source from args or stdin
        if len(sys.argv) > 2:
            # Source provided as argument
            plantuml_source = ' '.join(sys.argv[2:])
        else:
            # Read from stdin
            if sys.stdin.isatty():
                print("Error: No PlantUML source provided", file=sys.stderr)
                usage()
            plantuml_source = sys.stdin.read()

        generate_sequence(plantuml_source)

    elif command == 'generate-ascii':
        # Legacy compatibility
        print("Warning: 'generate-ascii' is deprecated, use 'sequence' instead", file=sys.stderr)
        plantuml_source = ' '.join(sys.argv[2:]) if len(sys.argv) > 2 else sys.stdin.read()
        generate_sequence(plantuml_source)

    else:
        print(f"Unknown command: {command}", file=sys.stderr)
        usage()

if __name__ == '__main__':
    main()
