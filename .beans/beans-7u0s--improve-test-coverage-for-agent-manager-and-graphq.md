---
# beans-7u0s
title: Improve test coverage for agent manager and GraphQL subscriptions
status: todo
type: task
priority: normal
created_at: 2026-03-13T17:49:21Z
updated_at: 2026-03-13T17:49:21Z
parent: beans-oyic
---

Agent manager (internal/agent/manager.go) has only ~100 lines of tests — no tests for concurrent session modifications, process lifecycle errors, or streaming failure modes. GraphQL subscription resolvers (especially the explicit nil payload behavior) have no dedicated tests. These are the highest-risk areas for regressions.
