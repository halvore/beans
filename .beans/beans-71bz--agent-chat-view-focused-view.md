---
# beans-71bz
title: Agent chat view (focused view)
status: completed
type: task
priority: normal
created_at: 2026-03-20T13:48:34Z
updated_at: 2026-03-20T14:15:34Z
parent: beans-mwfn
blocked_by:
    - beans-u2xe
---

Full-screen agent chat view with message history and streaming output. Render agent messages as markdown. Include a text input for sending messages. Toggled via hotkey from the minimized panel.

## Summary of Changes
- Created internal/tui/agentchat.go with agentChatModel
- Full-screen chat with viewport for message history and textinput for composing
- Message rendering: user (blue >), assistant (glamour markdown), tool (dimmed), info (italic)
- Per-bean subscription via waitForSessionUpdate for real-time updates
- Auto-scrolls to bottom on new content
- Keybindings: enter=send, esc=back, ctrl+s=stop agent
