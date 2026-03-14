---
# beans-y0by
title: Add subscription error recovery in frontend stores
status: todo
type: bug
priority: high
created_at: 2026-03-14T15:06:19Z
updated_at: 2026-03-14T15:06:19Z
parent: beans-5txp
---

When a WebSocket subscription errors out, stores set an error state but never retry. A network blip kills the subscription permanently until the user refreshes.

The graphql-ws client already has exponential backoff for connection-level reconnects, but store-level resubscription isn't handled.

## Proposed Fix

Implement a resubscribe mechanism in each store (BeansStore, WorktreeStore, AgentChatStore, AgentStatusesStore) that detects subscription errors and re-establishes the subscription with backoff.

## Affected Files

- frontend/src/lib/stores/beansStore.svelte.ts
- frontend/src/lib/stores/worktreeStore.svelte.ts
- frontend/src/lib/stores/agentChat.svelte.ts
- frontend/src/lib/stores/agentStatusesStore.svelte.ts
