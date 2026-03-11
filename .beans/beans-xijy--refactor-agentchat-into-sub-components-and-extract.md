---
# beans-xijy
title: Refactor AgentChat into sub-components and extract shared RenderedMarkdown
status: completed
type: task
priority: normal
created_at: 2026-03-11T18:55:27Z
updated_at: 2026-03-11T18:58:23Z
---

Split AgentChat.svelte (493 lines) into focused sub-components and extract a shared RenderedMarkdown component used by both AgentChat and BeanDetail.

## Tasks
- [x] Create RenderedMarkdown.svelte — shared markdown renderer with bean-link click handling
- [x] Create AgentMessages.svelte — message list with auto-scroll and streaming
- [x] Create PendingInteraction.svelte — plan approval and AskUserQuestion UI
- [x] Create AgentComposer.svelte — input, send, stop, mode toggle, compact/clear
- [x] Refactor AgentChat.svelte as thin orchestrator
- [x] Update BeanDetail.svelte to use RenderedMarkdown
- [x] Verify build has no warnings

## Summary of Changes

Split AgentChat.svelte (493 lines) into 4 focused components:

- **RenderedMarkdown.svelte** — shared async markdown renderer with bean-link click handling, used by both AgentChat and BeanDetail
- **AgentMessages.svelte** — message list with auto-scroll, streaming markdown rendering, and subagent activity display
- **PendingInteraction.svelte** — EXIT_PLAN approval and ASK_USER structured question UI
- **AgentComposer.svelte** — textarea input with localStorage persistence, send/stop buttons, plan/act mode toggle, compact/clear actions
- **AgentChat.svelte** — now a thin orchestrator (~80 lines) composing the three sub-components
- **BeanDetail.svelte** — replaced inline markdown rendering + bean-link handler with RenderedMarkdown component, removing ~20 lines of duplicated code
