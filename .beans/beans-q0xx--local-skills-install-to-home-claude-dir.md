---
# beans-q0xx
title: Local skills install to home .claude dir
status: completed
type: task
priority: normal
created_at: 2026-03-24T08:30:12Z
updated_at: 2026-03-24T08:32:07Z
---

When beans is set up with --local, skills should be stored in $HOME/.claude/skills/ instead of the project's .claude/commands/ directory

## Summary of Changes

- Refactored `installClaudeCodeCommands` to accept a target directory instead of computing it from projectDir
- Added `claudeCommandsDir` helper that returns `$HOME/.claude/skills/` for local projects and `<projectDir>/.claude/commands/` for in-repo projects
- Updated `beans init --local` to install stubs to `$HOME/.claude/skills/`
- Updated `beans skills init` to detect local projects and use the correct target directory
- Added tests for `claudeCommandsDir`
