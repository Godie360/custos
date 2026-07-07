import { redact, redactStack } from './redact';
import { CustosClient } from './client';

/**
 * Winston transport for Custos.
 * Usage:
 *   import winston from 'winston';
 *   import { CustosWinstonTransport } from '@custos/sdk';
 *
 *   const logger = winston.createLogger({
 *     transports: [new CustosWinstonTransport({ client: custosClient })],
 *   });
 */
export class CustosWinstonTransport {
  private readonly client: CustosClient;
  readonly name = 'custos';
  readonly level: string;

  constructor(opts: { client: CustosClient; level?: string }) {
    this.client = opts.client;
    this.level = opts.level ?? 'error';
  }

  log(info: Record<string, unknown>, callback: () => void): void {
    try {
      const err = info instanceof Error ? info : (info['error'] as Error | undefined);
      const message = String(info['message'] ?? info['msg'] ?? '');
      const frames = parseStack(err?.stack ?? '');

      this.client.enqueue({
        service: this.client['config'].service,
        environment: this.client['config'].environment,
        error_type: err?.name ?? 'Error',
        message: redact(message),
        stack_trace: redactStack(frames),
      });
    } catch {
      // Never throw into the host application
    } finally {
      callback();
    }
  }
}

/**
 * Pino transport for Custos.
 * Usage with pino-multi-stream or as a custom stream:
 *   import pino from 'pino';
 *   import { CustosPinoStream } from '@custos/sdk';
 *
 *   const stream = new CustosPinoStream({ client: custosClient });
 *   const logger = pino({ level: 'error' }, stream);
 */
export class CustosPinoStream {
  private readonly client: CustosClient;

  constructor(opts: { client: CustosClient }) {
    this.client = opts.client;
  }

  write(chunk: string): void {
    try {
      const log = JSON.parse(chunk) as Record<string, unknown>;
      // Pino level 50 = error, 60 = fatal
      const level = Number(log['level'] ?? 0);
      if (level < 50) return;

      const message = String(log['msg'] ?? log['message'] ?? '');
      const stack = String(log['stack'] ?? log['err']?.['stack'] ?? '');
      const errorType = String(log['type'] ?? log['err']?.['type'] ?? 'Error');

      this.client.enqueue({
        service: this.client['config'].service,
        environment: this.client['config'].environment,
        error_type: redact(errorType),
        message: redact(message),
        stack_trace: redactStack(parseStack(stack)),
      });
    } catch {
      // Never throw into the host application
    }
  }
}

function parseStack(stack: string): string[] {
  if (!stack) return [];
  return stack
    .split('\n')
    .filter(line => line.trim().startsWith('at '))
    .slice(0, 10);
}
