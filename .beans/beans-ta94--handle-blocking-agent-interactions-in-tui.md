---
# beans-ta94
title: Handle blocking agent interactions in TUI
status: completed
type: task
priority: normal
created_at: 2026-03-20T13:48:39Z
updated_at: 2026-03-20T14:15:35Z
parent: beans-mwfn
blocked_by:
    - beans-71bz
---

When the agent hits a blocking interaction (AskUserQuestion, plan approval, ExitPlanMode), show a TUI modal/prompt so the user can respond. The agent panel should change status to 'needs input' to draw attention.

## Summary of Changes
- Created internal/tui/agentinteraction.go with interactionModel
- Handles ExitPlanMode (plan approval with content preview), EnterPlanMode, AskUserQuestion
- Option selection with j/k navigation, enter to confirm, esc to dismiss
- Rendered as centered modal overlay using existing overlayModal pattern
- Auto-detected from agentUpdatedMsg and agentSessionMsg when PendingInteraction is set
