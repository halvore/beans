---
# beans-ta94
title: Handle blocking agent interactions in TUI
status: todo
type: task
priority: normal
created_at: 2026-03-20T13:48:39Z
updated_at: 2026-03-20T13:48:55Z
parent: beans-mwfn
blocked_by:
    - beans-71bz
---

When the agent hits a blocking interaction (AskUserQuestion, plan approval, ExitPlanMode), show a TUI modal/prompt so the user can respond. The agent panel should change status to 'needs input' to draw attention.
