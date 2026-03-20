---
# beans-mwfn
title: Agent sessions in TUI
status: completed
type: epic
priority: normal
tags:
    - idea
created_at: 2026-03-20T13:35:34Z
updated_at: 2026-03-20T14:17:34Z
---

Add agent session support to the Bubbletea TUI


## Requirements

- Initialize `AgentManager` in the TUI (like `serve.go` does), calling `agentMgr.SendMessage()` directly (no GraphQL needed since everything is in-process)
- **Minimized agent panel** in the bottom-right corner showing:
  - Agent status: idle, working on X, needs input
  - Current bean context
- **Hotkey** to toggle between:
  - Minimized corner panel (default) — shows status only
  - Focused agent view — full chat interface with message history, streaming output, and input
- Handle blocking interactions (plan approval, `AskUserQuestion`) via TUI modals/prompts
- Render streamed markdown output from the agent in the chat view
- Support sending messages to the agent
- **Agent-driven status management**: Agenten som implementerer en oppgave bestemmer selv utfallet:
  - Sett til `completed` hvis oppgaven er løst
  - Sett til `scrapped` hvis den ikke kan/bør løses
  - Sett til `draft` hvis den ikke er klar for implementering ennå
- **Worktree-modus** (konfigurerbar innstilling i TUI):
  - **Direkte modus** (standard): Agenten kjører rett i prosjektet — kun én agent om gangen
  - **Worktree-modus**: Agenten kjører i en egen git worktree — støtter multi-agent, flere oppgaver kan jobbes på parallelt

## Technical Notes

- The TUI already creates a `graph.Resolver` with `Core` but does not set `AgentMgr` — this needs to be initialized
- Agent manager can be used directly without the GraphQL layer
- Reference `internal/commands/serve.go` for full `AgentManager` initialization
- Reference `internal/agent/manager.go` for session management and pub/sub
- Bubbletea components needed: chat message list, input field, status bar widget, modal for blocking interactions
- Worktree-støtte finnes allerede i `internal/worktree/` — TUI-en trenger bare å kalle `startWork`-logikken for å opprette worktrees per bean
- Innstilling for worktree vs. direkte modus kan lagres i `.beans.yml` eller som en TUI-runtime toggle

## Summary of Changes
All 8 sub-tasks completed. The TUI now supports:
- Agent session management via direct AgentManager integration
- Status panel in bottom-right showing running agents and their state
- Full agent chat view with streaming markdown, message input, and tool display
- ctrl+a hotkey to toggle between chat and previous view
- Blocking interaction modals for plan approval and agent questions
- Agent-driven bean status updates via file watcher
- Configurable worktree mode for multi-agent support
- Start agent from list or detail view with 'a' key
