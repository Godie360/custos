# Phase 1 — Foundations

## Status

- `pending`
- Last updated: 2026-07-06

## Objective

Deliver an end-to-end vertical slice: one office service's errors captured by the SDK, ingested by the Go server, analyzed by Claude, and posted to Slack.

## Scope

Weeks 1–3.

## Features

- Go server skeleton with Chi router and project directory structure.
- PostgreSQL schema: projects, api_keys, events, issues.
- Apache Kafka setup: `custos.events` and `custos.analysis` topics.
- First SDK (stack TBD — Python / NestJS / Spring Boot).
- `POST /api/v1/ingest` endpoint with API key authentication.
- AI analysis worker: reads from Kafka, calls Claude API, writes result to DB.
- Slack notifier: posts AI explanation to configured webhook.
- Docker Compose: server + PostgreSQL + Kafka.

## Tasks

- [ ] Initialize Go module and project directory structure per `docs/design/architecture/project-structure.md`.
- [ ] Write PostgreSQL schema migration files (golang-migrate).
- [ ] Stand up Kafka in Docker Compose; create `custos.events` and `custos.analysis` topics.
- [ ] Implement `POST /api/v1/ingest` with Chi and API key middleware.
- [ ] Implement `queue.Producer` interface backed by `segmentio/kafka-go`.
- [ ] Build first SDK with error capture, local redaction, async batch send, and local retry.
- [ ] SDK: unit tests for redaction with synthetic secrets.
- [ ] Implement `AIAnalyzer` interface and Claude API adapter.
- [ ] Implement analysis worker: Kafka consumer → Claude API → write to DB.
- [ ] Implement Slack notifier: format and POST explanation to webhook URL.
- [ ] Docker Compose: server, PostgreSQL, Kafka, Zookeeper with env var template.
- [ ] Smoke test: trigger error in test service → verify Slack message received.

## Acceptance Criteria

- [ ] `POST /ingest` returns 202 (valid), 401 (bad key), 400 (bad payload).
- [ ] SDK redaction unit tests pass with synthetic secrets — secrets absent from outgoing payload.
- [ ] One real office service error produces a Slack explanation automatically end-to-end.
- [ ] `docker compose up` starts all services with no manual steps beyond setting env vars.

## Blockers

- Decision needed: which stack gets the first SDK? (Python / NestJS / Spring Boot)
- Required: Claude API key (`CUSTOS_AI_API_KEY`).
- Required: Slack webhook URL for office notification channel.

## Linked Tasks

- `docs/implementation/tasks/backlog.md`
