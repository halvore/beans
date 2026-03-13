---
# beans-r7oc
title: Improve subscription backpressure handling
status: todo
type: task
priority: low
created_at: 2026-03-13T17:49:10Z
updated_at: 2026-03-13T17:49:10Z
parent: beans-oyic
---

Non-blocking fanOut silently drops events when a subscriber's channel is full. For a UI that depends on subscription accuracy, this could cause state drift. Consider larger buffers, a catch-up mechanism, or full-state resync on overflow.
