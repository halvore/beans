---
# beans-yi9w
title: Consolidate agent store subscription ownership
status: todo
type: bug
priority: high
created_at: 2026-03-14T15:06:19Z
updated_at: 2026-03-14T15:06:19Z
parent: beans-5txp
---

Both WorkspaceView and AgentChat can independently create AgentChatStore instances and subscriptions for the same bean. This is confusing and potentially wasteful (duplicate WebSocket subscriptions).

## Current Behavior

- WorkspaceView creates agentStore = new AgentChatStore() and subscribes
- AgentChat creates its own ownStore = new AgentChatStore() as fallback
- If AgentChat doesn't receive externalStore, it creates a duplicate subscription

## Proposed Fix

Establish clear ownership: one component creates and owns the subscription, passes the store down. AgentChat should always receive its store as a prop, never create its own.

## Affected Files

- frontend/src/lib/components/WorkspaceView.svelte
- frontend/src/lib/components/AgentChat.svelte
