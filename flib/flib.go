// Package flib provides an embeddable Frictionless runtime.
// Downstream binaries (e.g. ark) import this to get the full
// Frictionless stack — ui-engine, MCP tools, Lua globals, HTTP
// API — without needing to access internal packages.
package flib

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/zot/frictionless/internal/mcp"
	"github.com/zot/ui-engine/cli"
)

// Config holds configuration for an embedded Frictionless runtime.
type Config struct {
	Dir     string // base directory for UI assets and state
	Host    string // bind host (default "127.0.0.1")
	Project string // project directory for skill installation (optional; derived from Dir if empty)
	Port    int    // preferred HTTP port (0 = auto-select)
}

// Runtime is an embedded Frictionless server. Create with New,
// then call Configure, Start, and StartAPI in sequence.
type Runtime struct {
	Cfg       *cli.Config
	uiServer  *cli.Server
	mcpServer *mcp.Server
}

// New creates a Frictionless runtime configured for embedding.
// Call Configure, Start, and StartAPI to bring it up.
func New(cfg Config) (*Runtime, error) {
	host := cfg.Host
	if host == "" {
		host = "127.0.0.1"
	}

	uiCfg := cli.DefaultConfig()
	uiCfg.Server.Dir = cfg.Dir
	uiCfg.Server.Port = cfg.Port // 0 = auto-select
	uiCfg.Server.Host = host
	// CRC: crc-FlibRuntime.md | R182 — site the ui-engine backend control
	// socket per-Dir so multiple embedded runtimes don't collide on the
	// shared /tmp/ui.sock default. Empty Dir keeps the ui-engine default.
	if cfg.Dir != "" {
		uiCfg.Server.Socket = filepath.Join(cfg.Dir, "ui.sock")
	}
	uiCfg.Lua.Enabled = true
	uiCfg.Lua.Hotload = true

	srv := cli.NewServer(uiCfg)
	srv.StartCleanupWorker(time.Hour)

	mcpSrv := mcp.NewServer(
		uiCfg,
		srv,
		srv.GetViewdefManager(),
		func(port int) (string, error) {
			return srv.StartAsync(port)
		},
		func() int {
			return srv.GetSessions().Count()
		},
	)

	if cfg.Project != "" {
		mcpSrv.ProjectDir = cfg.Project
	}

	srv.SetRootSessionProvider(func() string {
		return mcpSrv.GetCurrentSessionID()
	})

	return &Runtime{
		Cfg:       uiCfg,
		uiServer:  srv,
		mcpServer: mcpSrv,
	}, nil
}

// Configure prepares the server environment — creates directories,
// runs auto-install if needed. Call before Start.
func (r *Runtime) Configure() error {
	return r.mcpServer.Configure(r.Cfg.Server.Dir)
}

// Start starts the UI HTTP server and creates a Lua session with
// the mcp global. Returns the base URL (e.g. "http://127.0.0.1:PORT").
func (r *Runtime) Start() (string, error) {
	url, err := r.mcpServer.StartAndCreateSession()
	if err != nil {
		return "", fmt.Errorf("frictionless start: %w", err)
	}
	return url, nil
}

// RegisterAPI registers Frictionless API handlers (/api/*, /wait,
// /state, /variables) on an external mux. Use this instead of
// StartAPI when the embedding binary has its own listener.
func (r *Runtime) RegisterAPI(mux *http.ServeMux) {
	r.mcpServer.RegisterAPIRoutes(mux)
}

// StartAPI starts a standalone HTTP API server that serves /api/*,
// /wait, /state, /variables endpoints. Returns the port number.
// Use RegisterAPI instead when embedding into an existing server.
func (r *Runtime) StartAPI() (int, error) {
	port, err := r.mcpServer.StartHTTPServer()
	if err != nil {
		return 0, fmt.Errorf("frictionless API: %w", err)
	}
	return port, nil
}

// RunLua executes Lua code in the current session.
// Returns the result as a string (or empty on nil result).
func (r *Runtime) RunLua(code string) (string, error) {
	result, err := r.mcpServer.CallRun(code)
	if err != nil {
		return "", err
	}
	if result == nil {
		return "", nil
	}
	return fmt.Sprintf("%v", result), nil
}

// WithLua executes a closure in the Lua executor goroutine (thread-safe)
// without triggering afterBatch (no UI update push). This is the passive
// execution path — use it to register Go functions on the Lua mcp table
// or perform other Lua-side setup after Start returns.
// CRC: crc-FlibRuntime.md
func (r *Runtime) WithLua(fn func(rt *cli.LuaRuntime) error) error {
	vendedID := r.mcpServer.GetCurrentVendedID()
	if vendedID == "" {
		return fmt.Errorf("no active session")
	}
	luaSession := r.uiServer.GetLuaSession(vendedID)
	if luaSession == nil {
		return fmt.Errorf("session %s not found", vendedID)
	}
	_, err := luaSession.ExecuteInSession(vendedID, func() (interface{}, error) {
		return nil, fn(luaSession)
	})
	return err
}

// UIHandleFunc registers a custom HTTP handler on the UI server's mux.
// CRC: crc-FlibRuntime.md
func (r *Runtime) UIHandleFunc(pattern string, handler http.HandlerFunc) {
	r.uiServer.HttpEndpoint.HandleFunc(pattern, handler)
}

// ThemeBlock returns the frictionless theme HTML block (script + CSS links)
// for the given base directory. Suitable for injecting into HTML templates
// between <!-- #frictionless --> markers.
func ThemeBlock(baseDir string) (string, error) {
	themes, err := mcp.ListThemes(baseDir)
	if err != nil {
		return "", fmt.Errorf("listing themes: %w", err)
	}
	defaultTheme := mcp.GetCurrentTheme(baseDir)
	return mcp.GenerateThemeBlock(baseDir, themes, defaultTheme, "  "), nil
}

// InjectAllThemeBlocks patches all HTML files in html/ that contain
// frictionless markers with the current theme block (script + CSS links
// with cache-busting). Called at startup and when themes change.
func InjectAllThemeBlocks(baseDir string) error {
	return mcp.InjectAllThemeBlocks(baseDir)
}

// Shutdown gracefully stops both the UI and API servers.
func (r *Runtime) Shutdown(ctx context.Context) error {
	r.mcpServer.RemovePortFiles()
	if err := r.mcpServer.ShutdownHTTPServer(ctx); err != nil {
		log.Printf("flib: API shutdown error: %v", err)
	}
	r.uiServer.Shutdown(ctx)
	return nil
}
