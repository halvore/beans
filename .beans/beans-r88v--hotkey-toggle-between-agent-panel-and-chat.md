---
# beans-r88v
title: Hotkey toggle between agent panel and chat
status: completed
type: task
priority: normal
created_at: 2026-03-20T13:48:37Z
updated_at: 2026-03-20T14:15:34Z
parent: beans-mwfn
blocked_by:
    - beans-u2xe
    - beans-qa4i
    - beans-71bz
---

Add a hotkey (e.g. Ctrl+A or similar) that toggles between the minimized agent status panel and the full agent chat view. When toggling back to list view, the agent keeps running in the background.

## Summary of Changes
- ctrl+a toggles between current view and agent chat
- From list/detail: opens chat for current bean's agent or most recently active agent
- From agent chat: returns to previous view
- Agent keeps running in background
