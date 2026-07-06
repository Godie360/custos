# System Architecture

## Context

Three cooperating components: language SDKs installed inside customer applications, a central ingestion and AI-analysis server written in Go, and a Next.js admin dashboard.

## Requirements

- Ingestion API must handle concurrent SDK connections without dropping events.
- Filtering stage must run before any event reaches the AI engine.
- Deduplication must group repeated errors into a single issue.
- Public API must serve the dashboard and external integrations.

## Decisions

- **Go + Chi** for the backend server — concurrency model (goroutines), single-binary distribution, low memory footprint, approachable for open-source contributors. Chi stays close to stdlib `net/http` with no framework abstractions.
- **Redis Streams** for initial queuing — simplest durable queue; Kafka-compatible interface abstracts the implementation for future migration.
- **Next.js** for the dashboard — separate deployable from the Go server; consumes the Public REST API.

## Contracts

### A. Language SDKs

| Stack | Integration Point | Notes |
|---|---|---|
| Python | Custom `logging.Handler` | Works across Django, Flask, FastAPI via stdlib logging. |
| Node.js / NestJS | Winston/Pino transport or NestJS exception filter | Captures unhandled exceptions and HTTP error responses. |
| Java / Spring Boot | Logback or Log4j2 Appender | Plugs into existing appender config; no changes to business logic. |

Shared SDK contract:
- Capture: stack trace, error message, service name, environment, timestamp.
- Run local redaction before transmission.
- Batch and send asynchronously; queue locally and retry on failure.
- Never throw an uncaught exception into the host application.

### B. Ingestion & Analysis Server

| Sub-component | Responsibility |
|---|---|
| Ingestion API | Authenticate SDK traffic (per-project API key); write events to queue. |
| Filtering stage | Severity threshold, noisy-pattern blocklist, per-project rate limit. |
| Deduplication | Fingerprint = error type + normalized stack trace shape; group into issues. |
| AI Analysis Engine | Call Claude API; store explanation, likely cause, severity. |
| Notification Service | Route critical events to Slack, email, or webhook per routing rules. |
| Public API | REST endpoints for dashboard and external integrations. |

### C. Admin Dashboard (Next.js)

- Live issue feed grouped by fingerprint.
- Issue detail: raw log, AI explanation, occurrence graph.
- Analytics: error rate, top services, severity breakdown, MTTD.
- Security panel: auth anomalies, suspicious patterns, injection signatures.
- Alerting config, API key management, team/role management.

## Technology Stack

| Component | Technology | Alternative |
|---|---|---|
| Backend server | Go | — |
| HTTP router | Chi | Gin, Echo, Fiber |
| Message queue | Apache Kafka | — |
| Relational DB | PostgreSQL | MySQL |
| Log/event store | ClickHouse or TimescaleDB | Plain PostgreSQL (small scale) |
| Dashboard | TypeScript + React (TanStack Start) | — |
| AI provider | User-configured adapter (Claude, OpenAI, Gemini, Ollama) | — |
| Deployment | Docker Compose (self-host), Kubernetes (scale) | — |

## Acceptance Criteria

- [ ] Ingestion API handles concurrent SDK connections without dropping events.
- [ ] Filtering stage prevents unanalyzed noise from reaching the AI engine.
- [ ] Deduplication groups repeated errors into a single issue.
- [ ] AI explanation stored alongside raw event and surfaced in the dashboard.
