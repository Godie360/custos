# Custos

**Custos** is an open-source, AI-powered log intelligence platform. Language SDKs capture errors from your applications, forward them to a central server, and an AI worker explains what went wrong — severity, likely cause, and suggested checks — so your team spends less time digging through stack traces.

```
Your App  →  SDK  →  HTTP Ingest  →  Kafka  →  AI Worker  →  PostgreSQL
                                                                    ↓
                                              Dashboard / Alerts / API
```

---

## Features

- **Language SDKs** — Node.js (available), Python and Java (coming soon)
- **AI-powered analysis** — plug in any provider: Claude, OpenAI, Gemini, or a self-hosted Ollama model
- **Pluggable notifications** — Google Chat cards, email, and generic webhooks
- **Real-time dashboard** — TanStack Start (Phase 2)
- **Full OpenAPI spec** — browse and test the API in the built-in Swagger UI
- **Privacy-first** — SDKs strip secrets, tokens, and passwords locally before any data leaves your server

---

## Tech Stack

| Layer | Technology |
|---|---|
| Server | Go 1.23, [Chi](https://github.com/go-chi/chi) |
| Queue | [Apache Kafka](https://kafka.apache.org/) (`segmentio/kafka-go`) |
| Storage | PostgreSQL, [sqlc](https://sqlc.dev/) |
| AI providers | Claude · OpenAI · Gemini · Ollama |
| Dashboard | TanStack Start (Phase 2) |
| SDK | TypeScript / Node.js |

---

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/)
- (For local Go development) Go 1.23+
- (For SDK development) Node.js 18+

---

## Quick Start

### 1. Clone the repository

```bash
git clone git@github.com:Godie360/custos.git
cd custos
```

### 2. Configure environment

```bash
cp .env.example .env
```

Open `.env` and set the required values:

```env
# Required — generate a random 32+ char string
JWT_SECRET=your-secret-here

# Optional — enable AI analysis (choose one provider)
CUSTOS_AI_PROVIDER=claude          # claude | openai | gemini | ollama
CUSTOS_AI_API_KEY=sk-...
CUSTOS_AI_MODEL=claude-opus-4-8    # or gpt-4o, gemini-1.5-pro, etc.

# Optional — notifications
GOOGLE_CHAT_WEBHOOK_URL=https://chat.googleapis.com/v1/spaces/...
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=alerts@example.com
SMTP_PASS=your-smtp-password
NOTIFICATION_WEBHOOK_URL=https://your-webhook-endpoint.com
```

> The server starts without AI analysis if `CUSTOS_AI_PROVIDER` is not set — errors are still captured and stored.

### 3. Start the stack

```bash
docker compose up --build
```

| Service | URL |
|---|---|
| Custos API | http://localhost:8080 |
| Swagger UI | http://localhost:8081 |

### 4. Create your first project

```bash
curl -s -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{"name": "my-app", "description": "My first Custos project"}' | jq .
```

Copy the `api_key` from the response — you will pass it as `X-Custos-Key` in SDK configuration and API requests.

---

## SDK — Node.js

### Install

```bash
npm install @custos/sdk
```

### NestJS — global exception filter

```ts
// main.ts
import { NestFactory } from '@nestjs/core';
import { CustosClient, CustosExceptionFilter } from '@custos/sdk';
import { AppModule } from './app.module';

async function bootstrap() {
  const app = await NestFactory.create(AppModule);

  const custos = new CustosClient({
    apiKey: process.env.CUSTOS_API_KEY!,
    host: process.env.CUSTOS_HOST!,   // http://localhost:8080
    service: 'my-nestjs-app',
    environment: process.env.NODE_ENV ?? 'production',
  });

  app.useGlobalFilters(new CustosExceptionFilter(custos));
  await app.listen(3000);
}
bootstrap();
```

### Winston transport

```ts
import winston from 'winston';
import { CustosClient, CustosWinstonTransport } from '@custos/sdk';

const custos = new CustosClient({
  apiKey: process.env.CUSTOS_API_KEY!,
  host: process.env.CUSTOS_HOST!,
  service: 'my-app',
});

const logger = winston.createLogger({
  transports: [
    new winston.transports.Console(),
    new CustosWinstonTransport({ client: custos }),
  ],
});
```

### Graceful shutdown

```ts
process.on('SIGTERM', () => {
  custos.shutdown(); // flushes buffered events before exit
  process.exit(0);
});
```

See [`sdks/nodejs/README.md`](sdks/nodejs/README.md) for the full SDK reference.

---

## API Reference

The full OpenAPI 3.0 specification is at [`api/openapi.yaml`](api/openapi.yaml).

When the stack is running, open **http://localhost:8081** to browse and try every endpoint interactively.

### Key endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Health check |
| `POST` | `/api/v1/ingest` | Ingest an error event (requires `X-Custos-Key`) |
| `GET` | `/api/v1/issues` | List issues (filterable by service, severity, status) |
| `GET` | `/api/v1/issues/:id` | Get a single issue with AI analysis |
| `PATCH` | `/api/v1/issues/:id` | Update issue status |
| `GET` | `/api/v1/analytics/summary` | Severity counts and top services |
| `GET` | `/api/v1/projects` | List projects |
| `POST` | `/api/v1/projects` | Create a project |
| `POST` | `/api/v1/projects/:id/keys` | Generate an API key |

---

## Architecture

```
┌─────────────┐     HTTP POST      ┌──────────────────┐
│  Language   │  ─────────────►   │   Go + Chi API   │
│    SDK      │  X-Custos-Key      │   (port 8080)    │
└─────────────┘                   └────────┬─────────┘
                                           │ Kafka publish
                                           ▼
                                  ┌──────────────────┐
                                  │   Kafka Topic    │
                                  │  custos.events   │
                                  └────────┬─────────┘
                                           │ consume
                                           ▼
                                  ┌──────────────────┐     ┌───────────────┐
                                  │  Analysis Worker │────►│  AI Provider  │
                                  │  (same process)  │     │  (pluggable)  │
                                  └────────┬─────────┘     └───────────────┘
                                           │ store
                                           ▼
                                  ┌──────────────────┐     ┌───────────────┐
                                  │   PostgreSQL     │     │ Notifications │
                                  │  (issues, events)│     │ (Chat/email/  │
                                  └──────────────────┘     │  webhook)     │
                                                           └───────────────┘
```

### Package dependency rule

```
domain ← store ← service ← api
domain ← queue ← service
domain ← provider ← service
domain ← notification ← service
```

`internal/domain` has zero imports from other internal packages.

---

## Development

### Run with hot reload

```bash
make dev        # uses Air for Go hot reload
```

### Run tests

```bash
# Go
go test ./...

# Node.js SDK
cd sdks/nodejs && npx jest
```

### Lint

```bash
golangci-lint run ./...
```

### Database migrations

```bash
make migrate-up    # apply all pending migrations
make migrate-down  # roll back the last migration
```

### Build the server binary

```bash
make build        # outputs to bin/server
```

### Regenerate sqlc queries

After editing SQL files in `internal/store/postgres/queries/`:

```bash
sqlc generate
```

---

## Project Structure

```
custos/
├── cmd/server/          # Application entry point
├── internal/
│   ├── api/             # HTTP handlers, middleware, router
│   ├── config/          # Environment-based configuration
│   ├── domain/          # Core types — no internal imports
│   ├── notification/    # Google Chat, email, webhook
│   ├── provider/        # AI adapter registry + adapters
│   ├── queue/           # Kafka producer/consumer
│   ├── service/         # Business logic (ingestion, analysis, notification)
│   └── store/           # PostgreSQL store + sqlc generated code
├── migrations/          # SQL migration files
├── api/                 # OpenAPI specification
├── sdks/
│   └── nodejs/          # @custos/sdk — Node.js / NestJS SDK
├── docs/                # Design and implementation docs
├── docker/              # Dockerfiles
├── docker-compose.yml   # Production-like local stack
└── docker-compose.dev.yml  # Development stack with hot reload
```

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full guide.

Quick summary:
1. Fork the repo and create a branch from `develop` — `feat/your-feature` or `fix/your-fix`
2. Write tests for any new behaviour
3. Open a PR targeting `develop` — fill in the template
4. PRs are merged to `develop`; `develop` → `main` cuts a release

---

## License

[MIT](LICENSE)
