# Project Agent Guide

<!-- Single source of truth for every AI coding agent. CLAUDE.md is a symlink to this file. -->

## What this project is

Custos is an open-source, AI-powered log intelligence platform. Language SDKs (Python, Node.js, Java) capture errors from applications and forward them to a Go + Chi HTTP server. Events flow through Apache Kafka, an AI analysis worker calls a pluggable provider (Claude, OpenAI, Gemini, or Ollama), results are stored in PostgreSQL, and a TanStack dashboard gives teams live visibility.

Stack: Go 1.23 · Chi · Apache Kafka (`segmentio/kafka-go`) · PostgreSQL · TanStack Start · TypeScript.

## Workflow Authority

- Canonical workflow policy: `.agents/workflows/workflow-contract/spec/*`
- Canonical validator: `python3 .agents/workflows/workflow-contract/scripts/validate_workflow.py`

## Start Here

1. Classify the task.
2. Run `python3 .agents/workflows/workflow-contract/scripts/validate_workflow.py`.
3. Read `docs/design/` for approved truth.
4. Read `docs/implementation/` for the current plan.
5. If behavior is unresolved, read or create `docs/changes/proposed/`.
6. Inspect target code before editing.

## Documentation Workflow

Use the workflow contract for:
- Design docs, implementation docs, backlog and task updates
- Proposed changes and docs-vs-code reconciliation

Layer rules:
- `docs/design/` — approved product/system truth only.
- `docs/implementation/` — execution plans, phases, tasks, and status only.
- `docs/changes/proposed/` — unresolved proposals only.

Do not define net-new behavior in implementation docs. Put unresolved behavior in `docs/changes/proposed/` until accepted.

## Conventions

- **Commit format**: `type(scope): message` — e.g. `feat(ingestion): add fingerprint dedup`
- **Branch naming**: `feat/`, `fix/`, `chore/` prefix
- **Tests**: `go test ./...` from repo root
- **Lint**: `golangci-lint run ./...`
- **Hot reload (dev)**: `make dev`
- **Build**: `make build` → binary at `bin/server`
- **Migrations**: `make migrate-up` / `make migrate-down`

## Package Dependency Rule

```
domain ← store ← service ← api
domain ← queue ← service
domain ← provider ← service
domain ← notification ← service
```

`internal/domain` has zero imports from other internal packages. Never import a sibling package.

## Key Design Docs

- Architecture: `docs/design/architecture/system.md`
- Project structure: `docs/design/architecture/project-structure.md`
- AI provider adapter: `docs/design/integrations/ai-provider.md`
- SDK design: `docs/design/features/sdk-design.md`
