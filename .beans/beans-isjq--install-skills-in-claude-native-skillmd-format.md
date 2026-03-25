---
# beans-isjq
title: Install skills in Claude-native SKILL.md format
status: review
type: bug
priority: normal
created_at: 2026-03-25T11:34:59Z
updated_at: 2026-03-25T11:38:08Z
---

`beans init skills` installs skills as flat `.md` files in `~/.claude/skills/beans/` (e.g. `bplan.md`), but Claude Code's native skill discovery expects each skill in its own subdirectory as `SKILL.md` with YAML frontmatter (e.g. `~/.claude/skills/beans-bplan/SKILL.md`).

This means the beans skills only work through the `beans prime` hook, not as native Claude skills that appear in the skills list and can be invoked with `/`.

## Fix
- For Claude: install each skill as `~/.claude/skills/beans-<name>/SKILL.md` with auto-generated YAML frontmatter
- For Codex: keep the current flat format
- Update `discoverSkills` to handle the new directory structure
- Cleanup old flat files during install

## Tasks
- [x] Update `installSkills` to support native format (subdirectory + SKILL.md + frontmatter)
- [x] Update `skillsDir` for Claude to return `~/.claude/skills`
- [x] Update `discoverSkills` to find `beans-*/SKILL.md`
- [x] Update tests
- [x] Run tests

## Summary of Changes

- Changed skill installation for Claude to use native SKILL.md format: `~/.claude/skills/beans-<name>/SKILL.md` with YAML frontmatter
- Codex keeps the flat format: `~/.codex/skills/beans/<name>.md`
- Updated `discoverSkills` to find both native (`beans-*/SKILL.md`) and flat (`*.md`) formats
- Updated `extractSkillDescription` to handle YAML frontmatter (skips it, falls back to `description:` field)
- Added `skillFormat` type and `Format` field to `agentTool` struct
- Updated all tests for new format
