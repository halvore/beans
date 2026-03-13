---
# beans-rren
title: Avoid disk I/O while holding write lock in Core.Update()
status: todo
type: task
priority: normal
created_at: 2026-03-13T17:49:05Z
updated_at: 2026-03-13T17:49:05Z
parent: beans-oyic
---

In core.go Update(), the ETag calculation reads from disk while holding the write lock. If the filesystem is slow, this blocks all other goroutines. Consider reading the file before acquiring the lock, or caching the on-disk ETag.
