# Sequence: MCP State Change Waiting

**Source Spec:** specs/mcp.md (Section 8)

## Participants
- Agent: AI assistant (Claude Code) or background script
- WaitScript: scripts/wait-for-state.sh polling loop
- MCPServer: HTTP server handling wait endpoint (uses currentVendedID)
- UIServer: UI platform server providing ExecuteInSession with afterBatch browser updates
- Session: Lua session with mcp.state queue
- LuaCode: User's Lua application code (in browser)

## Scenario 1: Agent Waits for State Change (Single Event)

Agent starts background script that long-polls for state changes. User interaction
in browser calls mcp.pushState() to queue an event, triggering response.

```
     +-------+        +------------+        +-----------+        +----------+        +---------+        +---------+
     | Agent |        | WaitScript |        | MCPServer |        | UIServer |        | Session |        | LuaCode |
     +---+---+        +-----+------+        +-----+-----+        +-----+----+        +----+----+        +----+----+
         |                  |                     |                    |                  |                  |
         | Bash(run_in_background=true)           |                    |                  |                  |
         | wait-for-state.sh <url> 30             |                    |                  |                  |
         |----------------->|                     |                    |                  |                  |
         |                  |                     |                    |                  |                  |
         |                  | GET /wait?timeout=30|                    |                  |                  |
         |                  |-------------------->|                    |                  |                  |
         |                  |                     |                    |                  |                  |
         |                  |                     | GetCurrentSession  |                  |                  |
         |                  |                     |----------------------------------->|                  |
         |                  |                     |                    |                  |                  |
         |                  |                     | Check #mcp.state   |                  |                  |
         |                  |                     |----------------------------------->|                  |
         |                  |                     | (queue empty)      |                  |                  |
         |                  |                     |<-----------------------------------|                  |
         |                  |                     |                    |                  |                  |
         |                  |                     | AddWaiter(chan)    |                  |                  |
         |                  |                     |-----+              |                  |                  |
         |                  |                     |     | (blocks)     |                  |                  |
         |                  |                     |<----+              |                  |                  |
         |                  |                     |                    |                  |   (user clicks)  |
         |                  |                     |                    |                  |<-----------------|
         |                  |                     |                    |                  |                  |
         |                  |                     |                    |                  | mcp.pushState({  |
         |                  |                     |                    |                  |   app="contacts",|
         |                  |                     |                    |                  |   event="chat",  |
         |                  |                     |                    |                  |   text="hello"}) |
         |                  |                     |                    |                  |<-----------------|
         |                  |                     |                    |                  |                  |
         |                  |                     | NotifyStateChange  |                  |                  |
         |                  |                     |<-----------------------------------|                  |
         |                  |                     |                    |                  |                  |
         |                  |                     | AtomicSwapQueue()  |                  |                  |
         |                  |                     |----------------------------------->|                  |
         |                  |                     | (swap with {})     |                  |                  |
         |                  |                     |<--[{app:..}]-------|                  |                  |
         |                  |                     |                    |                  |                  |
         |                  |                     | SafeExecuteInSession(empty fn)       |                  |
         |                  |                     |------------------->|                  |                  |
         |                  |                     |                    | afterBatch()     |                  |
         |                  |                     |                    |-----+            |                  |
         |                  |                     |                    |<----+ (push)     |                  |
         |                  |                     |<-------------------|                  |                  |
         |                  |                     |                    |                  |                  |
         |                  | 200 OK [{...}]      |                    |                  |                  |
         |                  |<--------------------|                    |                  |                  |
         |                  |                     |                    |                  |                  |
         |                  | jq -c '.[]'         |                    |                  |                  |
         |                  | echo each event     |                    |                  |                  |
         |                  |-----+               |                    |                  |                  |
         |                  |<----+               |                    |                  |                  |
         |                  |                     |                    |                  |                  |
         | (TaskOutput: {"app":"contacts","event":"chat","text":"hello"})                |                  |
         |<-----------------|                     |                    |                  |                  |
         |                  |                     |                    |                  |                  |
     +---+---+        +-----+------+        +-----+-----+        +-----+----+        +----+----+        +----+----+
     | Agent |        | WaitScript |        | MCPServer |        | UIServer |        | Session |        | LuaCode |
     +-------+        +------------+        +-----------+        +----------+        +---------+        +---------+
```

