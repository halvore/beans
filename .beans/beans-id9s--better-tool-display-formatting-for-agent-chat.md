---
# beans-id9s
title: Better tool display formatting for agent chat
status: completed
type: task
priority: normal
created_at: 2026-03-10T15:15:44Z
updated_at: 2026-03-10T15:44:18Z
order: zzzzzs
---

Add proper formatting for ToolSearch, EnterWorktree, ExitWorktree, Agent, Skill, WebSearch, WebFetch in formatToolInput(). Improve default fallback.

## Notes

No e2e tests added — formatToolInput is only used for permission request display, which requires a live agent session (can't be seeded via JSONL). TOOL messages in the chat use backend-extracted summaries.

## Summary of Changes

Added proper formatting for ToolSearch, WebSearch, WebFetch, Agent, Skill, EnterWorktree, and ExitWorktree in `formatToolInput()`. Replaced raw JSON default with smart field extraction matching backend priority list.
