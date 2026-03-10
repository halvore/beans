---
# beans-bk5e
title: 'Agent chat: plan mode toggle'
status: completed
type: feature
priority: normal
created_at: 2026-03-09T17:50:42Z
updated_at: 2026-03-09T18:10:07Z
order: V0V
---

Allow switching between regular (default) and plan permission modes in the agent chat UI. Plan mode makes the agent read-only — it can explore and reason but not edit files. The toggle should be in the chat header/composer area. Switching modes requires killing and respawning the claude CLI process since --permission-mode is a startup arg.

## Summary of Changes

### Backend (Go)
- Added `PlanMode`, `ToolInvocations`, `PendingInteraction` fields to `Session` (types.go)
- Added `ToolInvocation`, `PendingInteraction`, `InteractionType` types (types.go)
- `buildClaudeArgs` passes `--permission-mode plan` when plan mode is enabled (claude.go)
- `SetPlanMode` preserves `SessionID` for `--resume` continuity (manager.go)
- `SendMessage` clears `PendingInteraction` and `ToolInvocations` on new user message (manager.go)
- `blockingInteraction()` maps ExitPlanMode/EnterPlanMode/AskUserQuestion to PendingInteraction (claude.go)
- `handleBlockingTool()` kills process, preserves session ID, toggles plan mode, reads plan file (claude.go)
- `findPlanFilePath()` scans structured ToolInvocation data for plan files (claude.go)
- `readOutput()` tracks ToolInvocations with tool name and input summary (claude.go)
- 15+ new tests covering blocking tools, plan file discovery, mode toggling, snapshots (manager_test.go)

### GraphQL
- Added `PendingInteraction` type and `InteractionType` enum to schema
- `pendingInteraction` field on `AgentSession`
- `agentSessionToModel` maps PendingInteraction to GraphQL model (agent_helpers.go)

### Frontend (Svelte)
- Added `PendingInteraction` type, `planMode`, `pendingInteraction` to `AgentSession` (agentChat.svelte.ts)
- Added `SET_AGENT_PLAN_MODE` mutation and `setPlanMode` method to `AgentChatStore`
- Work/Plan tab toggle with blue/orange colors in AgentChat composer (AgentChat.svelte)
- Approval card for EXIT_PLAN/ENTER_PLAN with rendered plan content
- AskUserQuestion banner prompting user to reply
- Approve/Reject buttons for mode-switch interactions
