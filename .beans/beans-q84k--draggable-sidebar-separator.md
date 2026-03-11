---
# beans-q84k
title: Draggable sidebar separator
status: todo
type: feature
created_at: 2026-03-11T20:57:39Z
updated_at: 2026-03-11T20:57:39Z
---

Make the sidebar width resizable by dragging the border between the sidebar and main content area.

## Current State
- Sidebar (`Sidebar.svelte`) has a fixed width of `w-56` (14rem)
- The separator is just a `border-r border-border` on the sidebar's nav element
- Layout is in `+layout.svelte`: sidebar and main content sit in a flex row

## Requirements
- [ ] Add a draggable separator/handle between the sidebar and main content
- [ ] Allow the user to resize the sidebar width by dragging the separator
- [ ] Persist the sidebar width (e.g. in localStorage) so it survives page reloads
- [ ] Set reasonable min/max width constraints (e.g. 150px–400px)
- [ ] Show a visual indicator (e.g. cursor change, highlight) on hover/drag
- [ ] Use existing Svelte components/patterns from the project where possible

## Implementation Notes
- The sidebar width is currently set via Tailwind class `w-56` on the `<nav>` in `Sidebar.svelte` — this will need to become a dynamic inline style
- Width persistence should use the same pattern as other UI state (see `uiState.svelte` and the `+layout.ts` load function for localStorage access)
- Consider a small dedicated component (e.g. `ResizeHandle.svelte`) for the drag handle
