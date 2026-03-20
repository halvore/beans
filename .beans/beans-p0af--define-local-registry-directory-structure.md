---
# beans-p0af
title: Define local registry directory structure
status: todo
type: task
priority: normal
created_at: 2026-03-20T08:33:25Z
updated_at: 2026-03-20T08:42:39Z
parent: beans-lhjq
---

Design and document the directory layout for $HOME/.local/beans. Should include:
- A registry file (e.g. registry.json or registry.yml) mapping project paths to their local beans directories
- Per-project subdirectories containing the .beans files

Example layout:
```
$HOME/.local/beans/
  registry.yml          # Maps project paths → local project dirs
  projects/
    <project-hash-or-name>/
      .beans/
        <bean-files>.md
```
