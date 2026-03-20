---
# beans-6yvz
title: Add beans list-projects command for local registry
status: draft
type: task
priority: normal
created_at: 2026-03-20T08:33:42Z
updated_at: 2026-03-20T08:33:47Z
parent: beans-lhjq
blocked_by:
    - beans-1803
---

Add a CLI command to manage the local registry:
- `beans projects list` — list all locally registered projects
- `beans projects remove <path>` — unregister a project (optionally delete its beans)

This gives users visibility into what's stored in their local beans directory.
