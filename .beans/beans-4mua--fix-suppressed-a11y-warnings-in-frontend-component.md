---
# beans-4mua
title: Fix suppressed a11y warnings in frontend components
status: todo
type: bug
priority: normal
created_at: 2026-03-13T17:49:14Z
updated_at: 2026-03-13T17:49:14Z
parent: beans-oyic
---

RenderedMarkdown.svelte, AgentMessages.svelte, and PlanningView.svelte use svelte-ignore to suppress accessibility warnings. Project rules say never suppress a11y warnings. Fix with proper keyboard handlers or semantic elements (e.g. use <a> tags in rendered markdown, use <button> instead of div with role=button).
