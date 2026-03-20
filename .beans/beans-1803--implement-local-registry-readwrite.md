---
# beans-1803
title: Implement local registry read/write
status: todo
type: task
priority: normal
created_at: 2026-03-20T08:33:28Z
updated_at: 2026-03-20T08:42:41Z
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
