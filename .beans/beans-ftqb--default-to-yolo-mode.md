---
# beans-ftqb
title: Default to yolo mode
status: completed
type: feature
priority: normal
created_at: 2026-03-10T12:03:01Z
updated_at: 2026-03-10T15:44:18Z
order: zzzzw
---

Currently, agent chat sessions start in **act mode** (`--permission-mode acceptEdits`) by default, requiring explicit user opt-in to switch to yolo mode (`--dangerously-skip-permissions`). This feature would make yolo mode the default, so new agent sessions run fully autonomously without requiring the user to toggle it on each time.

## Context

- Agent sessions are spawned in `internal/agent/claude.go` with permission flags based on `Session.YoloMode` / `Session.PlanMode`
- The current default is act mode (`acceptEdits`), with yolo mode toggled via `SetYoloMode` in the manager
- The UI toggle lives in `AgentChat.svelte`

## Tasks

- [ ] Change the default `YoloMode` on new sessions to `true`
- [ ] Update the UI toggle so it reflects the new default
- [ ] Consider making this configurable (e.g. a project-level setting)
- [ ] Update tests in `manager_test.go` to reflect the new default

## Summary of Changes

New agent sessions now default to yolo mode (--dangerously-skip-permissions) instead of act mode. Set YoloMode=true in both session creation paths (loadOrCreateSession and GetSession). Updated tests to reflect the new default. Filed beans-xfup as follow-up for a project-level setting to make this configurable.
