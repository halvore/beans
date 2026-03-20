---
# beans-l1eu
title: Update config/path resolution to check local registry
status: completed
type: task
priority: normal
created_at: 2026-03-20T08:33:36Z
updated_at: 2026-03-20T09:15:26Z
parent: beans-lhjq
blocked_by:
    - beans-1803
---

Modify the config and beans path resolution logic so that when no .beans.yml is found walking up from cwd:
1. Check $HOME/.local/beans/registry.yml for the current project path
2. If found, use the local beans directory as the beans path
3. Load config from the local directory

This affects:
- `config.LoadFromDirectory()` / `config.FindConfig()`
- `resolveBeansPath()` in internal/commands/root.go
- Any other path resolution that currently assumes .beans.yml is in the project tree

The fallback order becomes:
1. --beans-path flag / BEANS_PATH env
2. .beans.yml in project tree (walk up)
3. Local registry lookup for cwd
4. Error: not initialized


## Summary of Changes

Modified `internal/commands/root.go` to add local registry fallback when no `.beans.yml` is found in the project tree:

- Split the config loading in `PersistentPreRunE` to first try `FindConfig()`, then fall back to `loadFromLocalRegistry()`
- New `loadFromLocalRegistry()` function checks the local registry for the cwd, loads config from the local project directory if found, or returns a default config
- Keeps `pkg/config` dependency-free — the local registry import stays in the commands layer
- Added 4 tests covering all fallback scenarios
