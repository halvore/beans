---
# beans-541b
title: Consolidate duplicate ETag error types
status: todo
type: task
priority: normal
created_at: 2026-03-13T17:49:02Z
updated_at: 2026-03-13T17:49:02Z
parent: beans-oyic
---

ETagMismatchError and ETagRequiredError are defined in both internal/graph/resolver.go and pkg/beancore/core.go. Consolidate to one location (beancore, since it's the domain layer).
