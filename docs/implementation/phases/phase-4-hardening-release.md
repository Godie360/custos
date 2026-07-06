# Phase 4 — Hardening & Open-Source Release

## Status

- `pending`
- Last updated: 2026-07-06

## Objective

Make Custos genuinely self-hostable by any developer in under 15 minutes, finalize the open-source posture, and publish the repository.

## Scope

Weeks 12–14.

## Features

- Self-hosting quickstart: Docker Compose, one-command setup.
- BYO AI key configuration and documentation.
- Data retention controls: configurable automatic deletion of raw log bodies.
- Project README, CONTRIBUTING.md, architecture overview.
- License selection and public repository launch.

## Tasks

- [ ] Write and test Docker Compose self-hosting quickstart (server + dashboard + Postgres + Kafka).
- [ ] Implement and document `CUSTOS_AI_API_KEY` BYO key configuration.
- [ ] Implement data retention job: configurable TTL, deletes `events.raw_body` only.
- [ ] Write README: what it is, quickstart, configuration reference, architecture diagram.
- [ ] Write CONTRIBUTING.md: local dev setup, PR workflow, Makefile targets, code style.
- [ ] Select and apply license (MIT or Apache 2.0).
- [ ] Security review: no hardcoded secrets, all endpoints authenticated, key never logged.
- [ ] External developer test: follow README cold, self-host in under 15 minutes.
- [ ] Public repository launch.

## Acceptance Criteria

- [ ] `docker compose up` starts full stack with no manual steps beyond env vars.
- [ ] External developer (not on the team) self-hosts in under 15 minutes from README alone.
- [ ] BYO AI key documented and tested; key never stored in DB or appears in logs.
- [ ] Retention job verified: `raw_body` deleted, issue metadata and AI explanations preserved.
- [ ] No hardcoded secrets in the published codebase.

## Blockers

- Phases 1–3 complete and stable.
- License decision confirmed.

## Linked Tasks

- `docs/implementation/tasks/backlog.md`
