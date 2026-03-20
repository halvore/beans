---
# beans-1803
title: Implement local registry read/write
status: completed
type: task
priority: normal
created_at: 2026-03-20T08:33:28Z
updated_at: 2026-03-20T08:55:27Z
parent: beans-lhjq
blocked_by:
    - beans-p0af
---

Add a new pkg (e.g. pkg/localregistry or within pkg/config) that can:
- Read and write the registry file at $HOME/.local/beans/registry.yml
- Register a project (map project path → local beans directory)
- Unregister a project
- Look up a project by its path
- Create the local beans directory for a new project

The registry should store:
- Project absolute path
- Project name
- Local beans directory path
- Date registered

## Summary of Changes

Implemented `pkg/localregistry` package with:
- `Load()` / `Save()` for reading/writing the registry YAML at `$HOME/.local/beans/registry.yml`
- `Register()` to add a project (maps path → slug, creates local beans directory)
- `Unregister()` to remove a project by path
- `Lookup()` to find a project entry by its absolute path
- `ProjectDir()` / `ProjectBeansDir()` for resolving project directories
- Slug generation using `bean.Slugify()` with SHA256-based collision resolution
- `BEANS_LOCAL_DIR` env var support for overriding the default location
- 11 passing tests covering all operations including idempotency, collisions, save/reload
