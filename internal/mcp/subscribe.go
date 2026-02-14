// Package mcp — subscribe integration for the publisher pub/sub server.
// CRC: crc-MCPSubscribe.md | Seq: seq-publish-subscribe.md, seq-publisher-lifecycle.md
package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	lua "github.com/yuin/gopher-lua"
)

const (
	publisherAddr    = "http://localhost:25283"
	publisherRetry   = 500 * time.Millisecond
	subscribeBufSize = 1 << 20 // 1MB max message
)

// registerSubscribeMethod adds mcp:subscribe(topic, handler) to the mcp Lua global.
func (s *Server) registerSubscribeMethod(vendedID string, mcpTable *lua.LTable) {
	session := s.UiServer.GetLuaSession(vendedID)
	if session == nil {
		return
	}
	L := session.State

	L.SetField(mcpTable, "subscribe", L.NewFunction(func(L *lua.LState) int {
		// mcp:subscribe(topic, handler, opts) — arg 1 is self
		topic := L.CheckString(2)
		handler := L.CheckFunction(3)

		// Optional opts table with favicon field
		var favicon string
		if opts, ok := L.Get(4).(*lua.LTable); ok {
			if fav := opts.RawGetString("favicon"); fav != lua.LNil {
				favicon = fav.String()
			}
		}

		go s.pollLoop(vendedID, topic, handler, favicon)
		return 0
	}))
}

// pollLoop long-polls the publisher for messages on a topic and calls the Lua handler.
// The publisher is co-hosted by the MCP server; on connection error, retries after a short delay.
// The favicon is sent as a query parameter on the first request only.
func (s *Server) pollLoop(vendedID, topic string, handler *lua.LFunction, favicon string) {
	baseURL := fmt.Sprintf("%s/subscribe/%s", publisherAddr, topic)

	// Build first URL with favicon query param, then use baseURL for all subsequent requests
	nextURL := baseURL
	if favicon != "" {
		nextURL += "?favicon=" + url.QueryEscape(favicon)
	}

	for {
		resp, err := http.Get(nextURL)
		nextURL = baseURL

		if err != nil {
			time.Sleep(publisherRetry)
			continue
		}

		s.handlePollResponse(resp, vendedID, topic, handler)
	}
}

// handlePollResponse processes a single long-poll response and dispatches to the Lua handler.
func (s *Server) handlePollResponse(resp *http.Response, vendedID, topic string, handler *lua.LFunction) {
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNoContent:
		// Poll timeout, will reconnect on next iteration

	case http.StatusOK:
		body, err := io.ReadAll(io.LimitReader(resp.Body, subscribeBufSize))
		if err != nil {
			log.Printf("subscribe %s: read error: %v", topic, err)
			return
		}

		var data interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			log.Printf("subscribe %s: JSON parse error: %v", topic, err)
			return
		}

		s.callHandler(vendedID, handler, data)

	default:
		log.Printf("subscribe %s: unexpected status %d", topic, resp.StatusCode)
		time.Sleep(publisherRetry)
	}
}

// callHandler executes the Lua handler function in the session context with the parsed data.
func (s *Server) callHandler(vendedID string, handler *lua.LFunction, data interface{}) {
	_, err := s.SafeExecuteInSession(vendedID, func() (interface{}, error) {
		session := s.UiServer.GetLuaSession(vendedID)
		if session == nil {
			return nil, fmt.Errorf("session %s gone", vendedID)
		}
		L := session.State
		luaVal := session.GoToLua(data)
		return nil, L.CallByParam(lua.P{
			Fn:      handler,
			NRet:    0,
			Protect: true,
		}, luaVal)
	})
	if err != nil {
		log.Printf("subscribe handler error: %v", err)
	}
}
