---
# beans-iji0
title: Client-side text filter for backlog and board views
status: completed
type: feature
priority: normal
created_at: 2026-03-10T09:03:09Z
updated_at: 2026-03-10T15:44:18Z
order: zzzzs
---

Add a text input at the top of both the backlog and board panes that filters the visible beans client-side as the user types.

## Design

- Single shared filter string in `uiState.svelte.ts` (persisted to localStorage so it survives view toggles)
- Filter input component rendered above the bean list in both backlog and board views
- Matching: case-insensitive substring match against bean title, type, status, tags, and ID
- In backlog view: a top-level bean is shown if it OR any of its children match; matching children are shown, non-matching children are hidden
- In board view: filter each column's beans individually
- Clear button (×) inside the input to quickly reset the filter
- Keyboard shortcut: Cmd/Ctrl+F to focus the filter input (with preventDefault to avoid browser find)

## Tasks

- [x] Add `filterText` state to `uiState.svelte.ts`
- [x] Create a `FilterInput.svelte` component (text input with clear button)
- [x] Implement `matchesFilter(bean, text)` utility function
- [x] Integrate filter into backlog view in `+page.svelte` (filter `topLevelBeans` with recursive child matching)
- [x] Integrate filter into `BoardView.svelte` (filter `beansForStatus`)
- [x] Add Cmd/Ctrl+F keyboard shortcut to focus filter
- [x] Write e2e tests for filtering in both views

## Summary of Changes

Implemented client-side text filtering for backlog and board views:

- Added `filterText` reactive state to `UIState` with localStorage persistence
- Created `FilterInput.svelte` component with clear button (×)
- Created `matchesFilter()` utility: case-insensitive substring match against title, type, status, tags, and ID
- Backlog view: parent beans shown when they or any child matches; non-matching children hidden
- Board view: each column filtered individually
- Cmd/Ctrl+F keyboard shortcut focuses the filter input
- Filter state shared between backlog and board views
- 8 e2e tests covering all scenarios
