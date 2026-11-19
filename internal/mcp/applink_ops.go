package mcp

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zot/frictionless/applink"
)

// defineAppLinkOps gives a session's Lua ui.linkApp, ui.unlinkApp and
// ui.deleteApp. Each is a few file-system calls, so they are synchronous: they
// return nothing on success and raise a Lua error carrying the operation's
// error otherwise. Call it on the session's executor.
// CRC: crc-AppLinkOps.md | Seq: seq-applink.md#4.1 | R234
func (s *Server) defineAppLinkOps(vendedID string) {
	ref, ok := s.UiServer.SessionRef(vendedID)
	if !ok {
		return
	}
	ls := ref.Lua().Session()
	if ls == nil {
		return
	}
	L := ls.State
	ui, ok := L.GetGlobal("ui").(*lua.LTable)
	if !ok {
		return
	}
	baseDir := s.baseDir
	for name, op := range map[string]func(string, string) error{
		"linkApp":   applink.Activate,
		"unlinkApp": applink.Deactivate,
		"deleteApp": applink.Delete,
	} {
		L.SetField(ui, name, L.NewFunction(func(L *lua.LState) int {
			// Seq: seq-applink.md#4.1.2
			if err := op(baseDir, L.CheckString(1)); err != nil {
				L.RaiseError("%s", err.Error())
			}
			return 0
		}))
	}
}
