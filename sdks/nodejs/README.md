# @custos/sdk

Node.js and NestJS SDK for [Custos](https://github.com/iPFSoftwares/custos) — automatic error capture with AI-powered explanations.

## Install

```bash
npm install @custos/sdk
```

## Quick start

```ts
import { CustosClient } from '@custos/sdk';

const custos = new CustosClient({
  apiKey: process.env.CUSTOS_API_KEY!,
  host: process.env.CUSTOS_HOST!,   // e.g. http://localhost:8080
  service: 'payments-api',
  environment: process.env.NODE_ENV ?? 'production',
});
```

## NestJS — global exception filter

Add to `main.ts`:

```ts
import { NestFactory } from '@nestjs/core';
import { CustosClient, CustosExceptionFilter } from '@custos/sdk';
import { AppModule } from './app.module';

async function bootstrap() {
  const app = await NestFactory.create(AppModule);

  const custos = new CustosClient({
    apiKey: process.env.CUSTOS_API_KEY!,
    host: process.env.CUSTOS_HOST!,
    service: 'my-nestjs-app',
  });

  app.useGlobalFilters(new CustosExceptionFilter(custos));
  await app.listen(3000);
}
bootstrap();
```

## Winston transport

```ts
import winston from 'winston';
import { CustosClient, CustosWinstonTransport } from '@custos/sdk';

const custos = new CustosClient({ apiKey: '...', host: '...', service: 'my-app' });

const logger = winston.createLogger({
  transports: [
    new winston.transports.Console(),
    new CustosWinstonTransport({ client: custos }),
  ],
});

logger.error('Something went wrong', { error: new Error('oops') });
```

## Pino stream

```ts
import pino from 'pino';
import { CustosClient, CustosPinoStream } from '@custos/sdk';

const custos = new CustosClient({ apiKey: '...', host: '...', service: 'my-app' });

const logger = pino({ level: 'error' }, new CustosPinoStream({ client: custos }));
```

## Configuration

| Option | Type | Default | Description |
|---|---|---|---|
| `apiKey` | `string` | required | Custos project API key |
| `host` | `string` | required | Custos server URL |
| `service` | `string` | required | Name of this service |
| `environment` | `string` | `production` | Deployment environment |
| `batchSize` | `number` | `10` | Events to buffer before flushing |
| `flushIntervalMs` | `number` | `5000` | Auto-flush interval in ms |
| `maxRetries` | `number` | `3` | Retry attempts on network failure |

## Graceful shutdown

```ts
process.on('SIGTERM', () => {
  custos.shutdown(); // flushes any buffered events
  process.exit(0);
});
```

## Privacy

All events are redacted locally before leaving your server. The SDK strips bearer tokens, passwords, connection strings, JWT tokens, and AWS keys before sending anything to Custos.

## License

MIT
