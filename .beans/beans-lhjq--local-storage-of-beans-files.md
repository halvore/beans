---
# beans-lhjq
title: Local storage of .beans files
status: todo
type: epic
priority: normal
tags:
    - idea
created_at: 2026-03-20T08:33:18Z
updated_at: 2026-03-20T08:42:23Z
---

Support storing beans files outside the project directory, in a local registry at $HOME/.local/beans. When a project is initialized with --local, all bean files are stored under the local directory, keeping the project directory clean. If no .beans.yml exists in the project, the local registry is checked as a fallback.
