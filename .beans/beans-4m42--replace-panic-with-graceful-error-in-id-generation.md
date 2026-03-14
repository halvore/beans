---
# beans-4m42
title: Replace panic with graceful error in ID generation
status: todo
type: bug
priority: low
created_at: 2026-03-14T15:07:02Z
updated_at: 2026-03-14T15:07:02Z
parent: beans-5txp
---

bean/id.go panics on nanoid generation failure (line 52). While this should 'never happen with valid alphabet', a panic is a terrible failure mode for a library function. Should use log.Fatalf() or return an error instead.

## Affected Files

- pkg/bean/id.go
