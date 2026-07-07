import { redact, redactStack } from './redact';
import { CustosClient } from './client';

/**
 * NestJS global exception filter for Custos.
 *
 * Usage in main.ts:
 *   import { CustosExceptionFilter } from '@custos/sdk';
 *
 *   const app = await NestFactory.create(AppModule);
 *   app.useGlobalFilters(new CustosExceptionFilter(custosClient));
 *   await app.listen(3000);
 */
export class CustosExceptionFilter {
  private readonly client: CustosClient;

  constructor(client: CustosClient) {
    this.client = client;
  }

  catch(exception: unknown, host: unknown): void {
    try {
      const err = exception instanceof Error ? exception : new Error(String(exception));
      const frames = parseStack(err.stack ?? '');

      // Extract HTTP context if available (NestJS HttpArgumentsHost)
      let context = '';
      try {
        const ctx = (host as { switchToHttp: () => { getRequest: () => { url: string; method: string } } })
          .switchToHttp()
          .getRequest();
        context = `${ctx.method} ${ctx.url}`;
      } catch {
        // Not an HTTP context — ignore
      }

      const message = context ? `${redact(err.message)} [${context}]` : redact(err.message);

      this.client.enqueue({
        service: this.client['config'].service,
        environment: this.client['config'].environment,
        error_type: err.name,
        message,
        stack_trace: redactStack(frames),
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
