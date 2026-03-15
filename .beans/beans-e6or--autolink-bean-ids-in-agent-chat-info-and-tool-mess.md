---
# beans-e6or
title: Autolink bean IDs in agent chat INFO and TOOL messages
status: completed
type: bug
priority: normal
created_at: 2026-03-15T17:41:18Z
updated_at: 2026-03-15T17:41:24Z
---

Bean mentions in agent chat TOOL and INFO messages are not autolinked. Only ASSISTANT and USER messages go through renderMarkdown() which has the beanLinkExtension. TOOL and INFO messages use the decryptText action (plain textContent), so bean IDs appear as unlinked text.

## Summary of Changes

- Added `linkifyBeanIds()` utility to `markdown.ts` that HTML-escapes text and wraps bean ID patterns in clickable bean-link tags
- Extended `decryptText` action with optional `html` parameter — uses `innerHTML` instead of `textContent` once animation completes
- Updated `AgentMessages.svelte` to pass linkified HTML for INFO and TOOL messages
- Added e2e test verifying bean ID autolinking across all message types
- Extended `AgentSessionBuilder` to support `tool` and `info` message roles
