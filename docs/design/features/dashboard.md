# Admin Dashboard

## Context

Web-based interface for teams and administrators to monitor system health, investigate issues, configure alerts, and manage projects.

## Requirements

- Issue feed must update without full page reload.
- Issue detail must render AI explanation alongside raw log.
- Analytics charts must load within 2 seconds for a 30-day range on PostgreSQL.
- Alerting configuration changes must take effect within one analysis cycle.
- Unauthenticated requests must redirect to login.

## Decisions

- **TanStack Start** — full-stack React framework built on TanStack Router; type-safe file-based routing; familiar React mental model; easier for contributors not coming from a Next.js background.
- **TanStack Router** (MIT) — type-safe, file-based routing.
- **TanStack Query** (MIT) — async data fetching, caching, and background sync.
- **TanStack Table** (MIT) — headless table primitives for the issue feed and analytics tables.
- **shadcn/ui** (MIT) — accessible, unstyled component primitives; easy to customize.
- **Recharts** (MIT) — composable chart library for analytics views.
- Dashboard consumes the Go server's Public REST API exclusively — no direct DB access.

## Contracts

### Live Issue Feed
- Groups issues by fingerprint, not raw occurrence count.
- Filterable by: service, environment, severity, time range.
- TanStack Query polling or WebSocket for near-real-time updates.

### Issue Detail View
- Raw log body (redacted).
- AI explanation, likely cause, and suggested checks.
- Affected service and environment.
- First seen / last seen timestamps.
- Occurrence count over time (sparkline via Recharts).

### Analytics Overview
- Error rate over time (line chart — Recharts).
- Top failing services (ranked list — TanStack Table).
- Severity breakdown (bar chart — Recharts).
- Mean time to detection (MTTD).

### Security Panel
- Flagged authentication anomalies.
- Suspicious access patterns.
- Injection-style signatures.
- Visually distinct from ordinary bug issues (separate section, different badge color).

### Alerting Configuration
- Per-project thresholds: severity level, error rate.
- Notification channel setup: Slack webhook URL, SMTP config, generic webhook.
- Per-service ownership assignment.

### Project & Team Management
- API key create, label, rotate, revoke.
- Team member invite and role assignment.
- Roles: `admin`, `member`, `viewer`.

## Acceptance Criteria

- [ ] Issue feed updates without full page reload.
- [ ] Issue detail renders AI explanation alongside raw log.
- [ ] Analytics charts load within 2 seconds for 30-day range on PostgreSQL.
- [ ] Alerting config changes take effect within one analysis cycle.
- [ ] Unauthenticated requests redirect to login.
