---
# beans-mh8t
title: 'Fix: Web UI not reflecting bean updates from external processes'
status: completed
type: bug
priority: normal
created_at: 2026-03-10T11:34:53Z
updated_at: 2026-03-10T15:44:18Z
order: zzzw
---

When an agent (external CLI process) creates a bean and makes multiple updates, the web UI only shows the first version. Three fixes needed: (1) Fix continue bug in handleChanges that skips Write events after Remove/Rename, (2) Add logging when fanOut drops events, (3) Buffer the GraphQL subscription output channel to reduce backpressure.

## Summary of Changes

### Fix 1: `continue` bug in `handleChanges` (watcher.go)
Moved the `continue` statement inside the `!fileExists` check so that when a file gets both Remove and Write events in the same debounce batch, the Write handler still processes. Previously, any Remove event would skip the Write handler entirely via `continue`.

### Fix 2: Drop logging in `fanOut` (watcher.go)
Added a warning log when events are dropped due to a slow subscriber, providing observability into the silent event loss.

### Fix 3: Buffered GraphQL subscription output channel (schema.resolvers.go)
Changed the `out` channel in `BeanChanged` from unbuffered to buffered (64), reducing backpressure that caused the subscriber channel to fill up and events to be dropped by `fanOut`.

### Tests added (core_test.go)
- `TestHandleChanges_RemoveAndWriteInSameBatch` — regression test for the continue bug
- `TestHandleChanges_RemoveOnly` — ensures Remove-only still deletes correctly
- `TestFanOut_DropsEventsForSlowSubscriber` — verifies buffer behavior
