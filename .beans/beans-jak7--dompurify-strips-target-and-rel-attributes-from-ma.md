---
# beans-jak7
title: DOMPurify strips target and rel attributes from markdown links
status: completed
type: bug
priority: normal
created_at: 2026-03-11T18:06:44Z
updated_at: 2026-03-11T18:06:49Z
---

DOMPurify sanitization was stripping `target="_blank"` and `rel="noopener noreferrer"` from rendered markdown links because these attributes were not in the `ADD_ATTR` allowlist.

## Root Cause

In `frontend/src/lib/markdown.ts`, the custom link renderer correctly sets `target="_blank" rel="noopener noreferrer"` on all links, but DOMPurify.sanitize() only had `data-bean-id` in its `ADD_ATTR` list, so `target` and `rel` were stripped.

## Fix

Added `target` and `rel` to the DOMPurify `ADD_ATTR` allowlist.

## Tasks

- [x] Add `target` and `rel` to DOMPurify ADD_ATTR in markdown.ts
- [x] Add e2e test verifying external links have correct attributes

## Summary of Changes

- **`frontend/src/lib/markdown.ts`**: Added `target` and `rel` to DOMPurify `ADD_ATTR` allowlist so external links correctly open in new tabs.
- **`frontend/e2e/markdown-links.spec.ts`**: New e2e test that creates a bean with markdown and autolinked URLs, then verifies both have `target="_blank"` and `rel="noopener noreferrer"`.
