---
# beans-v0o2
title: beans prime silently exits with local storage
status: completed
type: bug
priority: normal
created_at: 2026-03-20T11:15:30Z
updated_at: 2026-03-20T11:19:59Z
---

## Problem

`beans prime` silently produces no output when the project uses local storage (i.e. `beans init --local`).

## Root Cause

In `internal/commands/prime.go:31-41`, the command has an early-exit check that calls `config.FindConfig(cwd)` to see if a beans project exists. This only looks for a `.beans.yml` file by walking up the directory tree. With local storage, the config lives in `~/.beans/projects/<hash>/` instead, so `FindConfig` returns empty and `prime` silently returns nil.

The main CLI's `PersistentPreRunE` in `root.go` correctly handles the local registry fallback (via `loadFromLocalRegistry`), but `prime`'s own check runs before that and short-circuits.

## Secondary Issue

Even if the early exit is fixed, `prime` uses `config.DefaultTypes`, `config.DefaultStatuses`, and `config.DefaultPriorities` (line 51-53) instead of reading from the project's actual loaded config. This means custom types/statuses/priorities are not reflected in the agent prompt.

## Expected Behavior

`beans prime` should:
1. Detect local-storage projects via the local registry (same as other commands)
2. Use the project's actual config for types, statuses, and priorities

## Steps to Reproduce

1. `beans init --local` in a directory
2. `beans prime` → produces no output
3. Compare with `beans list` → works correctly

## Todo

- [x] Fix the early-exit check in `prime.go` to also check the local registry
- [x] Use the loaded project config instead of defaults for types/statuses/priorities (N/A: these are hardcoded, not per-project)
- [x] Add tests for `beans prime` with local storage


## Summary of Changes

Fixed `beans prime` to work with local storage by replacing the early-exit check (which only called `config.FindConfig()`) with full config loading logic that also checks the local registry via `loadFromLocalRegistry()`. Added validation that the `.beans` directory exists when using local registry to preserve silent-exit behavior for non-project directories. Added three tests covering local storage, in-repo, and no-project scenarios.
