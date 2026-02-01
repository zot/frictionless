# Sequence: App Audit

**Source Spec:** specs/ui-audit.md

## Participants

- Agent: Claude agent (foreground or background)
- Server: MCPServer (tools.go, server.go)
- Auditor: Audit logic (audit.go)
- FS: Filesystem

## MCP Tool Flow

```
Agent                    Server                   Auditor                  FS
  |                        |                        |                       |
  |-- ui_audit(name) ----->|                        |                       |
  |                        |-- AuditApp(base,name)->|                       |
  |                        |                        |-- Read app.lua ------>|
  |                        |                        |<-- content -----------|
  |                        |                        |                       |
  |                        |                        |-- analyzeLua() ------>|
  |                        |                        |   (extract defs,      |
  |                        |                        |    calls, guards,     |
  |                        |                        |    factory functions, |
  |                        |                        |    outer scope calls) |
  |                        |                        |                       |
  |                        |                        |-- ReadDir viewdefs -->|
  |                        |                        |<-- file list ---------|
  |                        |                        |                       |
  |                        |                        |   for each .html:     |
  |                        |                        |-- Read file --------->|
  |                        |                        |<-- content -----------|
  |                        |                        |-- html.Parse() ------>|
  |                        |                        |-- walkDOM() --------->|
  |                        |                        |   (check violations)  |
  |                        |                        |                       |
  |                        |                        |-- findDeadMethods() ->|
  |                        |                        |   (lua defs unused,   |
  |                        |                        |    excludes factory-  |
  |                        |                        |    created methods)   |
  |                        |                        |                       |
  |                        |                        |-- findMissingMethods()|
  |                        |                        |   (viewdef calls w/o  |
  |                        |                        |    matching lua def)  |
  |                        |                        |                       |
  |                        |<-- AuditResult --------|                       |
  |<-- JSON response ------|                        |                       |
  |                        |                        |                       |
```

## HTTP API Flow

```
Agent                    Server                   Auditor
  |                        |                        |
  |-- POST /api/ui_audit ->|                        |
  |   {"name": "app"}      |                        |
  |                        |-- AuditApp() --------->|
  |                        |   (same as above)      |
  |                        |<-- AuditResult --------|
  |<-- JSON response ------|                        |
  |                        |                        |
```

## DOM Walk Detail

```
Auditor                          html.Node
  |                                  |
  |-- walkDOM(node, isListItem) ---->|
  |                                  |
  |   if node.Type == ElementNode:   |
  |     check element tag            |
  |     for each attr:               |
  |       if ui-action:              |
  |         checkButtonElement()     |
  |       if ui-class has "hidden:": |
  |         addViolation()           |
  |       if ui-value on checkbox:   |
  |         addViolation()           |
  |       if has operators:          |
  |         addViolation()           |
  |       extract method calls       |
  |                                  |
  |   if isListItem && tag==style:   |
  |     addViolation()               |
  |                                  |
  |   if isListItem && has "item.":  |
  |     addViolation()               |
  |                                  |
  |   for each child:                |
  |     walkDOM(child, isListItem)   |
  |                                  |
```
