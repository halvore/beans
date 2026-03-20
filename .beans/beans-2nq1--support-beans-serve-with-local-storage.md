---
# beans-2nq1
title: Support beans serve with local storage
status: completed
type: task
priority: normal
created_at: 2026-03-20T08:33:39Z
updated_at: 2026-03-20T09:23:33Z
parent: beans-lhjq
blocked_by:
    - beans-l1eu
    - beans-wfkc
---

Ensure `beans serve` works correctly when beans are stored locally:
- File watching should monitor the local beans directory instead of .beans/ in the project
- The web UI should function identically regardless of storage location
- Worktree integration needs consideration: worktrees may need to copy or symlink local beans, or the serve process needs to resolve local paths for worktree agents


## Summary of Changes

Added `ProjectRoot` field to `Config` to distinguish between the config storage directory and the actual project directory (git repo root). This matters when beans are stored locally via the local registry, where `ConfigDir()` points to `~/.local/beans/projects/<slug>/` but the project actually lives elsewhere.

Changes:
- `pkg/config/config.go`: Added `projectRoot` field, `ProjectRoot()` getter, and `SetProjectRoot()` setter. Falls back to `ConfigDir()` when not set.
- `internal/commands/root.go`: `loadFromLocalRegistry()` now calls `cfg.SetProjectRoot(entry.Path)` so the actual project path is preserved.
- `internal/commands/serve.go`: Replaced all `filepath.Dir(core.Root())` and `cfg.ConfigDir()` references for project root with `cfg.ProjectRoot()`. This fixes: worktree manager repo root, forge detection, agent system prompts, terminal working directory, and project name derivation.
- Added tests for `ProjectRoot` in both `pkg/config/` and `internal/commands/`.
