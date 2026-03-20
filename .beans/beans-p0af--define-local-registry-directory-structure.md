---
# beans-p0af
title: Define local registry directory structure
status: completed
type: task
priority: normal
created_at: 2026-03-20T08:33:25Z
updated_at: 2026-03-20T08:45:52Z
parent: beans-lhjq
---

Design and document the directory layout for $HOME/.local/beans. Should include:
- A registry file (e.g. registry.json or registry.yml) mapping project paths to their local beans directories
- Per-project subdirectories containing the .beans files

## Tasks

- [x] Define the directory layout
- [x] Define the registry file format
- [x] Document project directory naming strategy
- [x] Consider edge cases (renamed projects, moved directories, multiple checkouts)

## Design

### Directory Layout

```
$HOME/.local/beans/
  registry.yml                    # Maps project paths → local project dirs
  projects/
    <project-slug>/
      .beans.yml                  # Project config (same format as in-repo config)
      .beans/
        <bean-files>.md           # Bean markdown files
        .beans/.conversations/    # Agent conversation history (gitignored equivalent)
```

### Why $HOME/.local/beans/?

This follows the XDG Base Directory convention ($XDG_DATA_HOME defaults to $HOME/.local/share, but $HOME/.local is the broader user-local prefix). It keeps local bean storage separate from the existing $HOME/.beans/ directory which is used for worktrees.

The path can be overridden via a `BEANS_LOCAL_DIR` environment variable.

### Registry File Format (registry.yml)

```yaml
# $HOME/.local/beans/registry.yml
# Auto-managed by `beans init --local`. Do not edit manually.
projects:
  - path: /Users/alice/projects/myapp
    slug: myapp
    registered_at: 2026-03-20T10:00:00Z
  - path: /Users/alice/projects/api-service
    slug: api-service
    registered_at: 2026-03-15T08:30:00Z
```

Fields per project entry:
- **path** (required): Absolute path to the project root directory. Used as the lookup key.
- **slug** (required): Directory name under `projects/`. Derived from `project.name` in config, or the project directory basename. Must be unique.
- **registered_at** (required): ISO 8601 timestamp of when the project was registered.

### Project Directory Naming (slug)

The `slug` determines the subdirectory name under `projects/`. Strategy:

1. Use the project name from `beans init --local --name <name>`, or fall back to the basename of the project directory.
2. Sanitize: lowercase, replace non-alphanumeric chars with hyphens, trim leading/trailing hyphens.
3. If a slug collision occurs (different project path, same slug), append a short hash suffix: `myapp-a3f1`.

Examples:
- `/Users/alice/projects/MyApp` → slug: `myapp`
- `/Users/alice/work/myapp` (collision) → slug: `myapp-b7e2`

### Per-Project Structure

Each project directory mirrors what would normally exist in the repo:

- `.beans.yml` — Full project config (same schema as in-repo). The `beans.path` field is omitted or set to `.beans` (relative to this directory).
- `.beans/` — Bean files, exactly as they'd appear in the repo's `.beans/` directory.
- `.beans/.conversations/` — Agent conversation data (not version-controlled).

This means `beancore.New()` can be pointed at `$HOME/.local/beans/projects/<slug>/.beans/` with zero changes to the core logic.

### Edge Cases

**Project directory renamed or moved:**
The registry entry becomes stale. `beans` commands run from the new path won't find a match. The user can re-register with `beans init --local` (which detects the orphaned entry by slug and updates the path), or use `beans projects remove <old-path>` to clean up.

**Multiple checkouts of the same repo:**
Each checkout path gets its own registry entry and its own local beans directory. They are independent — no sharing or syncing between checkouts.

**Project directory deleted:**
The registry entry and local beans directory persist. `beans projects list` can flag entries whose project paths no longer exist. `beans projects remove` cleans up both the registry entry and the local directory (with confirmation).

**Slug collision:**
Handled by appending a short hash suffix derived from the project path. See naming strategy above.

## Summary of Changes

Defined the complete local registry directory structure:
- **Root**: `$HOME/.local/beans/` (overridable via `BEANS_LOCAL_DIR` env var)
- **Registry**: `registry.yml` mapping project paths to slugs with timestamps
- **Projects**: `projects/<slug>/` directories each containing `.beans.yml` and `.beans/` — fully compatible with existing `beancore.New()`
- **Naming**: Slugs derived from project name/basename, sanitized, with hash suffix on collision
- **Edge cases**: Covered renamed/moved projects, multiple checkouts, deleted projects, and slug collisions
