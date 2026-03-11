---
# beans-p5qk
title: Add diff view to status pane
status: completed
type: feature
priority: normal
created_at: 2026-03-11T22:11:34Z
updated_at: 2026-03-11T22:15:55Z
---

When a file is clicked in the ChangesPane file list, show its diff in the bottom half of the pane.

## Summary of Changes

### Backend
- Added `FileDiff(dir, filePath string, staged bool)` function to `internal/gitutil/status.go` — returns unified diff output for a specific file (handles staged, unstaged, and untracked files)
- Added `fileDiff(filePath, staged, path)` GraphQL query to the schema
- Implemented the resolver with the same worktree path validation as `fileChanges`
- Added 3 tests for `FileDiff` (unstaged, staged, untracked)

### Frontend
- Modified `ChangesPane.svelte` to make file rows clickable (using `<button>` elements with `cursor-pointer`)
- Clicking a file fetches its diff via the new `fileDiff` GraphQL query and shows it in the bottom half of the pane using a `SplitPane` with vertical direction
- Clicking the same file again or pressing the X button closes the diff view
- The selected file is highlighted with `bg-surface-alt`
- Diff lines are color-coded: green for additions, red for deletions, accent for hunk headers
- Added `.diff-add`, `.diff-del`, `.diff-hunk` utility classes to `layout.css`
- Selection auto-clears when the file disappears from the changes list (e.g., after a commit)
