---
# beans-tt6n
title: Use git remote URL as local project identifier
status: completed
type: task
priority: normal
created_at: 2026-03-24T10:52:28Z
updated_at: 2026-03-24T10:55:01Z
---

When a project is initialized with --local, use the upstream git remote URL as the project identifier instead of the directory name. This prevents collisions when multiple repos have the same name but are hosted on different remotes.

## Approach
- Add RemoteURL field to ProjectEntry in the registry
- Detect origin remote URL during registration via git remote get-url origin
- Derive slug from URL path (owner-repo format), falling back to directory basename if no remote
- Handle both HTTPS and SSH URL formats
- Existing registrations remain compatible (empty URL field)

## Tasks
- [x] Add RemoteURL field to ProjectEntry
- [x] Add git remote URL detection utility
- [x] Add URL-to-slug parsing (handle HTTPS + SSH formats)
- [x] Update makeSlug to prefer remote URL when available
- [x] Update Register to detect and store remote URL
- [x] Add tests for URL parsing and slug generation
- [x] Add tests for registration with remote URL

## Summary of Changes

- Added `RemoteURL` field to `ProjectEntry` in the local registry
- Added `gitutil.RemoteURL()` to detect the origin remote URL
- Added `slugFromRemoteURL()` that parses both HTTPS and SSH git URL formats into owner-repo slugs
- Updated `makeSlug()` to prefer remote URL-derived slugs, falling back to project name/directory basename
- Updated `Register()` to accept and store the remote URL
- Updated `initLocalProject()` to detect and pass the git remote URL during registration
- Existing registrations remain compatible (empty RemoteURL field)
