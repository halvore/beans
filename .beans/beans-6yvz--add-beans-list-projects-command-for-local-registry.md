---
# beans-6yvz
title: Add beans list-projects command for local registry
status: completed
type: task
priority: normal
created_at: 2026-03-20T08:33:42Z
updated_at: 2026-03-20T09:29:36Z
parent: beans-lhjq
blocked_by:
    - beans-1803
---

Add a CLI command to manage the local registry:
- `beans projects list` — list all locally registered projects
- `beans projects remove <path>` — unregister a project (optionally delete its beans)

This gives users visibility into what's stored in their local beans directory.

## Summary of Changes

Implemented `beans projects` CLI command group with two subcommands:

- `beans projects list` — Lists all locally registered projects (slug, path, registration date). Supports `--json` for machine-readable output.
- `beans projects remove <path>` — Unregisters a project from the local registry. Supports `--delete-data` to also remove the project's local beans data, and `--json` for structured output.

The commands bypass core initialization (no .beans directory needed) since they operate directly on the local registry.

### Files
- `internal/commands/projects.go` — New command implementation
- `internal/commands/projects_test.go` — 5 tests covering list (empty, populated) and remove (success, with data deletion, not found)
- `internal/commands/register.go` — Registered the new command
- `internal/commands/root.go` — Skip core init for projects subcommands
