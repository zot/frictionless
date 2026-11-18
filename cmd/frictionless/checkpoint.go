package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zot/frictionless/checkpoint"
)

// splitDirFlag removes `--dir DIR` or `--dir=DIR` from args, returning the base
// directory (default .ui) and the remaining arguments.
func splitDirFlag(args []string) (string, []string) {
	baseDir := ".ui"
	var rest []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--dir" && i+1 < len(args) {
			baseDir = args[i+1]
			i++
		} else if v, ok := strings.CutPrefix(args[i], "--dir="); ok {
			baseDir = v
		} else {
			rest = append(rest, args[i])
		}
	}
	return baseDir, rest
}

// runCheckpoint is `frictionless checkpoint CMD APP [ARG] [--dir DIR]`.
// CRC: crc-Checkpoint.md | Seq: seq-checkpoint.md#1.1 | R203, R204, R205
func runCheckpoint(args []string) int {
	baseDir, rest := splitDirFlag(args)
	err := checkpoint.Command(rest, filepath.Join(baseDir, "apps"), os.Stdout, func() { notifyCheckpointChange(baseDir) })
	if err != nil {
		// Seq: seq-checkpoint.md#1.2
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

// notifyCheckpointChange asks a running server's app-console to refresh its
// checkpoint display. It is best effort: every failure is ignored. Under a
// server (FRICTIONLESS_MCP set) it does nothing, because a checkpoint run from
// Lua would post to the very executor that is waiting on it.
// CRC: crc-Checkpoint.md | Seq: seq-checkpoint.md#1.1.5 | R79, R205
func notifyCheckpointChange(baseDir string) {
	if os.Getenv("FRICTIONLESS_MCP") != "" {
		return
	}
	port, err := os.ReadFile(filepath.Join(baseDir, "mcp-port"))
	if err != nil {
		return
	}
	body, _ := json.Marshal(map[string]string{"code": "if appConsole then appConsole._checkpointsTime = 0 end"})
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Post("http://127.0.0.1:"+strings.TrimSpace(string(port))+"/api/ui_run", "application/json", bytes.NewReader(body))
	if err == nil {
		resp.Body.Close()
	}
}
