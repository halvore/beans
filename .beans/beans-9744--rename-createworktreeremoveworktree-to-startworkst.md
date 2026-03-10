---
# beans-9744
title: Rename createWorktree/removeWorktree to startWork/stopWork
status: completed
type: feature
priority: normal
created_at: 2026-03-10T17:10:03Z
updated_at: 2026-03-10T17:17:12Z
---

Refactor the worktree-related GraphQL mutations:

- Rename `createWorktree` to `startWork` and `removeWorktree` to `stopWork`
- `startWork` only takes a `beanId` argument
- `startWork` returns an error if the bean already has a worktree
- Branch and path are derived automatically from the bean ID (not returned/configurable)
- `stopWork` mutation to remove the worktree

## Tasks

- [x] Rename `createWorktree` mutation to `startWork` in GraphQL schema
- [x] Rename `removeWorktree` mutation to `stopWork` in GraphQL schema
- [x] Update resolvers for the renamed mutations
- [x] Derive branch name and path automatically from bean ID
- [x] Add error when bean already has a worktree in `startWork`
- [x] Update CLI commands that use these mutations
- [x] Update frontend code that calls these mutations
- [x] Update tests
- [x] Run `mise codegen` to regenerate GraphQL code

## Summary of Changes

- Renamed `createWorktree` → `startWork` and `removeWorktree` → `stopWork` in GraphQL schema
- Updated resolver method names (`CreateWorktree` → `StartWork`, `RemoveWorktree` → `StopWork`)
- Ran `mise codegen` to regenerate the generated GraphQL code
- Updated frontend `WorktreeStore` to use `startWork`/`stopWork` mutations and method names
- Updated `BeanDetail.svelte` to call the renamed store methods
- No CLI commands needed updating (none directly used these mutations)
- Branch/path were already derived automatically from bean ID
- Worktree creation already errored on duplicates via the worktree manager
- All tests pass (21 e2e + all Go unit tests)
