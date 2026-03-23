---
# beans-xj84
title: Local registry lookup fails in git worktrees
status: completed
type: bug
created_at: 2026-03-23T12:03:31Z
updated_at: 2026-03-23T12:03:31Z
---

When beans uses local storage (beans init --local), the CLI resolves the beans path via the local registry by matching cwd against registered project paths. In a git worktree, cwd differs from the registered path, so the lookup fails and beans are not found.

Fix: when direct lookup fails, detect if we're in a git worktree via gitutil.MainWorktreeRoot and retry the lookup with the main repo path.

## Summary of Changes

- Modified loadFromLocalRegistry in internal/commands/root.go to fall back to MainWorktreeRoot when direct registry lookup fails
- Added test case verifying worktree-to-main-repo resolution works
