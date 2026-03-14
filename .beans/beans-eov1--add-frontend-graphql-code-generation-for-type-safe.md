---
# beans-eov1
title: Add frontend GraphQL code generation for type safety
status: completed
type: feature
priority: low
created_at: 2026-03-14T15:07:02Z
updated_at: 2026-03-14T17:06:44Z
parent: beans-5txp
---

The backend generates types from the GraphQL schema via mise codegen (gqlgen), but the frontend doesn't validate queries against the schema at build time. A schema mismatch between frontend queries and backend resolvers would only be caught at runtime.

## Proposed Fix

Add graphql-codegen (or similar) to the frontend build pipeline to generate TypeScript types from the schema and validate all query/mutation documents.

## Affected Files

- frontend/package.json (new dev dependency)
- frontend/codegen.ts (new config)
- frontend/src/lib/graphqlClient.ts (use generated types)


## Tasks

- [x] Install graphql-codegen dependencies
- [x] Create codegen.ts config
- [x] Create operations.graphql with all operations (using proper GraphQL fragments)
- [x] Run codegen to generate typed document nodes
- [x] Update stores to use generated types and document nodes
- [x] Update components to use generated types and document nodes
- [x] Wire up pnpm codegen script and mise codegen task
- [x] Verify build passes with no warnings


## Summary of Changes

Added frontend GraphQL code generation using `@graphql-codegen/cli` with `typescript`, `typescript-operations`, and `typed-document-node` plugins:

- **`frontend/codegen.ts`** — codegen config pointing at the backend schema (`internal/graph/schema.graphqls`)
- **`frontend/src/lib/graphql/operations.graphql`** — all GraphQL operations (queries, mutations, subscriptions) extracted from scattered inline `gql` strings into a single file with proper GraphQL fragment syntax
- **`frontend/src/lib/graphql/generated.ts`** — auto-generated TypeScript types + typed `DocumentNode` objects that provide full type inference with urql's client methods
- **All stores and components** updated to import generated types and document nodes instead of hand-written interfaces and inline `gql` strings
- **`mise codegen`** now runs both backend (gqlgen) and frontend (graphql-codegen) in sequence
- **`pnpm codegen`** script added to `frontend/package.json`

Key design decisions:
- Used urql's own `TypedDocumentNode` type (via `import type from 'urql'`) instead of `@graphql-typed-document-node/core` which has `"main": ""` breaking Vite's module resolution
- Generated file is committed (matching the backend's `generated.go` convention) so builds work without running codegen
- Hand-written interfaces replaced with re-exports from generated types (e.g., `export type Bean = BeanFieldsFragment`)
