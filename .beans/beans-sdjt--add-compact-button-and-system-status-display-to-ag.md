---
# beans-sdjt
title: Add Compact button and system status display to agent chat
status: completed
type: feature
priority: normal
created_at: 2026-03-10T08:48:48Z
updated_at: 2026-03-10T15:44:18Z
order: zzzzk
---

Added Compact button (sends /compact to Claude Code), system status display (shows 'compacting...' instead of 'thinking...'), frontend debug logging of agent messages, and grouped Compact/Clear button styling.

## Summary of Changes

### Compact Button
- Sends `/compact` as a user message to trigger Claude Code's built-in context compaction
- Grouped with Clear button using joined tab styling (rounded-l / rounded-r)

### System Status Display
- Parse `system` events with `subtype: "status"` from Claude Code's stream-json output
- Added `SystemStatus` field to Session, GraphQL schema, and frontend
- Shows "compacting..." (or any future status) instead of "thinking..." in both the message area and composer bar

### Debug Logging
- Backend: log unhandled stream-json events for discovering new event types
- Frontend: log user/tool messages immediately, assistant messages on turn completion (RUNNING → IDLE), system status changes, and errors to browser console

### Rules Updates
- Added GraphQL subscription nil-handling rule to CLAUDE.md
- Added agent chat e2e testing strategy to frontend rules
