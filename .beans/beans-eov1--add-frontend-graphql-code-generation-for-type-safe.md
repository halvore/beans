---
# beans-eov1
title: Add frontend GraphQL code generation for type safety
status: todo
type: feature
priority: low
created_at: 2026-03-14T15:07:02Z
updated_at: 2026-03-14T15:07:02Z
parent: beans-5txp
---

The backend generates types from the GraphQL schema via mise codegen (gqlgen), but the frontend doesn't validate queries against the schema at build time. A schema mismatch between frontend queries and backend resolvers would only be caught at runtime.

## Proposed Fix

Add graphql-codegen (or similar) to the frontend build pipeline to generate TypeScript types from the schema and validate all query/mutation documents.

## Affected Files

- frontend/package.json (new dev dependency)
- frontend/codegen.ts (new config)
- frontend/src/lib/graphqlClient.ts (use generated types)
