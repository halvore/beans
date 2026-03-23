---
# beans-nneu
title: Agent dev skills and review status
status: review
type: feature
priority: normal
created_at: 2026-03-23T13:05:50Z
updated_at: 2026-03-23T14:22:23Z
---

Add a skills system for beans development workflows. Skills are .md files in .beans/skills/ that agents can invoke. Also add a 'review' status. Initial skills: /plan, /review, /ship, /investigate. beans prime outputs a catalog of available skills.

## Summary of Changes

### Skills System
- Created `.beans/skills/` directory with 4 skill files:
  - **plan.md** — Critical bean planning with pushback and discussion before creating beans
  - **review.md** — Pre-PR code review (correctness, design, tests, security)
  - **ship.md** — Prepare and open a pull request (test, update beans, commit, create PR)
  - **investigate.md** — Systematic hypothesis-driven debugging
- Updated `beans prime` to discover skill files from `.beans/skills/` and include a skills catalog in the agent prompt
- Skills are auto-discovered from `.md` files; description extracted from first heading

### Review Status
- Added `review` status to DefaultStatuses in `pkg/config/config.go`
- Updated all frontend references: STATUS_ORDER, statusColors, BeanForm statuses, BoardView columns
- Added CSS theme variables for review status (violet color scheme, light + dark modes)
- Updated GraphQL schema comment and regenerated frontend types

### Tests
- Added `TestDiscoverSkills` and `TestExtractSkillDescription` tests
- Updated all existing status-related tests for the new 6-status set

### Skills Installation
- Default skills are now embedded in the binary via `//go:embed`
- `beans init` automatically installs default skills during project initialization
- `beans init --local` also installs skills in the local storage directory
- New `beans skills init` command installs/updates default skills for existing projects
- `beans skills init --force` overwrites existing skill files
- Added `TestInstallDefaultSkills` tests (install, no-overwrite, force-overwrite)
- Fixed PersistentPreRunE skip logic to not skip `skills init` subcommand
