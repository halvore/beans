---
# beans-txc5
title: Start agent from bean detail view
status: completed
type: task
priority: normal
created_at: 2026-03-20T13:48:46Z
updated_at: 2026-03-20T14:15:35Z
parent: beans-mwfn
blocked_by:
    - beans-u2xe
    - beans-71bz
    - beans-qa4i
---

Add a keybinding in the bean detail view to start an agent session for the selected bean. Should send the bean context (title, description, body) as the initial message. Respects the worktree/direct mode setting.

## Summary of Changes
- Added 'a' keybinding in detail.go and list.go to emit startAgentMsg
- startAgentMsg handler builds context from bean title/type/status/body
- Enforces single-agent in direct mode (stops existing agent for different bean)
- Opens agent chat after starting agent
- Added to help overlay
