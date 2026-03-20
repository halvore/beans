---
# beans-u2xe
title: Initialize AgentManager in TUI
status: todo
type: task
created_at: 2026-03-20T13:48:31Z
updated_at: 2026-03-20T13:48:31Z
parent: beans-mwfn
---

Wire up AgentManager in the TUI's New() function, similar to how serve.go does it. The TUI model should hold a reference to the agent manager so other components can use it. No UI changes needed — just the plumbing.
