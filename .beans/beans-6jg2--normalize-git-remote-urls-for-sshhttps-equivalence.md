---
# beans-6jg2
title: Normalize git remote URLs for SSH/HTTPS equivalence in registry lookup
status: completed
type: bug
priority: normal
created_at: 2026-03-24T13:50:52Z
updated_at: 2026-03-24T13:51:42Z
---

LookupByRemoteURL does exact string comparison, so SSH (git@github.com:owner/repo.git) and HTTPS (https://github.com/owner/repo.git) URLs pointing to the same repo don't match. This causes 'beans prime' to fail in worktrees cloned via a different protocol than the registered project.

## Summary of Changes

Added `normalizeRemoteURL` function that extracts a canonical `host/owner/repo` form from both SSH and HTTPS git remote URLs. Updated `LookupByRemoteURL` to normalize both sides before comparing, so `git@github.com:o/r.git` and `https://github.com/o/r.git` are treated as equivalent.
