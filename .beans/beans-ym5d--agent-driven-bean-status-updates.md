---
# beans-ym5d
title: Agent-driven bean status updates
status: completed
type: task
priority: normal
created_at: 2026-03-20T13:48:41Z
updated_at: 2026-03-20T14:15:35Z
parent: beans-mwfn
---

Let the agent update the bean status based on outcome: completed (solved), scrapped (not feasible), or draft (not ready). This should happen via the agent's system prompt instructions and/or tool access to beans CLI.

## Summary of Changes
- agentMgr.SetOnTurnComplete wired in Run() to send beansChangedMsg on agent turn completion
- File watcher in beancore.Core already picks up bean file changes made by agents
- Bean list refreshes automatically when agent modifies bean status
