---
# beans-qa4i
title: Agent status panel (minimized view)
status: completed
type: task
priority: normal
created_at: 2026-03-20T13:48:33Z
updated_at: 2026-03-20T14:15:15Z
parent: beans-mwfn
blocked_by:
    - beans-u2xe
---

Add a small status widget in the bottom-right corner of the TUI. Always visible when agent sessions exist. No interactivity — display only.

Must show:
- How many agents are currently running
- Per-agent summary: bean title + state (working, needs input, idle)
- Compact enough to fit in a corner without disrupting the main view

## Summary of Changes
- Created internal/tui/agentpanel.go with agentPanelModel
- Panel shows agent count and per-agent status (running/needs input/idle/error)
- Bottom-right overlay compositing using existing overlayLine helper
- Refreshes on agentUpdatedMsg from global subscription
- Resolves bean titles via resolver for display
