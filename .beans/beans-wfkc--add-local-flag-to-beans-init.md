---
# beans-wfkc
title: Add --local flag to beans init
status: todo
type: task
priority: normal
created_at: 2026-03-20T08:33:31Z
updated_at: 2026-03-20T08:42:30Z
parent: beans-lhjq
blocked_by:
    - beans-1803
---

Extend `beans init` with a `--local` flag that:
1. Creates the local beans directory at $HOME/.local/beans/projects/<project-name>/
2. Registers the project in the local registry
3. Does NOT create a .beans/ directory or .beans.yml in the project
4. Stores config (prefix, project name, etc.) alongside the local beans dir or in the registry

When `--local` is used, the project directory should remain completely untouched by beans.
