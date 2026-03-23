---
# beans-taii
title: Prime template hardcodes .beans/skills/ path, breaks local storage
status: in-progress
type: bug
created_at: 2026-03-23T22:07:08Z
updated_at: 2026-03-23T22:07:08Z
---

When using local storage (beans init --local), skills are stored at ~/.local/beans/projects/<slug>/.beans/skills/ but the prime template hardcodes .beans/skills/ as the path agents should read from. This means agents can't find skill files.
