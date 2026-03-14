---
# beans-vffa
title: Improve error handling and add structured logging
status: todo
type: task
created_at: 2026-03-14T15:06:45Z
updated_at: 2026-03-14T15:06:45Z
parent: beans-5txp
---

Several places silently swallow errors or use inconsistent logging. This makes debugging production issues harder.

## Silent Failures to Fix

- Worktree loadMeta() silently returns nil on JSON unmarshal failure
- DragOrder mutations log errors but show nothing to the user
- Agent session load failures return nil without logging

## Proposed Changes

1. Adopt Go's slog package for structured logging across the backend
2. Add user-facing error toasts on the frontend for mutation failures
3. At minimum, log warnings where errors are currently swallowed

## Affected Files (Backend)

- internal/worktree/worktree.go (loadMeta)
- internal/agent/manager.go (GetSession)
- Throughout: replace log.Printf with slog

## Affected Files (Frontend)

- frontend/src/lib/stores/dragOrder.ts (add error toast)
- frontend/src/lib/components/ (add toast component)
