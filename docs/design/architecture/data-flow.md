# Data Flow

## Context

End-to-end path an event takes from a running service to the dashboard and notification targets.

## Requirements

- No event reaches the AI engine without passing the filtering stage.
- Local SDK redaction must run before any HTTP transmission.
- Events must be retried locally if the server is unreachable.
- Deduplication must prevent repeated AI calls for the same issue fingerprint.

## Decisions

- Redaction runs in the SDK, not the server — the server must never be trusted with unredacted data.
- Queue sits between ingestion and analysis — decouples SDK traffic spikes from AI processing rate.

## Contracts

- **SDK → Ingestion API**: HTTPS POST, JSON payload, `X-Custos-Key` header.
- **Ingestion API → Queue**: Redis Streams XADD.
- **Queue → AI Analysis Worker**: Redis Streams XREAD; worker calls Claude API, writes result to DB.
- **Server → Dashboard**: REST API over HTTPS; session or API key auth.
- **Server → Notification targets**: Slack webhook, SMTP, or generic HTTPS webhook.

## Flow

1. Service runs with Custos SDK attached to its existing logger.
2. SDK captures error — stack trace, message, service name, environment, timestamp.
3. SDK runs local redaction pass (regex) to strip secrets and PII.
4. SDK batches event and POSTs asynchronously to `POST /api/v1/ingest`.
5. Ingestion API authenticates request; validates payload; writes to queue.
6. Filtering stage reads from queue; applies severity threshold, blocklist, and rate limit. Events that fail filtering are stored raw but not enqueued for analysis.
7. Surviving events are fingerprinted and checked against known issues. Known issue: increment occurrence count, update last_seen. New issue: create issue record, enqueue for AI analysis.
8. AI Analysis Worker reads from analysis queue; calls Claude API; writes explanation, likely cause, and severity to the issue record.
9. Notification Service evaluates routing rules; delivers alerts for critical events.
10. Dashboard queries stored data via Public REST API.

## Acceptance Criteria

- [ ] SDK redaction verified by unit test with synthetic secrets before HTTP call.
- [ ] Events queued locally and retried when server returns 503.
- [ ] No event in analysis queue bypassed the filtering stage.
- [ ] Same fingerprint does not trigger duplicate AI analysis calls.
