---
# beans-u2xe
title: Initialize AgentManager in TUI
status: completed
type: task
priority: normal
created_at: 2026-03-20T13:48:31Z
updated_at: 2026-03-20T14:15:08Z
parent: beans-mwfn
---

Wire up AgentManager in the TUI's New() function, similar to how serve.go does it. The TUI model should hold a reference to the agent manager so other components can use it. No UI changes needed — just the plumbing.

## Summary of Changes
- Added agent.Manager creation in internal/commands/tui.go with context provider
- Updated tui.New() and tui.Run() to accept *agent.Manager
- Set up global subscription channel and waitForAgentUpdate/waitForSessionUpdate commands
- Wired agentMgr.SetOnTurnComplete to refresh bean list on agent turn completion
- Added agentMgr.Shutdown() and UnsubscribeGlobal cleanup on exit
