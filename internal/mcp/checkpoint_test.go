// CRC: crc-MCPScript.md | Test: test-Checkpoint.md
package mcp

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Test: the script hands checkpoint off to the binary (R201)
func TestScriptHandsCheckpointToBinary(t *testing.T) {
	script, err := os.ReadFile(filepath.Join("..", "..", "install", "mcp"))
	if err != nil {
		t.Fatal(err)
	}
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	bin := filepath.Join(home, ".claude", "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{
		filepath.Join(base, "mcp"):         string(script),
		filepath.Join(bin, "fossil"):       "#!/bin/sh\n", // present, so the script downloads nothing
		filepath.Join(bin, "frictionless"): "#!/bin/sh\necho \"args: $*\"\nexit 3\n",
	} {
		if err := os.WriteFile(name, []byte(content), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	cmd := exec.Command(filepath.Join(base, "mcp"), "checkpoint", "count", "demo")
	cmd.Env = append(os.Environ(), "HOME="+home)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()

	if want := "args: checkpoint count demo --dir " + base + "\n"; out.String() != want {
		t.Errorf("script output %q, want the binary's %q", out.String(), want)
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 3 {
		t.Fatalf("script exit = %v, want the binary's exit status 3", err)
	}
}
