---
# beans-2nq1
title: Support beans serve with local storage
status: todo
type: task
priority: normal
created_at: 2026-03-20T08:33:39Z
updated_at: 2026-03-20T08:42:43Z
parent: beans-lhjq
blocked_by:
    - beans-l1eu
    - beans-wfkc
---

Ensure `beans serve` works correctly when beans are stored locally:
- File watching should monitor the local beans directory instead of .beans/ in the project
- The web UI should function identically regardless of storage location
- Worktree integration needs consideration: worktrees may need to copy or symlink local beans, or the serve process needs to resolve local paths for worktree agents
