---
# beans-m0fp
title: Restore animate-pulse on subagent activity lines
status: completed
type: bug
priority: normal
created_at: 2026-03-17T13:55:03Z
updated_at: 2026-03-17T13:59:51Z
---

The subagent activity output lines in the agent chat no longer pulse. They should have animate-pulse while active.

## Summary of Changes

Added `animate-pulse` class to subagent activity lines in AgentMessages.svelte to restore the pulsing animation while subagents are working.
