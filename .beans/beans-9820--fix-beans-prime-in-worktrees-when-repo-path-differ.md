---
# beans-9820
title: Fix beans prime in worktrees when repo path differs from registry
status: completed
type: bug
priority: normal
created_at: 2026-03-24T12:27:09Z
updated_at: 2026-03-24T12:29:11Z
---

When using beans with --local storage and running beans prime in a git worktree whose main repo is cloned at a different path than what's in the local registry, the lookup fails because it only matches by path. The fix is to also match by git remote URL when the path-based lookup fails.

## Summary of Changes

- Added `LookupByRemoteURL` method to the local registry that matches projects by git remote URL
- Updated `loadFromLocalRegistry` in root.go to fall back to remote URL matching when a worktree's main repo path doesn't match the registry
- This fixes the case where a repo is cloned at a different path (e.g. Conductor workspaces) than where it was originally registered with `beans init --local`
- Added unit tests for both the new registry method and the worktree remote URL fallback path
