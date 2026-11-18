// CRC: crc-Checkpoint.md | Test: test-Checkpoint.md
package main

import (
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/zot/frictionless/checkpoint"
)

// Test: the binary's notify never changes the exit status (R203, R205)
func TestCheckpointNotifyFailureKeepsExitZero(t *testing.T) {
	if _, err := checkpoint.Find(); err != nil {
		t.Skip("fossil is not installed")
	}
	base := t.TempDir()
	app := filepath.Join(base, "apps", "demo")
	if err := os.MkdirAll(app, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(app, "app.lua"), []byte("-- v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A port nothing listens on: the notify is attempted and fails.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	if err := os.WriteFile(filepath.Join(base, "mcp-port"), []byte(strconv.Itoa(port)), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FRICTIONLESS_MCP", "")

	if code := runCheckpoint([]string{"save", "demo", "--dir", base}); code != 0 {
		t.Fatalf("save exited %d with no server running, want 0", code)
	}
}
