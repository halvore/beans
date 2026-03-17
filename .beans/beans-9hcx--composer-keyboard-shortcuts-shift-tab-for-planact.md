---
# beans-9hcx
title: 'Composer keyboard shortcuts: Shift-Tab for plan/act toggle, Escape for stop'
status: completed
type: feature
priority: normal
created_at: 2026-03-17T11:12:25Z
updated_at: 2026-03-17T11:12:45Z
---

Add keyboard shortcuts to the agent composer: Shift-Tab toggles between Plan and Act mode, Escape stops the active agent.

## Summary of Changes

Added two keyboard shortcuts to `AgentComposer.svelte`'s `handleKeydown` handler:
- **Shift+Tab**: toggles between Plan and Act mode (disabled while agent is running)
- **Escape**: stops the active agent (only when agent is running)
