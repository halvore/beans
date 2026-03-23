---
# beans-x9ew
title: beans prime prompts agent to ask about init mode
status: completed
type: feature
priority: normal
created_at: 2026-03-23T12:36:42Z
updated_at: 2026-03-23T12:36:48Z
---

When beans prime detects no project, output instructions telling the agent to ask the user whether to use beans init (in-repo) or beans init --local (local storage) instead of silently exiting.

## Summary of Changes

- Modified `internal/commands/prime.go` to output an initialization prompt (instead of silently exiting) when no beans project is found
- The prompt tells agents they MUST ask the user whether to use `beans init` (in-repo) or `beans init --local` (local storage)
- Updated `internal/commands/prime_test.go` to verify the new behavior
