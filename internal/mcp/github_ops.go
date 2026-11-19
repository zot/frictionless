package mcp

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	lua "github.com/yuin/gopher-lua"
	"github.com/zot/frictionless/githubinstall"
)

const (
	githubInvestigateDeadline = 60 * time.Second
	githubInstallDeadline     = 120 * time.Second
)

// defineGitHubOps gives a session's Lua ui.githubInvestigate and
// ui.githubInstall as host operations: each returns at once, the network and
// file work run off the executor, and the result lands in the field the caller
// names. Call it on the session's executor.
// CRC: crc-GitHubOps.md | Seq: seq-github-install.md#1.1 | R239, R245, R246
func (s *Server) defineGitHubOps(vendedID string) {
	ref, ok := s.UiServer.SessionRef(vendedID)
	if !ok {
		return
	}
	appsDir := filepath.Join(s.baseDir, "apps")
	lv := ref.Lua()

	// Seq: seq-github-install.md#1.1.1
	lv.DefineHostOp("githubInvestigate", githubInvestigateDeadline, func(args []lua.LValue) (func(context.Context) (any, error), error) {
		src, err := sourceArg(args)
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) (any, error) {
			inv, err := githubinstall.Investigate(ctx, src)
			if err != nil {
				return nil, err
			}
			return inv.AsLua(), nil
		}, nil
	})

	lv.DefineHostOp("githubInstall", githubInstallDeadline, func(args []lua.LValue) (func(context.Context) (any, error), error) {
		src, err := sourceArg(args)
		if err != nil {
			return nil, err
		}
		if len(args) < 2 || args[1].Type() != lua.LTString {
			return nil, errors.New("githubInstall: an app name is required")
		}
		name := args[1].String()
		// Seq: seq-github-install.md#2.2 | R245
		return func(ctx context.Context) (any, error) {
			if err := githubinstall.Install(ctx, src, appsDir, name); err != nil {
				return nil, err
			}
			return true, nil
		}, nil
	})
}

// sourceArg reads the first argument, a GitHub tree URL, on the executor.
func sourceArg(args []lua.LValue) (githubinstall.Source, error) {
	if len(args) == 0 || args[0].Type() != lua.LTString {
		return githubinstall.Source{}, errors.New("a GitHub tree URL is required")
	}
	return githubinstall.ParseURL(args[0].String())
}
