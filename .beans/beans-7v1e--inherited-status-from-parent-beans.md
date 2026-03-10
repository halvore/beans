---
# beans-7v1e
title: Inherited status from parent beans
status: completed
type: feature
priority: normal
created_at: 2026-03-08T22:09:06Z
updated_at: 2026-03-10T08:08:34Z
order: zzzk
---

Add inherited/effective status that propagates terminal statuses (scrapped, completed) from parent beans to children transitively. Children keep their explicit status unchanged, but expose an inherited status for display and filtering purposes.

## Summary of Changes

Implemented inherited status: Core.InheritedStatus(), GraphQL fields, ready filter exclusion, list/show/TUI display with ↑status annotation
