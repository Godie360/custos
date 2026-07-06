# Phase 2 — Core Platform

## Status

- `pending`
- Last updated: 2026-07-06

## Objective

Expand SDK coverage to all in-use stacks, add filtering and deduplication so only meaningful issues reach the AI engine, and ship the first web dashboard.

## Scope

Weeks 4–7.

## Features

- Second and third SDKs (remaining stacks from Phase 1 decision).
- Filtering stage in the ingestion pipeline (severity threshold, blocklist, rate limit).
- Fingerprint-based deduplication and issue grouping.
- Next.js dashboard scaffold.
- Dashboard: live issue feed and issue detail view.
- Dashboard authentication (session-based).
- API key management UI.

## Tasks

- [ ] Build second SDK.
- [ ] Build third SDK.
- [ ] Implement filtering stage: severity threshold, noisy-pattern blocklist, per-project rate limit.
- [ ] Implement fingerprint computation and deduplication logic.
- [ ] Scaffold TanStack Start dashboard with TanStack Router, TanStack Query, TanStack Table, shadcn/ui, and Recharts.
- [ ] Build live issue feed page (grouped by fingerprint, filterable).
- [ ] Build issue detail page (raw log, AI explanation, occurrence graph).
- [ ] Implement dashboard authentication (session tokens, login/logout).
- [ ] Build API key management UI: create, label, revoke.
- [ ] Add `GET /api/v1/issues` and `GET /api/v1/issues/:id` REST endpoints.

## Acceptance Criteria

- [ ] All in-use office stacks report errors into the same issue feed.
- [ ] Repeated errors increment occurrence count; no duplicate issues created.
- [ ] Dashboard live feed shows new issues without full page reload.
- [ ] Issue detail renders AI explanation alongside raw log body.
- [ ] Unauthenticated dashboard requests redirect to login.
- [ ] Filtering stage: noisy events stored but not published to analysis topic.

## Blockers

- Phase 1 complete and stable.
- SDK build order decision (which stack second / third).

## Linked Tasks

- `docs/implementation/tasks/backlog.md`