## Scenario 2: Multiple Events Accumulated

Multiple events pushed before wait response - all returned in single array.

```
     +-------+        +------------+        +-----------+        +----------+        +---------+        +---------+
     | Agent |        | WaitScript |        | MCPServer |        | UIServer |        | Session |        | LuaCode |
     +---+---+        +-----+------+        +-----+-----+        +-----+----+        +----+----+        +----+----+
         |                  |                     |                    |                  |                  |
         |                  | GET /wait?timeout=30|                    |                  |                  |
         |                  |-------------------->|                    |                  |                  |
         |                  |                     | AddWaiter(chan)    |                  |                  |
         |                  |                     |-----+ (blocks)     |                  |                  |
         |                  |                     |<----+              |                  |                  |
         |                  |                     |                    |                  |                  |
         |                  |                     |                    |                  | mcp.pushState({  |
         |                  |                     |                    |                  |   app="c",       |
         |                  |                     |                    |                  |   event="btn",   |
         |                  |                     |                    |                  |   id="save"})    |
         |                  |                     |                    |                  |<-----------------|
         |                  |                     |                    |                  |                  |
         |                  |                     | NotifyStateChange  |                  |                  |
         |                  |                     |<-----------------------------------|                  |
         |                  |                     |                    |                  |                  |
         |                  |                     |                    |                  | mcp.pushState({  |
         |                  |                     |                    |                  |   app="c",       |
         |                  |                     |                    |                  |   event="btn",   |
         |                  |                     |                    |                  |   id="cancel"})  |
         |                  |                     |                    |                  |<-----------------|
         |                  |                     |                    |                  |                  |
         |                  |                     | NotifyStateChange  |                  |                  |
         |                  |                     |<-----------------------------------|                  |
         |                  |                     |                    |                  |                  |
         |                  |                     | AtomicSwapQueue()  |                  |                  |
         |                  |                     |----------------------------------->|                  |
         |                  |                     | (swap with {})     |                  |                  |
         |                  |                     |<--[{..},{..}]------|                  |                  |
         |                  |                     |                    |                  |                  |
         |                  |                     | SafeExecuteInSession(empty fn)       |                  |
         |                  |                     |------------------->|                  |                  |
         |                  |                     |                    | afterBatch()     |                  |
         |                  |                     |                    |-----+ (push)     |                  |
         |                  |                     |                    |<----+            |                  |
         |                  |                     |<-------------------|                  |                  |
         |                  |                     |                    |                  |                  |
         |                  | 200 OK [{...},{...}]|                    |                  |                  |
         |                  |<--------------------|                    |                  |                  |
         |                  |                     |                    |                  |                  |
         |                  | jq -c '.[]'         |                    |                  |                  |
         |                  | (outputs 2 lines)   |                    |                  |                  |
         |                  |-----+               |                    |                  |                  |
         |                  |<----+               |                    |                  |                  |
         |                  |                     |                    |                  |                  |
         | (TaskOutput line 1: {"app":"c","event":"btn","id":"save"})  |                  |                  |
         | (TaskOutput line 2: {"app":"c","event":"btn","id":"cancel"})|                  |                  |
         |<-----------------|                     |                    |                  |                  |
         |                  |                     |                    |                  |                  |
     +---+---+        +-----+------+        +-----+-----+        +-----+----+        +----+----+        +----+----+
     | Agent |        | WaitScript |        | MCPServer |        | UIServer |        | Session |        | LuaCode |
     +-------+        +------------+        +-----------+        +----------+        +---------+        +---------+
```

## Scenario 3: Wait Timeout (Empty Queue)

Agent waits but no events are pushed within timeout period.

