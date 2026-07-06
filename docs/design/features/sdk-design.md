# SDK Design

## Context

One SDK per supported stack. SDKs are the only Custos component installed inside customer applications. They must be invisible in normal operation and must never cause a failure that would not have occurred without them.

## Requirements

- Hook into the framework's existing logging mechanism — no changes to business logic required.
- Capture: stack trace, error message, service name, environment, timestamp, request context where available.
- Run local regex-based redaction before any data leaves the host.
- Batch events and send asynchronously; never block the calling thread.
- Queue locally and retry with exponential backoff if the ingestion server is unreachable.
- Never throw an uncaught exception into the host application.

## Decisions

- SDKs ship in the native language of the framework they instrument — makes installation frictionless.
- Redaction runs locally in the SDK, not the server.
- Async batching with local retry fully decouples SDK availability from server availability.
- All SDK packages are open source and published to the standard registry for their language.

## Contracts

### Python SDK
- Integration: custom `logging.Handler` subclass.
- Works across Django, Flask, FastAPI — all use stdlib `logging`.
- Published: PyPI as `custos-sdk`.
- License: MIT.

### Node.js / NestJS SDK
- Integration: Winston or Pino transport, or NestJS global exception filter.
- Captures unhandled exceptions and HTTP error responses.
- Published: npm as `@custos/sdk`.
- License: MIT.

### Java / Spring Boot SDK
- Integration: Logback `Appender` or Log4j2 `Appender`.
- Plugs into existing appender configuration; no business logic changes.
- Published: Maven Central as `io.custos:custos-sdk`.
- License: Apache 2.0.

### Shared Event Payload Schema

```json
{
  "service": "string",
  "environment": "string",
  "error_type": "string",
  "message": "string",
  "stack_trace": ["string"],
  "timestamp": "ISO 8601",
  "sdk_version": "string"
}
```

## Acceptance Criteria

- [ ] SDK installation documented in under 5 lines of user code per stack.
- [ ] Redaction unit tests confirm secrets are stripped before the HTTP call is made.
- [ ] Integration test confirms events are queued and retried when server returns 503.
- [ ] No SDK exception propagates to the host application under any error condition.
