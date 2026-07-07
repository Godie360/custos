# Contributing to Custos

Thank you for your interest in contributing. This guide covers everything you need to get a change merged.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Before You Start](#before-you-start)
- [Branch Strategy](#branch-strategy)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Commit Messages](#commit-messages)
- [Pull Requests](#pull-requests)
- [Code Standards](#code-standards)
- [Testing](#testing)
- [Project Structure](#project-structure)

---

## Code of Conduct

Be respectful and constructive. We welcome contributors of all backgrounds and experience levels.

---

## Before You Start

- For **bug fixes** — open an issue first so we can confirm the bug and discuss the fix.
- For **new features** — open an issue or a discussion before writing code. Features go through the design doc workflow in `docs/`.
- For **small improvements** (typos, docs, minor refactors) — go ahead and open a PR directly.

---

## Branch Strategy

| Branch | Purpose |
|---|---|
| `main` | Stable, released code. Never commit directly here. |
| `develop` | Integration branch. All PRs target here. |
| `feat/*` | New features |
| `fix/*` | Bug fixes |
| `chore/*` | Tooling, dependencies, config |

Always branch off `develop`:

```bash
git checkout develop
git pull origin develop
git checkout -b feat/your-feature-name
```

---

## Development Setup

### Prerequisites

- Go 1.23+
- Docker and Docker Compose
- Node.js 18+ (SDK development only)
- [`golangci-lint`](https://golangci-lint.run/usage/install/)
- [`sqlc`](https://docs.sqlc.dev/en/latest/overview/install.html)

### 1. Clone and configure

```bash
git clone git@github.com:Godie360/custos.git
cd custos
cp .env.example .env
# Edit .env with your local values
```

### 2. Start the dev stack

```bash
make dev
```

This starts PostgreSQL and Kafka in Docker, then runs the Go server with Air hot reload.

### 3. Apply migrations

```bash
make migrate-up
```

---

## Making Changes

### Go server

- Follow the package dependency rule strictly:
  ```
  domain ← store ← service ← api
  domain ← queue ← service
  domain ← provider ← service
  domain ← notification ← service
  ```
- `internal/domain` must have zero imports from other internal packages.
- Add new behaviour in `service/`, expose it via `api/handler/`.
- Write new SQL queries in `internal/store/postgres/queries/*.sql`, then run `sqlc generate`.

### Node.js SDK

```bash
cd sdks/nodejs
npm install
npm test          # must pass before opening a PR
npm run build     # ensure TypeScript compiles cleanly
```

---

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): short description

Optional longer body explaining the why, not the what.
```

**Types:** `feat` · `fix` · `chore` · `docs` · `test` · `refactor` · `perf`

**Scopes:** `api` · `queue` · `store` · `provider` · `notification` · `sdk` · `docker` · `migrations` · `docs`

**Examples:**

```
feat(sdk): add Pino stream transport
fix(store): handle null stack_trace in ListIssues
chore(docker): pin postgres image to 16-alpine
docs(api): add example responses to openapi.yaml
```

Keep the subject line under 72 characters. Use the body to explain context, not mechanics.

---

## Pull Requests

1. Target `develop`, never `main`.
2. Fill in the pull request template — summary and test plan are required.
3. All CI checks must pass (Go tests, lint, SDK tests).
4. Request a review from a maintainer.
5. Squash and merge is preferred to keep `develop` history readable.

### PR checklist

- [ ] Tests added or updated for the change
- [ ] `go test ./...` passes locally
- [ ] `golangci-lint run ./...` passes locally
- [ ] No new secrets or credentials committed
- [ ] PR description explains the why, not just the what

---

## Code Standards

### Go

- `gofmt` and `goimports` must be clean (enforced by CI).
- No naked `panic()` calls outside of `main`.
- Return errors — do not log and swallow them in library code.
- Keep handlers thin: validation only. Business logic belongs in `service/`.
- Do not add comments that restate what the code already says. Only comment non-obvious decisions.

### TypeScript (SDK)

- `strict: true` is enabled — no `any` unless genuinely unavoidable.
- The SDK must **never** throw an uncaught exception into the host application. Wrap all outbound calls in try/catch.
- Redact all sensitive data locally before any HTTP transmission.

---

## Testing

### Go

```bash
go test ./...                  # all packages
go test ./internal/service/... # specific package
```

Integration tests that require PostgreSQL or Kafka are skipped automatically when the services are not available.

### Node.js SDK

```bash
cd sdks/nodejs
npx jest                       # all tests
npx jest tests/redact.test.ts  # specific file
```

---

## Project Structure

```
custos/
├── cmd/server/          # main() — wires everything together
├── internal/
│   ├── api/             # HTTP layer (handlers, middleware, router)
│   ├── config/          # Config loaded from environment variables
│   ├── domain/          # Shared types — no internal deps allowed
│   ├── notification/    # Outbound alerts (Google Chat, email, webhook)
│   ├── provider/        # AI adapter registry and adapters
│   ├── queue/           # Kafka producer and consumer
│   ├── service/         # Business logic
│   └── store/           # Database access (sqlc-generated + wrappers)
├── migrations/          # Numbered SQL migration files
├── api/                 # openapi.yaml
├── sdks/nodejs/         # @custos/sdk npm package
├── docs/
│   ├── design/          # Approved product and system truth
│   ├── implementation/  # Execution plans, phases, task status
│   └── changes/         # Proposals under review
├── docker/              # Dockerfiles
└── .github/             # CI workflows and PR template
```

---

## Getting Help

- Open a GitHub issue for bugs or feature requests.
- Start a GitHub Discussion for questions or ideas.

We appreciate every contribution, no matter how small.
