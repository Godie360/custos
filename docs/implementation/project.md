# Project Implementation

## Overview

Custos — open-source, AI-powered log intelligence platform. Go + Chi backend. Next.js dashboard. Three SDKs (Python, Node.js/NestJS, Java/Spring Boot). Apache Kafka queue. PostgreSQL + ClickHouse storage. Claude API for AI analysis.

## Current Priorities

- Phase 1: Go ingestion server, first SDK, Kafka queue, Claude API integration, Slack delivery.

## Active Phases

- [ ] [Phase 1 — Foundations](phases/phase-1-foundations.md) — Weeks 1–3
- [ ] [Phase 2 — Core Platform](phases/phase-2-core-platform.md) — Weeks 4–7
- [ ] [Phase 3 — Analytics & Alerting](phases/phase-3-analytics-alerting.md) — Weeks 8–11
- [ ] [Phase 4 — Hardening & Open-Source Release](phases/phase-4-hardening-release.md) — Weeks 12–14

## Deferred Phases

- [ ] ClickHouse migration (post-release, when log volume outgrows PostgreSQL)
- [ ] Kubernetes deployment manifests (post-release)

## Milestones

| Milestone | Week | Success Criteria |
|---|---|---|
| First SDK + explanation | 3 | One office service's errors explained in Slack automatically. |
| Multi-stack + dashboard v1 | 7 | All stacks report in; team views issues in browser. |
| Analytics + alerting | 11 | Critical errors auto-notify; trends visible. |
| Public open-source release | 14 | External dev self-hosts via Docker Compose in under 15 min. |

## Dependencies

- Claude API key (`CUSTOS_AI_API_KEY`)
- PostgreSQL instance
- Apache Kafka instance
- Slack webhook URL (Phase 1 notifications)

## Linked Artifacts

- phases: `docs/implementation/phases/`
- tasks: `docs/implementation/tasks/backlog.md`
- status: `docs/implementation/status/weekly-status.md`
- design: `docs/design/`
