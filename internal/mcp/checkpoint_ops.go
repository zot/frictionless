package mcp

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	lua "github.com/yuin/gopher-lua"
	"github.com/zot/frictionless/checkpoint"
)

const (
	checkpointStatusDeadline  = 60 * time.Second
	checkpointInstallDeadline = 120 * time.Second
)

// defineCheckpointOps gives a session's Lua ui.checkpointStatus and
// ui.checkpointInstall as host operations: each returns at once, fossil runs off
// the executor, and the result lands in the field the caller names. Call it on
// the session's executor.
// CRC: crc-CheckpointOps.md | Seq: seq-checkpoint.md#5.1 | R223
func (s *Server) defineCheckpointOps(vendedID string) {
	ref, ok := s.UiServer.SessionRef(vendedID)
	if !ok {
		return
	}
	appsDir := filepath.Join(s.baseDir, "apps")
	lv := ref.Lua()

	// Seq: seq-checkpoint.md#5.3
	lv.DefineHostOp("checkpointStatus", checkpointStatusDeadline, func(args []lua.LValue) (func(context.Context) (any, error), error) {
		names, err := luaStrings(args)
		if err != nil {
			return nil, err
		}
		return statusWork(appsDir, names), nil
	})

	lv.DefineHostOp("checkpointInstall", checkpointInstallDeadline, func(args []lua.LValue) (func(context.Context) (any, error), error) {
		if len(args) == 0 || args[0].Type() != lua.LTString {
			return nil, errors.New("checkpointInstall: an app name is required")
		}
		name := args[0].String()
		// Seq: seq-checkpoint.md#6.2 | R225
		return func(context.Context) (any, error) {
			fossil, err := checkpoint.Find()
			if err != nil {
				return nil, err
			}
			if err := (checkpoint.App{Dir: filepath.Join(appsDir, name), Fossil: fossil}).Install(); err != nil {
				return nil, err
			}
			return true, nil
		}, nil
	})
}

// luaStrings reads the first argument, a Lua list of strings, on the executor.
func luaStrings(args []lua.LValue) ([]string, error) {
	var tbl lua.LValue = lua.LNil
	if len(args) > 0 {
		tbl = args[0]
	}
	list, ok := tbl.(*lua.LTable)
	if !ok {
		return nil, errors.New("checkpointStatus: a list of app names is required")
	}
	names := make([]string, 0, list.Len())
	for i := 1; i <= list.Len(); i++ {
		names = append(names, list.RawGetInt(i).String())
	}
	return names, nil
}

// statusWork reads each app's checkpoint status off the executor. An app whose
// status cannot be read gets an error entry; the others still report.
// CRC: crc-CheckpointOps.md | Seq: seq-checkpoint.md#5.4 | R224
func statusWork(appsDir string, names []string) func(context.Context) (any, error) {
	return func(ctx context.Context) (any, error) {
		fossil, err := checkpoint.Find()
		if err != nil {
			return nil, err
		}
		result := make(map[string]any, len(names))
		for _, name := range names {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			st, err := checkpoint.App{Dir: filepath.Join(appsDir, name), Fossil: fossil}.Status()
			if err != nil {
				// Seq: seq-checkpoint.md#5.4.1
				result[name] = map[string]any{"error": err.Error()}
				continue
			}
			result[name] = map[string]any{"count": st.Count, "downloaded": st.Downloaded, "changed": st.Changed}
		}
		return result, nil
	}
}
