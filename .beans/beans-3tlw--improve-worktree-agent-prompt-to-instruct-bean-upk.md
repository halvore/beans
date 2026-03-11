---
# beans-3tlw
title: Improve worktree agent prompt to instruct bean upkeep
status: completed
type: task
created_at: 2026-03-11T21:26:46Z
updated_at: 2026-03-11T21:26:46Z
---

The worktree-specific agent prompt was minimal — just bean metadata and body. It didn't instruct the agent to keep the bean updated during work.

Added instructions to the injected prompt telling the agent to:
- Check off completed todo items
- Update the description if implementation diverges from the plan
- Add a Summary of Changes section and mark the bean completed when done

## Summary of Changes

Updated the worktree agent prompt builder in `internal/commands/serve.go` to append bean upkeep instructions after the bean description.
