---
# beans-kpvs
title: Centralize ID normalization logic
status: todo
type: task
created_at: 2026-03-14T15:06:45Z
updated_at: 2026-03-14T15:06:45Z
parent: beans-5txp
---

NormalizeID() is called everywhere — resolvers, validators, helpers — with repeated patterns. A single RequireFullID(id string) (string, error) method on Core, used consistently, would reduce noise and potential bugs.

## Current Pattern (repeated ~20+ times)

fullID, ok := c.NormalizeID(id)
if !ok {
    return fmt.Errorf("bean not found: %s", id)
}

## Proposed Pattern

fullID, err := c.RequireFullID(id)
if err != nil {
    return err
}

## Affected Files

- pkg/beancore/core.go
- internal/graph/schema.resolvers.go
- internal/graph/helpers.go
