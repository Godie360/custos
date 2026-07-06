# Custos — Domain Overview

## Context

Custos is an open-source, SDK-based log intelligence platform. Developers install a lightweight SDK in their application; the SDK captures errors automatically, forwards them to a central service, and an AI engine explains what happened, why it likely happened, and what to check next.

Name: "Custos" is Latin for guardian, keeper, or watchman.

## Problem

- **Fragmented log formats** — every framework logs errors differently; no consistent cross-service view.
- **Manual, reactive triage** — developers copy raw log output into a separate AI chat tool by hand; slow and unscalable for a team.
- **Delayed incident response** — without automated detection and alerting, outages are noticed late, often by users rather than the system.

## Requirements

- One SDK per service → automatic plain-language error explanations with no manual copy-pasting.
- Support mixed stacks (Python, Node.js/NestJS, Java/Spring Boot) from a common backend and dashboard.
- Admin dashboard with real-time analytics: error trends, affected services, severity breakdown, uptime signals.
- Detect security-relevant patterns (repeated auth failures, injection attempts) and flag them distinctly from ordinary bugs.
- Notify the right person automatically when a service goes down or a critical error occurs.
- Ship as a genuinely self-hostable open-source project; one-command local setup.

## Decisions

- Backend server: **Go** — best balance of raw throughput, low resource cost, and contributor approachability for open source.
- HTTP router: **Chi** — stays close to stdlib `net/http`; no framework-specific abstractions; proven for high-concurrency small-request workloads.
- Dashboard frontend: **TypeScript + React (Next.js)** — rich interactive analytics UI.
- AI provider: **Claude API (Anthropic)** — configurable; self-hosters can bring their own key.
- Each SDK is written in the native language of the framework it instruments.

## Contracts

- No behavior is defined by implementation code alone — all decisions flow through approved design docs.
- SDKs never block or crash the host application.
- No sensitive data leaves the host before local SDK redaction runs.

## Acceptance Criteria

- [ ] One office service's errors explained in plain language automatically, with no manual copy-paste.
- [ ] All in-use stacks (Python, NestJS, Spring Boot) report into one dashboard.
- [ ] Critical errors auto-notify the right person.
- [ ] External developer can self-host via Docker Compose in under 15 minutes.
