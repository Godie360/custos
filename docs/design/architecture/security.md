# Security & Privacy Design

## Context

Custos handles application error data that may include sensitive context. Privacy and security must be enforced at the SDK level before data leaves the host, not at the server.

## Requirements

- SDK redaction must strip: authorization headers, bearer tokens, passwords, connection strings, and common secret patterns before any HTTP transmission.
- API keys must be rotatable without downtime.
- Dashboard access must require authentication; role-based access control for team management.
- Raw log bodies must be independently deletable (retention) without cascading to issue metadata or AI analysis.
- Self-hosters must be able to supply their own AI API key; Custos must never store or log it.

## Decisions

- **Local redaction in the SDK** — the server must never receive unredacted data. Redaction is not optional and cannot be disabled.
- **Per-project API keys** — scoped and rotatable; all SDK-to-server traffic is authenticated via `X-Custos-Key` header.
- **TLS on all network hops** — including internal hops where practical.
- **BYO AI key** — self-hosters set `CUSTOS_AI_API_KEY` env var; key is read at runtime, never stored in DB, never logged.
- **Configurable data retention** — `retention_delete_at` on raw event bodies; deletion job runs on schedule without touching issue metadata.

## Contracts

- API key rotation: new key is issued before the old key is revoked; no gap in SDK delivery.
- Retention deletion: targets `events.raw_body` only; `issues` and `events.ai_explanation` are preserved.
- Auth: all dashboard routes and API endpoints return 401 for unauthenticated requests.

## Acceptance Criteria

- [ ] SDK redaction verified by unit test — synthetic secrets absent from outgoing HTTP payload.
- [ ] API key rotation tested: events delivered during rotation window with no 401s.
- [ ] All endpoints return 401 for missing or invalid credentials.
- [ ] Retention job deletes raw bodies; issue metadata and AI explanations remain intact.
- [ ] `CUSTOS_AI_API_KEY` never appears in application logs or error messages.
