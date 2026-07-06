# Project Structure

## Context

The repository must be structured so that external contributors can orient themselves quickly, understand boundaries between components, and contribute to any layer without needing to understand the whole system.

## Requirements

- Clear separation between the Go server, TanStack dashboard, and SDKs.
- Go code follows the standard community layout (`cmd/`, `internal/`, `pkg/`).
- Each SDK lives in its own directory under `sdks/` with its own package manifest and README.
- Infrastructure and tooling files live at the root alongside a `Makefile` for common tasks.
- Database migrations are versioned and co-located with the server code.
- A new contributor can identify the correct file to edit for any component within 2 minutes of reading the structure.

## Decisions

- `domain/` is the dependency root — zero imports from other internal packages. All core types and interfaces live here.
- `internal/` for all Go packages not intended for external import.
- `pkg/` for public Go packages safe to import externally (shared event payload schema; future Go SDK).
- `cmd/server/` as the single binary entry point — wires all dependencies.
- `service/` layer owns business logic; handlers are thin.
- Interfaces live in the package that **uses** them, not the package that implements them (Go convention).
- No `utils`, `helpers`, or `common` packages — name packages by what they do.

## Package Dependency Rule

```
domain ← store ← service ← api
domain ← queue ← service
domain ← provider ← service
domain ← notification ← service
```

`domain` has zero dependencies on other internal packages. No package imports its sibling (`store` never imports `service`; `api` never imports `store` directly).

## Contracts

```
custos/
│
├── cmd/
│   └── server/
│       └── main.go                        # Wire dependencies, start server
│
├── internal/
│   │
│   ├── domain/                            # Core types and interfaces — zero internal deps
│   │   ├── event.go                       # Event, RawEvent structs
│   │   ├── issue.go                       # Issue, Fingerprint, Status types
│   │   ├── project.go                     # Project, APIKey types
│   │   ├── analysis.go                    # AIAnalyzer interface, AnalysisEvent, AnalysisResult
│   │   ├── notification.go                # Notifier interface, AlertPayload
│   │   └── errors.go                      # Sentinel errors (ErrNotFound, ErrUnauthorized …)
│   │
│   ├── api/                               # HTTP layer — Chi router, handlers, middleware
│   │   ├── router.go                      # Route registration, middleware chain
│   │   ├── handler/
│   │   │   ├── ingest.go                  # POST /api/v1/ingest
│   │   │   ├── issues.go                  # GET /api/v1/issues, GET /api/v1/issues/:id
│   │   │   ├── analytics.go               # GET /api/v1/analytics/*
│   │   │   └── projects.go                # Project + API key management
│   │   └── middleware/
│   │       ├── apikey.go                  # API key auth
│   │       ├── requestid.go               # Inject X-Request-ID
│   │       └── logger.go                  # Structured request logging (slog)
│   │
│   ├── service/                           # Business logic — orchestrates domain + store + queue
│   │   ├── ingestion.go                   # Receive, filter, fingerprint, deduplicate
│   │   ├── analysis.go                    # Analysis worker: read queue → AI → store
│   │   └── notification.go                # Evaluate rules, dispatch alerts
│   │
│   ├── store/                             # Data access — interface + implementations
│   │   ├── store.go                       # Store interface (IssueStore, EventStore, ProjectStore)
│   │   └── postgres/
│   │       ├── postgres.go                # *sql.DB setup, migration runner
│   │       ├── issues.go                  # IssueStore implementation
│   │       ├── events.go                  # EventStore implementation
│   │       └── projects.go                # ProjectStore implementation
│   │
│   ├── queue/                             # Message queue — interface + Kafka implementation
│   │   ├── queue.go                       # Producer and Consumer interfaces
│   │   └── kafka/
│   │       ├── producer.go                # segmentio/kafka-go producer
│   │       └── consumer.go                # segmentio/kafka-go consumer
│   │
│   ├── provider/                          # AI adapters — implement domain.AIAnalyzer
│   │   ├── registry.go                    # Load(cfg) returns correct AIAnalyzer or error
│   │   ├── claude/
│   │   │   └── claude.go
│   │   ├── openai/
│   │   │   └── openai.go
│   │   ├── gemini/
│   │   │   └── gemini.go
│   │   └── ollama/
│   │       └── ollama.go
│   │
│   ├── notification/                      # Notification delivery — implement domain.Notifier
│   │   ├── slack/
│   │   │   └── slack.go
│   │   ├── email/
│   │   │   └── email.go
│   │   └── webhook/
│   │       └── webhook.go
│   │
│   └── config/
│       └── config.go                      # Load all config from env vars into a Config struct
│
├── pkg/                                   # Public packages safe to import externally
│   └── event/
│       └── payload.go                     # Shared SDK event payload schema (JSON)
│
├── migrations/                            # golang-migrate versioned SQL files
│   ├── 000001_init_schema.up.sql
│   └── 000001_init_schema.down.sql
│
├── dashboard/                             # TanStack Start app
│   ├── src/
│   │   ├── routes/                        # TanStack Router file-based routes
│   │   ├── components/                    # Reusable UI components (shadcn/ui)
│   │   ├── lib/
│   │   │   ├── api.ts                     # API client (TanStack Query hooks)
│   │   │   └── types.ts                   # Shared TypeScript types
│   │   └── main.tsx
│   ├── package.json
│   └── tsconfig.json
│
├── sdks/
│   ├── python/
│   │   ├── custos/
│   │   │   ├── __init__.py
│   │   │   ├── handler.py                 # logging.Handler implementation
│   │   │   └── redact.py                  # Redaction logic
│   │   ├── tests/
│   │   ├── pyproject.toml
│   │   └── README.md
│   ├── nodejs/
│   │   ├── src/
│   │   │   ├── index.ts
│   │   │   ├── transport.ts               # Winston/Pino transport
│   │   │   └── redact.ts
│   │   ├── tests/
│   │   ├── package.json
│   │   └── README.md
│   └── java/
│       ├── src/main/java/io/custos/
│       │   ├── CustosAppender.java        # Logback/Log4j2 appender
│       │   └── Redactor.java
│       ├── src/test/
│       ├── pom.xml
│       └── README.md
│
├── docker/
│   ├── Dockerfile.server
│   └── Dockerfile.dashboard
│
├── .github/
│   ├── workflows/
│   │   ├── ci.yml                         # Test + lint on every PR
│   │   └── release.yml                    # Build + publish on tag
│   └── PULL_REQUEST_TEMPLATE.md
│
├── docker-compose.yml                     # Self-hosting: server + dashboard + Postgres + Kafka
├── docker-compose.dev.yml                 # Dev overrides: Air hot reload, exposed ports
├── Makefile
├── .golangci.yml                          # golangci-lint config
├── .air.toml                              # Air hot reload config (dev only)
├── go.mod
├── go.sum
├── AGENTS.md
├── CLAUDE.md -> AGENTS.md
├── CONTRIBUTING.md
└── README.md
```