```
     +-------+        +------------+        +-----------+
     | Agent |        | WaitScript |        | MCPServer |
     +---+---+        +-----+------+        +-----+-----+
         |                  |                     |
         | Bash(run_in_background=true)           |
         |----------------->|                     |
         |                  |                     |
         |                  | GET /wait?timeout=5 |
         |                  |-------------------->|
         |                  |                     |
         |                  |                     | AddWaiter(chan)
         |                  |                     |-----+
         |                  |                     |     | (5s passes)
         |                  |                     |<----+
         |                  |                     |
         |                  | 204 No Content      |
         |                  |<--------------------|
         |                  |                     |
         |                  | (loop continues)    |
         |                  | GET /wait?timeout=5 |
         |                  |-------------------->|
         |                  |                     |
     +---+---+        +-----+------+        +-----+-----+
     | Agent |        | WaitScript |        | MCPServer |
     +-------+        +------------+        +-----------+
```

## Scenario 4: No Active Session

Wait request when server has no active session (not RUNNING).

```
     +------------+        +-----------+
     | WaitScript |        | MCPServer |
     +-----+------+        +-----+-----+
           |                     |
           | GET /wait?timeout=30|
           |-------------------->|
           |                     |
           |                     | GetCurrentSession
           |                     |-----+
           |                     |     | (none)
           |                     |<----+
           |                     |
           | 404 Not Found       |
           |<--------------------|
           |                     |
           | exit 1              |
           |-----+               |
           |<----+               |
     +-----+------+        +-----+-----+
     | WaitScript |        | MCPServer |
     +------------+        +-----------+
```

## Scenario 5: Events Queued Before Wait

Events pushed before agent starts waiting - returned immediately.

```
     +-----------+        +----------+        +---------+        +---------+        +------------+
     | MCPServer |        | UIServer |        | Session |        | LuaCode |        | WaitScript |
     +-----+-----+        +-----+----+        +----+----+        +----+----+        +-----+------+
           |                    |                  |                  |                    |
           |                    |                  | mcp.pushState({  |                    |
           |                    |                  |   app="x",...})  |                    |
           |                    |                  |<-----------------|                    |
           |                    |                  |                  |                    |
           | (no waiters)       |                  |                  |                    |
           |<-----------------------------------|                  |                    |
           |                    |                  |                  |                    |
           |                    |                  |                  |   GET /wait?t=30   |
           |<-----------------------------------------------------------+--------------|
           |                    |                  |                  |                    |
           | Check #mcp.state   |                  |                  |                    |
           |---------------------------------->|                  |                    |
           | (queue has items)  |                  |                  |                    |
           |<----------------------------------|                  |                    |
           |                    |                  |                  |                    |
           | AtomicSwapQueue()  |                  |                  |                    |
           |---------------------------------->|                  |                    |
           | (swap with {})     |                  |                  |                    |
           |<--[{app:"x",...}]--|                  |                  |                    |
           |                    |                  |                  |                    |
           | SafeExecuteInSession(empty fn)        |                  |                    |
           |------------------->|                  |                  |                    |
           |                    | afterBatch()     |                  |                    |
           |                    |-----+ (push)     |                  |                    |
           |                    |<----+            |                  |                    |
           |<-------------------|                  |                  |                    |
           |                    |                  |                  |                    |
           | 200 OK [{...}]     |                  |                  |                    |
           |-----------------------------------------------------------+-------------->|
           |                    |                  |                  |                    |
     +-----+-----+        +-----+----+        +----+----+        +----+----+        +-----+------+
     | MCPServer |        | UIServer |        | Session |        | LuaCode |        | WaitScript |
     +-----------+        +----------+        +---------+        +---------+        +------------+
```

## Implementation Notes

- Wait endpoint: `GET /wait?timeout=N` (max 120 seconds, default 30)
- Uses server's currentVendedID (single distinguished session)
- `mcp.state` is initialized as empty table `{}` on session start
- Events pushed via `mcp.pushState({...})` - Lua function that queues and signals
- When wait responds, atomically swap queue with empty table in Lua context
- Atomic swap ensures no events lost between read and subsequent writes
- Multiple waiters supported (broadcast on state change)
- Waiter cleanup on timeout, client disconnect, or session destroy
- Response is JSON array of all accumulated events
- Script uses `jq -c '.[]'` to output each event as compact JSON on its own line
- **Browser Update:** After draining the queue, calls `SafeExecuteInSession` with an empty function to trigger `afterBatch`, ensuring UIs monitoring the event queue refresh (see Section 4.1 of mcp.md)
