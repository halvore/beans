---
# beans-fydu
title: Replace findMainBeanFile directory scan with in-memory lookup
status: todo
type: task
priority: normal
created_at: 2026-03-13T17:49:07Z
updated_at: 2026-03-13T17:49:07Z
parent: beans-oyic
---

findMainBeanFile() does os.ReadDir() and linearly scans filenames on every worktree change. Core already has a beans map indexed by ID — use it instead of hitting the filesystem.
