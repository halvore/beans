---
# beans-taii
title: Prime template hardcodes .beans/skills/ path, breaks local storage
status: completed
type: bug
priority: normal
created_at: 2026-03-23T22:07:08Z
updated_at: 2026-03-23T22:10:21Z
---

When using local storage (beans init --local), skills are stored at ~/.local/beans/projects/<slug>/.beans/skills/ but the prime template hardcodes .beans/skills/ as the path agents should read from. This means agents can't find skill files.

## Summary of Changes

The prime template was hardcoding `.beans/skills/` as the path for skill files. When using local storage (`beans init --local`), skills live at `~/.local/beans/projects/<slug>/.beans/skills/`, so agents couldn't find them.

Fixed by passing the resolved skills directory path into the template data and using it dynamically in the prompt output.