## Go Tooling

| Tool | Purpose | License |
|---|---|---|
| `golangci-lint` | Linter aggregator (errcheck, staticcheck, govet, unused…) | GPL-3.0 |
| `air` | Hot reload for Go in development | MIT |
| `golang-migrate` | Versioned DB migrations (CLI + library) | MIT |
| `testify` | Test assertions and mocks | MIT |
| `slog` | Structured logging (stdlib, Go 1.21+) | BSD (stdlib) |
| `segmentio/kafka-go` | Kafka producer/consumer | MIT |

## Makefile Targets

| Target | Purpose |
|---|---|
| `make dev` | Start full stack with hot reload (docker-compose.dev.yml) |
| `make build` | Build the Go binary |
| `make test` | Run all Go tests |
| `make test-cover` | Run tests with coverage report |
| `make lint` | Run golangci-lint |
| `make migrate-up` | Apply pending DB migrations |
| `make migrate-down` | Roll back last migration |
| `make migrate-create name=<n>` | Create a new versioned migration file |
| `make generate` | Run go generate (mocks, etc.) |

## Package Rules for Contributors

- No `utils`, `helpers`, or `common` packages — name packages by what they do.
- No package imports its sibling — `service` never imports `api`; `store` never imports `service`.
- `domain` imports nothing from `internal/` — it is the dependency root.
- Interfaces live in the package that **uses** them, not the package that implements them.
- Every exported type and function in `domain/` must have a doc comment.

## Acceptance Criteria

- [ ] `go build ./cmd/server/` succeeds from the repo root.
- [ ] `docker compose up` starts server, dashboard, Postgres, and Kafka with no manual steps.
- [ ] Each SDK directory contains its own README, install instructions, and test suite.
- [ ] `make test` runs the full Go test suite from the repo root.
- [ ] `make lint` runs with zero errors on a clean checkout.
- [ ] A new contributor can identify the correct file to edit for any component within 2 minutes of reading this document.
