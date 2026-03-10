---
# beans-xvs4
title: Inject bean context into agent conversation
status: completed
type: feature
priority: normal
created_at: 2026-03-10T12:14:45Z
updated_at: 2026-03-10T15:44:18Z
order: zzzzy
---

Automatically inject bean context (title, type, status, body) into the Claude Code process stdin before the first user message, so the agent knows what it's working on without the user having to explain. Only on first spawn (not resume), skip for __central__ sessions.

## Summary of Changes

- Added `ContextProvider` callback type and field to `Manager` struct
- Updated `NewManager` to accept a `ContextProvider` parameter
- In `spawnAndRun`, inject bean context via stdin before the first user message (only on fresh sessions, not `--resume`)
- Wired up the context provider in `serve.go` to format bean ID, title, type, status, priority, and body
- Updated all `NewManager` call sites in tests (`manager_test.go`, `store_test.go`)

### Files modified
- `internal/agent/manager.go`
- `internal/agent/claude.go`
- `internal/commands/serve.go`
- `internal/agent/manager_test.go`
- `internal/agent/store_test.go`
