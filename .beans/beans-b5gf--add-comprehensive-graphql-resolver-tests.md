---
# beans-b5gf
title: Add comprehensive GraphQL resolver tests
status: todo
type: task
priority: high
created_at: 2026-03-14T15:06:19Z
updated_at: 2026-03-14T15:06:19Z
parent: beans-5txp
---

The GraphQL resolvers are the API contract between frontend and backend, but they're essentially untested. This is the most impactful test gap in the codebase.

## Scope

- Table-driven tests for all mutations (create, update, delete)
- Relationship validation tests (parent type hierarchy, blocking cycles)
- Concurrent ETag conflict scenarios
- Subscription lifecycle tests (connect, receive, cleanup)
- Filter/query tests for edge cases

## Affected Files

- internal/graph/schema.resolvers_test.go (needs major expansion)
