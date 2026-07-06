# Backlog

## Status

pending

## Objective

Phase 1 initial tasks. Tasks for Phases 2–4 will be broken into individual task files as each phase begins.

## Implementation Checklist

- [ ] Initialize Go module and project structure per `docs/design/architecture/project-structure.md`.
- [ ] Write PostgreSQL schema migration files using `golang-migrate`.
- [ ] Stand up Apache Kafka in Docker Compose; create `custos.events` and `custos.analysis` topics.
- [ ] Implement `POST /api/v1/ingest` with Chi router and API key middleware.
- [ ] Implement `queue.Producer` interface backed by `segmentio/kafka-go`.
- [ ] Build first SDK — decision needed: Python / NestJS / Spring Boot.
- [ ] SDK: error capture, local redaction, async batching, retry with exponential backoff.
- [ ] SDK: unit tests for redaction with synthetic secrets.
- [ ] Implement `AIAnalyzer` interface and Claude API adapter.
- [ ] Implement analysis worker: Kafka consumer → Claude API → write result to DB.
- [ ] Implement Slack notifier: format and POST explanation to webhook.
- [ ] Docker Compose: server, PostgreSQL, Kafka, with env var template (`.env.example`).
- [ ] Smoke test: trigger error in test service → verify Slack message received end-to-end.
