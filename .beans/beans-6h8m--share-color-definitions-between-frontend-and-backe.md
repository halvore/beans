---
# beans-6h8m
title: Share color definitions between frontend and backend
status: todo
type: task
priority: low
created_at: 2026-03-14T15:07:02Z
updated_at: 2026-03-14T15:07:02Z
parent: beans-5txp
---

Status/type colors are defined separately in frontend CSS (layout.css) and backend config (internal/config/config.go). These must be manually kept in sync, which is error-prone.

## Proposed Fix

Extract color definitions to a shared JSON file (or generate CSS from the Go config, or vice versa) so there's a single source of truth.

## Affected Files

- frontend/src/app.css or layout.css
- internal/config/config.go
