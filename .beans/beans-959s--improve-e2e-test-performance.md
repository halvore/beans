---
# beans-959s
title: Improve E2E test performance
status: completed
type: task
priority: low
created_at: 2026-03-14T15:07:02Z
updated_at: 2026-03-14T16:41:58Z
parent: beans-5txp
---

Each E2E test spins up a full server (mkdir + git init + beans init + spawn server + wait for port). As the test suite grows, this will become painful.

## Proposed Improvements

- Share server fixtures for read-only tests
- Parallelize test workers where possible
- Consider lighter-weight integration tests for some scenarios
- Evaluate whether some E2E tests could be replaced with component tests

## Affected Files

- frontend/e2e/fixtures.ts
- frontend/e2e/*.spec.ts

## Summary of Changes

Optimized E2E test performance with three improvements:
- Worker-scoped `beansTemplate` fixture: runs `beans init` once per worker and copies `.beans/` dir per test via `cpSync`, avoiding expensive CLI invocation for every test
- Faster `waitForServer`: TCP connect probe before HTTP fetch, reduced polling intervals (25ms/50ms vs 100ms)
- Shared `GIT_ENV` constant: deduplicated git environment config
- Enabled `fullyParallel: true` in Playwright config
