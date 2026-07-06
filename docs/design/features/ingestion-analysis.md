# Ingestion & Analysis Server

## Context

Central Go service that receives events from all SDKs, filters and deduplicates them, coordinates AI analysis, persists results, and exposes the data to the dashboard and notification service.

## Requirements

- `POST /api/v1/ingest` returns 202 for valid events, 401 for missing/invalid key, 400 for bad payload.
- Filtering stage must prevent noisy events from reaching the analysis queue.
- Deduplication must not create duplicate issues for the same fingerprint.
- AI analysis worker must survive Claude API errors without data loss.
- Notification must be delivered within 60 seconds of a critical event being analyzed.

## Decisions

- **Kafka** for the queue — Apache Kafka (Apache 2.0); Go client: `segmentio/kafka-go` (MIT). Durable, replayable, and the final queue choice (not a stepping stone).
- Queue interface is abstracted behind a Go interface so the backing implementation can be swapped without touching ingestion or worker code.
- AI analysis is in a separate worker process that reads from Kafka — decouples ingestion throughput from AI processing rate.

## Contracts

### Ingestion API

- Route: `POST /api/v1/ingest`
- Auth: `X-Custos-Key: <api_key>`
- Body: SDK event payload (see SDK Design doc).
- Response: `202 Accepted` (async processing).

### Filtering Stage

Rules applied in order:
1. Minimum severity threshold (configurable per project; default: `error`).
2. Known-noisy pattern blocklist (configurable regex list per project).
3. Per-project AI analysis rate limit (max analyses per hour; configurable).

Events that fail filtering: stored in DB as raw events; not published to analysis topic.

### Deduplication

- Fingerprint: `SHA256(error_type + normalized_stack_frames)`.
- Known issue → increment `occurrence_count`, update `last_seen`. No re-analysis unless severity escalated.
- New issue → insert issue record, publish to Kafka analysis topic.

### AI Analysis Worker

- Reads from Kafka `custos.analysis` topic.
- Builds prompt: error type, message, top stack frames, service context.
- Calls Claude API (`AIAnalyzer` interface — default implementation: Claude).
- Structured response: `explanation`, `likely_cause`, `severity`, `suggested_checks`.
- Writes result to `issues` table.
- On failure (3 retries with backoff): marks issue `analysis_failed`; surfaced in dashboard for manual retry.

### Notification Service

- Triggered after AI analysis writes to DB.
- Rule dimensions: severity level, service name, project ownership.
- Delivery targets: Slack webhook, SMTP email, generic HTTPS webhook.
- Target: notification delivered within 60 seconds of analysis completing.

### Public REST API

Key endpoints:

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/issues` | List issues (filterable by project, service, severity, time) |
| GET | `/api/v1/issues/:id` | Issue detail with AI explanation |
| GET | `/api/v1/analytics/summary` | Error rate, top services, severity breakdown |
| GET | `/api/v1/projects` | List projects for authenticated user |
| POST | `/api/v1/projects/:id/keys` | Create API key |
| DELETE | `/api/v1/projects/:id/keys/:kid` | Revoke API key |

## Acceptance Criteria

- [ ] `POST /ingest` returns correct status codes for all auth and validation cases.
- [ ] Filtering stage: noisy events stored but not published to analysis topic.
- [ ] Deduplication: same fingerprint does not produce a second issue record.
- [ ] AI worker: Claude API 429/500 handled with retry; no event lost after 3 failures (marked `analysis_failed`).
- [ ] Notification delivered within 60 seconds of critical event analysis.
