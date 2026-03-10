---
# beans-v77a
title: Add Clear button to agent conversations
status: completed
type: feature
priority: normal
created_at: 2026-03-10T08:15:02Z
updated_at: 2026-03-10T15:44:18Z
order: zzzzV
---

Add a Clear button next to the plan/act/yolo mode switcher in agent conversations. It should clear the conversation history (messages, session state, and persisted JSONL). Needs backend ClearSession method + GraphQL mutation + frontend button.

## Summary of Changes

### Backend
- Added `store.clear()` method to delete persisted JSONL conversation file
- Added `Manager.ClearSession()` method that stops the process, removes the in-memory session, and clears persistence
- Added `clearAgentSession` GraphQL mutation
- Added 3 unit tests for ClearSession (removes session, notifies subscribers, handles nonexistent)

### Frontend
- Added `CLEAR_AGENT_SESSION` GraphQL mutation and `clearSession()` method to AgentChatStore
- Added Clear button next to the plan/act/yolo mode switcher with matching `btn-tab-sm` design
- Button is disabled when agent is running or conversation is empty

### Bug Fixes (follow-up)
- Fixed subscription resolver skipping `nil` sessions — now sends an empty session so the UI resets reactively
- Added global `button:disabled { cursor: not-allowed }` CSS rule
- Removed redundant cursor classes from Clear button

### E2E Test
- Added `agent-chat.spec.ts` that seeds a JSONL conversation file, opens the chat, clicks Clear, and verifies the UI resets (no Claude process spawned)
