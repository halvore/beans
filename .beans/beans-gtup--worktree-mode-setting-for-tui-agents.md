---
# beans-gtup
title: Worktree mode setting for TUI agents
status: completed
type: task
priority: normal
created_at: 2026-03-20T13:48:44Z
updated_at: 2026-03-20T14:17:34Z
parent: beans-mwfn
---

Add a configurable setting (in .beans.yml or TUI toggle) for agent execution mode: direct (runs in project dir, single agent) or worktree (each agent gets its own git worktree, enables multi-agent). Use existing internal/worktree/ package.

## Summary of Changes
- Added WorktreeMode bool to AgentConfig in pkg/config/config.go
- Added IsWorktreeMode() helper method
- TUI command creates worktree.Manager when worktree_mode is true
- startAgentMsg handler creates worktree per bean in worktree mode
- In worktree mode, multiple agents can run concurrently (no stop-previous logic)
- In direct mode (default), single-agent with stop-previous behavior
