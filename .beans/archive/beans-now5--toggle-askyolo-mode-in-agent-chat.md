---
# beans-now5
title: Toggle Ask/Yolo mode in agent chat
status: completed
type: feature
priority: normal
created_at: 2026-03-09T18:04:46Z
updated_at: 2026-03-10T08:12:46Z
order: V0N
---

Add a toggle in the agent chat UI that switches the Claude Code CLI session between Ask mode (requires confirmation for tool use) and Yolo mode (auto-accepts tool calls). This gives users control over how autonomously the agent operates. Consider a toggle button in the chat header or composer area with clear visual indication of the current mode.

## Summary of Changes

Added YOLO mode toggle to agent chat, mirroring the existing Plan mode pattern:

- **Backend**: Added `YoloMode` field to `Session` struct, `SetYoloMode` method to Manager, `--dangerously-skip-permissions` flag in `buildClaudeArgs` (YOLO takes precedence over Plan mode)
- **GraphQL**: Added `yoloMode` field to `AgentSession` type and `setAgentYoloMode` mutation
- **Frontend**: Added `yoloMode` to store subscription/interface, `setYoloMode` mutation method, Ask/YOLO toggle buttons (red danger styling for YOLO)
- **Tests**: 6 new tests covering SetYoloMode lifecycle and buildClaudeArgs with YOLO flag
